package zoox

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/template"

	"golang.org/x/sync/errgroup"

	"github.com/go-errors/errors"
	"github.com/go-zoox/cache"
	"github.com/go-zoox/chalk"
	"github.com/go-zoox/core-utils/cast"
	"github.com/go-zoox/core-utils/regexp"
	"github.com/go-zoox/i18n"
	jsonrpcServer "github.com/go-zoox/jsonrpc/server"
	"github.com/go-zoox/kv"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/session"
	"github.com/go-zoox/websocket"
	"github.com/go-zoox/zoox/components/application/cmd"
	"github.com/go-zoox/zoox/components/application/cron"
	"github.com/go-zoox/zoox/components/application/debug"
	"github.com/go-zoox/zoox/components/application/env"
	"github.com/go-zoox/zoox/components/application/jobqueue"
	"github.com/go-zoox/zoox/components/application/runtime"

	"github.com/go-zoox/mq"
	"github.com/go-zoox/pubsub"

	"github.com/go-zoox/kv/redis"
)

// HandlerFunc defines the request handler used by zoox
type HandlerFunc func(ctx *Context)

// GroupFunc defines the group handler used by zoox
type GroupFunc func(group *RouterGroup)

// Middleware defines the signature of the middleware function.
type Middleware = HandlerFunc

// WsHandlerFunc defines the websocket handler used by zoox
type WsHandlerFunc func(ctx *Context, conn websocket.Server)

// JSONRPCHandlerFunc defines the jsonrpc handler used by zoox
type JSONRPCHandlerFunc func(registry jsonrpcServer.Server)

// Application is the handler for all requests.
type Application struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
	// templates
	templates     *template.Template
	templateFuncs template.FuncMap
	//
	notfound HandlerFunc
	//
	cache cache.Cache
	//
	cron  cron.Cron
	queue jobqueue.JobQueue
	//
	cmd cmd.Cmd
	// i18n
	i18n i18n.I18n
	//
	Env    env.Env
	Logger *logger.Logger
	// Debug
	debug debug.Debug
	// Runtime
	runtime runtime.Runtime

	//
	jsonrpcRegistry jsonrpcServer.Server
	//
	pubsub pubsub.PubSub
	mq     mq.MQ

	//
	Config ApplicationConfig

	// once
	once struct {
		//
		debug   sync.Once
		runtime sync.Once
		//
		cache sync.Once
		cron  sync.Once
		queue sync.Once
		//
		i18n sync.Once
		//
		jsonrpcRegistry sync.Once
		//
		pubsub sync.Once
		mq     sync.Once
		//
		cmd sync.Once
	}

	// tls cert loader
	tlsCertLoader func(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error)
}

// ApplicationConfig defines the config of zoox.Application.
type ApplicationConfig struct {
	Protocol  string
	Host      string
	Port      int
	HTTPSPort int

	//
	NetworkType      string
	UnixDomainSocket string

	// TLS
	// TLS Certificate
	TLSCertFile string
	// TLS Private Key
	TLSKeyFile string
	// TLS Ca Certificate
	TLSCaCertFile string
	//
	TLSCert []byte
	TLSKey  []byte

	//
	LogLevel string `config:"log_level"`
	//
	SecretKey string `config:"secret_key"`
	//
	Session session.Config `config:"session"`
	//
	Cache cache.Config `config:"cache"`
	//
	Redis ApplicationConfigRedis `config:"redis"`

	//
	Banner string
}

// ApplicationConfigRedis defines the config of redis.
type ApplicationConfigRedis struct {
	Host     string `config:"host"`
	Port     int    `config:"port"`
	DB       int    `config:"db"`
	Username string `config:"username"`
	Password string `config:"password"`
}

// New is the constructor of zoox.Application.
func New() *Application {
	app := &Application{
		router:        newRouter(),
		templateFuncs: template.FuncMap{},
		notfound:      NotFound(),
	}

	app.RouterGroup = newRouterGroup(app, "")
	app.groups = []*RouterGroup{app.RouterGroup}

	app.Env = env.New()

	app.Logger = logger.New(&logger.Options{
		Level: app.Config.LogLevel,
	})

	// global middlewares
	for _, mf := range DefaultMiddlewares {
		// fmt.Println("zoox: default middleware:", name)
		mf(app)
	}

	// global groups
	for gprefix, gfn := range DefaultGroupsFns {
		// fmt.Println("zoox: default group:", gprefix)
		g := app.Group(gprefix)
		gfn(g)
	}

	return app
}

