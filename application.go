package zoox

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/go-zoox/core-utils/cast"
	"github.com/go-zoox/core-utils/regexp"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox/components/context/cache"
	"github.com/go-zoox/zoox/components/context/cron"
	"github.com/go-zoox/zoox/components/context/debug"
	"github.com/go-zoox/zoox/components/context/env"
	"github.com/go-zoox/zoox/components/context/queue"
	"github.com/go-zoox/zoox/components/context/websocket"
	"github.com/go-zoox/zoox/rpc/jsonrpc"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(ctx *Context)

// Middleware defines the signature of the middleware function.
type Middleware = HandlerFunc

// WsHandlerFunc defines the websocket handler used by gee
type WsHandlerFunc func(ctx *Context, client *websocket.WebSocketClient)

// WsGorillaHandlerFunc defines the websocket handler used by gee
type WsGorillaHandlerFunc func(ctx *Context, client *websocket.Conn)

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
	SecretKey string
	//
	SessionMaxAge time.Duration
	LogLevel      string
	//
	CacheConfig *cache.Config
	cache       cache.Cache
	//
	cron  cron.Cron
	queue queue.Queue
	//
	Env    env.Env
	Logger *logger.Logger
	// Debug
	debug debug.Debug

	// TLS Certificate
	TLSCertFile string
	// TLS Private Key
	TLSKeyFile string
	// TLS Ca Certificate
	TLSCaCertFile string

	//
	Protocol string
	Host     string
	Port     int
	//
	NetworkType      string
	UnixDomainSocket string
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
		Level: app.LogLevel,
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
	if app.SecretKey == "" {
		app.SecretKey = DefaultSecretKey
	}

	if app.Protocol == "" {
		app.Protocol = "http"
	}

	if app.Host == "" {
		app.Host = "0.0.0.0"
	}

	if app.Port == 0 {
		app.Port = 8080
	}

	if app.NetworkType == "" {
		app.NetworkType = "tcp"
	}

	if app.SessionMaxAge == 0 {
		app.SessionMaxAge = DefaultSessionMaxAge
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
	var addrX string
	if len(addr) > 0 && addr[0] != "" {
		addrX = addr[0]
	}

	// Pattern@1 => :8080
	if regexp.Match(addrX, ":\\d+") {
		app.Port = cast.ToInt(addrX[1:])
	} else if regexp.Match(addrX, "\\s+:\\d+") {
		// Pattern@2 => 127.0.0.1:8080
		parts := strings.Split(addrX, ":")
		app.Host = cast.ToString(parts[0])
		app.Port = cast.ToInt(parts[1])
	} else if regexp.Match(addrX, "^http://\\s+:\\d+") {
		// Pattern@3 => http://127.0.0.1:8080
		u, err := url.Parse(addrX)
		if err != nil {
			return fmt.Errorf("failed to parse addr(%s): %v", addrX, err)
		}

		app.Protocol = u.Scheme
		parts := strings.Split(u.Host, ":")
		app.Host = cast.ToString(parts[0])
		app.Port = cast.ToInt(parts[1])
	} else if regexp.Match(addrX, "^unix://") {
		// Pattern@4 => unix:///tmp/xxx.sock
		app.Protocol = "unix"
		app.NetworkType = "unix"
		app.UnixDomainSocket = addrX[7:]
	} else if regexp.Match(addrX, "^/") {
		// Pattern@4 => /tmp/xxx.sock
		app.Protocol = "unix"
		app.NetworkType = "unix"
		app.UnixDomainSocket = addrX
	}

	// config
	if err := app.applyDefaultConfig(); err != nil {
		return fmt.Errorf("failed to apply default config: %v", err)
	}

	listener, err := net.Listen(app.NetworkType, app.Address())
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{
		Addr:    app.Address(),
		Handler: app,
	}

	// TLS Ca Certificate
	if app.TLSCaCertFile != "" {
		pool := x509.NewCertPool()
		caCrt, err := ioutil.ReadFile(app.TLSCaCertFile)
		if err != nil {
			return fmt.Errorf("failed to read tls ca certificate")
		}
		pool.AppendCertsFromPEM(caCrt)

		if server.TLSConfig == nil {
			server.TLSConfig = &tls.Config{
				ClientCAs:  pool,
				ClientAuth: tls.RequireAndVerifyClientCert,
			}
		}
	}

	// TLS Certificate and Private Key
	if app.TLSCertFile != "" {
		// if app.TLSCertFile != "" && app.TLSCert == nil {
		// 	tlsCaCert, err := ioutil.ReadFile(app.TLSCertFile)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	app.TLSCert = tlsCaCert
		// }

		if app.NetworkType == "unix" {
			logger.Info("Server started at unixs://%s", app.Address())
		} else {
			logger.Info("Server started at https://%s", app.Address())
		}

		// if err := http.ServeTLS(listener, app, app.TLSCertFile, app.TLSKeyFile); err != nil {
		// 	return err
		// }

		return server.ServeTLS(listener, app.TLSCertFile, app.TLSKeyFile)
	}

	if app.NetworkType == "unix" {
		logger.Info("Server started at unix://%s", app.Address())
	} else {
		logger.Info("Server started at http://%s", app.Address())
	}
	// if err := http.Serve(listener, app); err != nil {
	// 	return err
	// }

	// if only config tls ca, should reset nil
	if server.TLSConfig != nil {
		server.TLSConfig = nil
	}

	return server.Serve(listener)
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

// IsProd returns true if the app is in production mode.
func (app *Application) IsProd() bool {
	return app.Env.Get("MODE") == "production"
}

// CreateJSONRPC creates a new CreateJSONRPC handler.
func (app *Application) CreateJSONRPC(path string) jsonrpc.Server[*Context] {
	rpc := jsonrpc.NewServer[*Context]()

	app.Post(path, func(ctx *Context) {
		request, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		defer ctx.Request.Body.Close()

		response, err := rpc.Invoke(ctx, request)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.Status(http.StatusOK)
		ctx.Write(response)
	})

	return rpc
}

// Cache ...
func (app *Application) Cache() cache.Cache {
	if app.cache == nil {
		app.cache = cache.New(app.CacheConfig)
	}

	return app.cache
}

// Cron ...
func (app *Application) Cron() cron.Cron {
	if app.cron == nil {
		app.cron = cron.New()
	}

	return app.cron
}

// Queue ...
func (app *Application) Queue() queue.Queue {
	if app.queue == nil {
		app.queue = queue.New()
	}

	return app.queue
}

// Debug ...
func (app *Application) Debug() debug.Debug {
	if app.cache == nil {
		app.debug = debug.New(app.Logger)
	}

	return app.debug
}

// Address ...
func (app *Application) Address() string {
	if app.NetworkType == "unix" {
		return app.UnixDomainSocket
	} else {
		return fmt.Sprintf("%s:%d", app.Host, app.Port)
	}
}

// H is a shortcut for map[string]interface{}
type H map[string]interface{}