// NotFound defines the 404 handler, replaced of built in not found handler.
func (app *Application) NotFound(h HandlerFunc) {
	app.notfound = h
}

// Fallback is the default handler for all requests.
func (app *Application) Fallback(h HandlerFunc) {
	app.NotFound(h)
}

// defaultConfig
func (app *Application) applyDefaultConfig() error {
	if err := app.applyDefaultConfigFromEnv(); err != nil {
		return err
	}

	if app.Config.SecretKey == "" {
		app.Config.SecretKey = DefaultSecretKey
	}

	if app.Config.Protocol == "" {
		app.Config.Protocol = "http"
	}

	if app.Config.Host == "" {
		app.Config.Host = "0.0.0.0"
	}

	if app.Config.Port == 0 {
		app.Config.Port = 8080
	}

	if app.Config.NetworkType == "" {
		app.Config.NetworkType = "tcp"
	}

	if app.Config.Session.MaxAge == 0 {
		app.Config.Session.MaxAge = DefaultSessionMaxAge
	}

	if app.Config.Cache.Config == nil {
		if app.Config.Redis.Host != "" {
			app.Config.Cache = kv.Config{
				Engine: "redis",
				Config: &redis.Config{
					Host:     app.Config.Redis.Host,
					Port:     app.Config.Redis.Port,
					Password: app.Config.Redis.Password,
					DB:       app.Config.Redis.DB,
					Prefix:   "eunomia",
				},
			}
		}
	}

	return nil
}

func (app *Application) applyDefaultConfigFromEnv() error {
	if app.Config.Port == 0 && os.Getenv(BuiltInEnvPort) != "" {
		app.Config.Port = cast.ToInt(os.Getenv(BuiltInEnvPort))
	}

	if app.Config.HTTPSPort == 0 && os.Getenv(BuiltInEnvHTTPSPort) != "" {
		app.Config.HTTPSPort = cast.ToInt(os.Getenv(BuiltInEnvHTTPSPort))
	}

	if app.Config.LogLevel == "" && os.Getenv(BuiltInEnvLogLevel) != "" {
		app.Config.LogLevel = os.Getenv(BuiltInEnvLogLevel)
	}

	if app.Config.SecretKey == "" && os.Getenv(BuiltInEnvSecretKey) != "" {
		app.Config.SecretKey = os.Getenv(BuiltInEnvSecretKey)
	}

	if app.Config.Session.MaxAge == 0 && os.Getenv(BuiltInEnvSessionMaxAge) != "" {
		app.Config.Session.MaxAge = cast.ToDuration(os.Getenv(BuiltInEnvSessionMaxAge))
	}

	if app.Config.Redis.Host == "" && os.Getenv(BuiltInEnvRedisHost) != "" {
		app.Config.Redis.Host = os.Getenv(BuiltInEnvRedisHost)
	}

	if app.Config.Redis.Port == 0 && os.Getenv(BuiltInEnvRedisPort) != "" {
		app.Config.Redis.Port = cast.ToInt(os.Getenv(BuiltInEnvRedisPort))
	}

	if app.Config.Redis.Username == "" && os.Getenv(BuiltInEnvRedisUser) != "" {
		app.Config.Redis.Username = os.Getenv(BuiltInEnvRedisUser)
	}

	if app.Config.Redis.Password == "" && os.Getenv(BuiltInEnvRedisPass) != "" {
		app.Config.Redis.Password = os.Getenv(BuiltInEnvRedisPass)
	}

	if app.Config.Redis.DB == 0 && os.Getenv(BuiltInEnvRedisDB) != "" {
		app.Config.Redis.DB = cast.ToInt(os.Getenv(BuiltInEnvRedisDB))
	}

	return nil
}

// Run defines the method to start the server
// Example:
//
//			IP:
//		   default(http://0.0.0.0:8080): Run(":8080")
//			 port(http://0.0.0.0:8888): Run(":8888")
//			 host+port(http://127.0.0.1:8888): Run("127.0.0.1:8888")
//
//	   HTTP:
//			 scheme://host+port(http://127.0.0.1:8888): Run("http://127.0.0.1:8888")
//
//			Unix Domain Socket:
//				/tmp/xxx.sock: Run("unix:///tmp/xxx.sock")
func (app *Application) Run(addr ...string) (err error) {
	// show banner
	app.showBanner()

	// parse addr
	if err := app.parseAddr(addr...); err != nil {
		return err
	}

	// apply default config
	if err := app.applyDefaultConfig(); err != nil {
		return fmt.Errorf("failed to apply default config: %v", err)
	}

	// show app info in debug mode
	app.showAppInfo()

	// show runtime info
	app.showRuntimeInfo()

	// serve
	return app.serve()
}

func (app *Application) createContext(w http.ResponseWriter, req *http.Request) *Context {
	return newContext(app, w, req)
}

// SetTemplates set the template
func (app *Application) SetTemplates(dir string, fns ...template.FuncMap) {
	if len(fns) > 0 && fns[0] != nil {
		app.templateFuncs = fns[0]
	}

	app.templates = template.Must(template.New("").Funcs(app.templateFuncs).ParseGlob(dir + "/*"))
}

// SetBanner sets the banner
func (app *Application) SetBanner(banner string) {
	app.Config.Banner = banner
}

func (app *Application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := app.createContext(w, req)

	var middlewares []HandlerFunc

	for _, group := range app.groups {
		if strings.HasPrefix(ctx.Path, group.prefix) {
			// @TODO /v1 => /v1/
			// if ctx.Path == group.prefix && !strings.HasSuffix(group.prefix, "/") {
			// 	ctx.Path += "/"
			// }

			middlewares = append(middlewares, group.middlewares...)
		}
	}

	ctx.handlers = middlewares
	app.router.handle(ctx)
}

// SetTLSCertLoader set the tls cert loader
func (app *Application) SetTLSCertLoader(loader func(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error)) {
	app.tlsCertLoader = loader
}

// IsProd returns true if the app is in production mode.
func (app *Application) IsProd() bool {
	return app.Env.Get("MODE") == "production"
}

// JSONRPCRegistry get a new JSONRPCRegistry handler.
func (app *Application) JSONRPCRegistry() jsonrpcServer.Server {
	app.once.jsonrpcRegistry.Do(func() {
		app.jsonrpcRegistry = jsonrpcServer.New()
	})

	return app.jsonrpcRegistry
}

// PubSub get a new PubSub handler.
func (app *Application) PubSub() pubsub.PubSub {
	if app.Config.Redis.Host == "" {
		panic("redis config is required for pubsub in application")
	}

	app.once.pubsub.Do(func() {
		app.pubsub = pubsub.New(&pubsub.Config{
			RedisHost:     app.Config.Redis.Host,
			RedisPort:     app.Config.Redis.Port,
			RedisDB:       app.Config.Redis.DB,
			RedisUsername: app.Config.Redis.Username,
			RedisPassword: app.Config.Redis.Password,
		})
	})

	return app.pubsub
}

// MQ get a new MQ handler.
func (app *Application) MQ() mq.MQ {
	if app.Config.Redis.Host == "" {
		panic("redis config is required for mq in application")
	}

	app.once.mq.Do(func() {
		app.mq = mq.New(&mq.Config{
			RedisHost:     app.Config.Redis.Host,
			RedisPort:     app.Config.Redis.Port,
			RedisDB:       app.Config.Redis.DB,
			RedisUsername: app.Config.Redis.Username,
			RedisPassword: app.Config.Redis.Password,
		})
	})

	return app.mq
}

// Cache ...
func (app *Application) Cache() cache.Cache {
	app.once.cache.Do(func() {
		app.cache = cache.New(&app.Config.Cache)
	})

	return app.cache
}

// Cron ...
func (app *Application) Cron() cron.Cron {
	app.once.cron.Do(func() {
		app.cron = cron.New()
	})

	return app.cron
}

// JobQueue ...
func (app *Application) JobQueue() jobqueue.JobQueue {
	app.once.queue.Do(func() {
		app.queue = jobqueue.New()
	})

	return app.queue
}

// Cmd ...
func (app *Application) Cmd() cmd.Cmd {
	app.once.cmd.Do(func() {
		app.cmd = cmd.New(context.Background())
	})

	return app.cmd
}

// I18n ...
func (app *Application) I18n() i18n.I18n {
	app.once.queue.Do(func() {
		app.i18n = i18n.New()
	})

	return app.i18n
}

// Debug ...
func (app *Application) Debug() debug.Debug {
	app.once.debug.Do(func() {
		app.debug = debug.New(app.Logger)
	})

	return app.debug
}

// Runtime ...
func (app *Application) Runtime() runtime.Runtime {
	app.once.runtime.Do(func() {
		app.runtime = runtime.New(app.Logger)
	})

	return app.runtime
}

// Address ...
func (app *Application) Address() string {
	if app.Config.NetworkType == "unix" {
		return app.Config.UnixDomainSocket
	}

	return fmt.Sprintf("%s:%d", app.Config.Host, app.Config.Port)
}

// AddressHTTPS ...
func (app *Application) AddressHTTPS() string {
	if app.Config.NetworkType == "unix" {
		return app.Config.UnixDomainSocket
	}

	return fmt.Sprintf("%s:%d", app.Config.Host, app.Config.HTTPSPort)
}

// AddressForLog ...
func (app *Application) AddressForLog() string {
	if app.Config.NetworkType == "unix" {
		return app.Config.UnixDomainSocket
	}

	if app.Config.Host == "0.0.0.0" {
		return fmt.Sprintf("127.0.0.1:%d", app.Config.Port)
	}

	return fmt.Sprintf("%s:%d", app.Config.Host, app.Config.Port)
}

// AddressHTTPSForLog ...
func (app *Application) AddressHTTPSForLog() string {
	if app.Config.NetworkType == "unix" {
		return app.Config.UnixDomainSocket
	}

	if app.Config.Host == "0.0.0.0" {
		return fmt.Sprintf("127.0.0.1:%d", app.Config.HTTPSPort)
	}

	return fmt.Sprintf("%s:%d", app.Config.Host, app.Config.HTTPSPort)
}

// showBanner ...
func (app *Application) showBanner() {
	// allow custom banner
	if app.Config.Banner != "" {
		log.Println(app.Config.Banner)
	}

	// banner
	log.Printf(banner, chalk.Green("v"+Version), chalk.Blue(website))
}

// parseAddr ...
func (app *Application) parseAddr(addr ...string) error {
	var addrX string
	if len(addr) > 0 && addr[0] != "" {
		addrX = addr[0]
	} else {
		if os.Getenv("PORT") != "" {
			addrX = ":" + os.Getenv("PORT")
		}
	}

	if addrX != "" {
		// Pattern@1 => :8080
		if regexp.Match("^:\\d+$", addrX) {
			app.Config.Port = cast.ToInt(addrX[1:])
		} else if regexp.Match("^[\\w\\.]+:\\d+$", addrX) {
			// Pattern@2 => 127.0.0.1:8080
			parts := strings.Split(addrX, ":")
			app.Config.Host = cast.ToString(parts[0])
			app.Config.Port = cast.ToInt(parts[1])
		} else if regexp.Match("^http://[\\w\\.]+:\\d+", addrX) {
			// Pattern@3 => http://127.0.0.1:8080
			u, err := url.Parse(addrX)
			if err != nil {
				return fmt.Errorf("failed to parse addr(%s): %v", addrX, err)
			}

			app.Config.Protocol = u.Scheme
			app.Config.Host = u.Hostname()
			app.Config.Port = cast.ToInt(u.Port())
		} else if regexp.Match("^unix://", addrX) {
			// Pattern@4 => unix:///tmp/xxx.sock
			app.Config.Protocol = "unix"
			app.Config.NetworkType = "unix"
			app.Config.UnixDomainSocket = addrX[7:]
		} else if regexp.Match("^/", addrX) {
			// Pattern@4 => /tmp/xxx.sock
			app.Config.Protocol = "unix"
			app.Config.NetworkType = "unix"
			app.Config.UnixDomainSocket = addrX
		}
	}

	return nil
}

// showAppInfo ...
func (app *Application) showAppInfo() {
	app.Debug().Info(app)
}

// showRuntimeInfo ...
func (app *Application) showRuntimeInfo() {
	app.Runtime().Print()
}

// serve ...
func (app *Application) serve() error {
	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return app.serveHTTP(ctx)
	})

	g.Go(func() error {
		return app.serveHTTPS(ctx)
	})

	return g.Wait()
}

// serveHTTP ...
func (app *Application) serveHTTP(ctx context.Context) error {
	listener, err := net.Listen(app.Config.NetworkType, app.Address())
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{
		Addr:    app.Address(),
		Handler: app,
	}

	go func() {
		<-ctx.Done() // 当上下文被取消时，停止服务器
		server.Close()
	}()

	if app.Config.NetworkType == "unix" {
		logger.Info("Server started at unix://%s", app.AddressForLog())
	} else {
		logger.Info("Server started at http://%s", app.AddressForLog())
	}

	// 等待所有 goroutine 完成
	return server.Serve(listener)
}

// serveHTTPS ...
func (app *Application) serveHTTPS(ctx context.Context) error {
	// if HTTPSPort is not set, ignore set https
	if app.Config.HTTPSPort == 0 {
		return nil
	}

	listener, err := net.Listen(app.Config.NetworkType, app.AddressHTTPS())
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{
		Addr:    app.AddressHTTPS(),
		Handler: app,
	}

	go func() {
		<-ctx.Done() // 当上下文被取消时，停止服务器
		server.Close()
	}()

	var config *tls.Config

	// TLS Ca Certificate
	if app.Config.TLSCaCertFile != "" {
		pool := x509.NewCertPool()
		caCrt, err := ioutil.ReadFile(app.Config.TLSCaCertFile)
		if err != nil {
			return fmt.Errorf("failed to read tls ca certificate")
		}
		pool.AppendCertsFromPEM(caCrt)

		if config == nil {
			config = &tls.Config{}
		}

		config.ClientCAs = pool
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	// @1 load tls from file: TLS Certificate and Private Key
	if app.Config.TLSCertFile != "" && app.Config.TLSKeyFile != "" {
		certPEMBlock, err := os.ReadFile(app.Config.TLSCertFile)
		if err != nil {
			return fmt.Errorf("failed to read tls certificate: %v", err)
		}
		app.Config.TLSCert = certPEMBlock

		keyPEMBlock, err := os.ReadFile(app.Config.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read tls private key: %v", err)
		}
		app.Config.TLSKey = keyPEMBlock

		// // if app.Config.TLSCertFile != "" && app.TLSCert == nil {
		// // 	tlsCaCert, err := ioutil.ReadFile(app.Config.TLSCertFile)
		// // 	if err != nil {
		// // 		return err
		// // 	}
		// // 	app.TLSCert = tlsCaCert
		// // }

		// if app.Config.NetworkType == "unix" {
		// 	logger.Info("Server started at unixs://%s", app.AddressForLog())
		// } else {
		// 	logger.Info("Server started at https://%s", app.AddressForLog())
		// }

		// // if err := http.ServeTLS(listener, app, app.Config.TLSCertFile, app.Config.TLSKeyFile); err != nil {
		// // 	return err
		// // }

		// return server.ServeTLS(listener, app.Config.TLSCertFile, app.Config.TLSKeyFile)
	}

	// @2 load tls from memory: TLS Certificate and Private Key
	if app.Config.TLSCert != nil && app.Config.TLSKey != nil {
		cert, err := tls.X509KeyPair(app.Config.TLSCert, app.Config.TLSKey)
		if err != nil {
			return err
		}

		if config == nil {
			config = &tls.Config{}
		}

		config.Certificates = []tls.Certificate{cert}
	}

	// @3 load tls by sni
	// reference:
	//	 - https://medium.com/@satrobit/how-to-build-https-servers-with-certificate-lazy-loading-in-go-bff5e9ef2f1f
	//
	if app.tlsCertLoader != nil {
		if config == nil {
			config = &tls.Config{}
		}

		config.GetCertificate = app.tlsCertLoader
	}

	if config == nil {
		return errors.New("failed to start https server, tls config is required; you can set tls cert and key by app.Config.TLSCertFile and app.Config.TLSKeyFile, or app.Config.TLSCert and app.Config.TLSKey, or app.SetTLSCertLoader method")
	}

	if app.Config.NetworkType == "unix" {
		logger.Info("Server started at unix://%s", app.AddressHTTPSForLog())
	} else {
		logger.Info("Server started at https://%s", app.AddressHTTPSForLog())
	}

	return server.Serve(tls.NewListener(listener, config))
}

// H is a shortcut for map[string]interface{}
type H map[string]interface{}
