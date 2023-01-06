package zoox

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"text/template"

	"github.com/go-zoox/kv/typing"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox/rpc/jsonrpc"
	"github.com/gorilla/websocket"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(ctx *Context)

// Middleware defines the signature of the middleware function.
type Middleware = HandlerFunc

// WsHandlerFunc defines the websocket handler used by gee
type WsHandlerFunc func(ctx *Context, client *WebSocketClient)

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
	LogLevel  string
	//
	CacheConfig *typing.Config
	cache       *Cache
	//
	cron  *Cron
	queue *Queue
	//
	Env    *Env
	Logger *logger.Logger
	// Debug
	debug *Debug

	// TLS Certificate
	TLSCertFile string
	// TLS Private Key
	TLSKeyFile string
	// TLS Ca Certificate
	TLSCaCertFile string
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

	app.Env = newEnv()

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

// Run defines the method to start the server
// Example:
//
//		IP:
//	   default(http://0.0.0.0:8080): Run(":8080")
//		 port(http://0.0.0.0:8888): Run(":8888")
//		 host+port(http://127.0.0.1:8888): Run("127.0.0.1:8888")
//
//		Unix Domain Socket:
//			/tmp/xxx.sock: Run("unix:///tmp/xxx.sock")
func (app *Application) Run(addr ...string) error {
	addrX := ":8080"
	if len(addr) > 0 && addr[0] != "" {
		addrX = addr[0]
	}

	// if err := http.ListenAndServe(addrX, app); err != nil {
	// 	return err
	// }

	typ := "tcp"
	if addrX[0] == '/' {
		typ = "unix"
	} else if strings.HasPrefix(addrX, "unix://") {
		typ = "unix"
		addrX = addrX[7:]
	}

	listener, err := net.Listen(typ, addrX)
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{
		Addr:    ":8088",
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

		if typ == "unix" {
			logger.Info("Server started at unixs://%s", addrX)
		} else {
			logger.Info("Server started at https://%s", addrX)
		}

		// if err := http.ServeTLS(listener, app, app.TLSCertFile, app.TLSKeyFile); err != nil {
		// 	return err
		// }

		return server.ServeTLS(listener, app.TLSCertFile, app.TLSKeyFile)
	}

	if typ == "unix" {
		logger.Info("Server started at unix://%s", addrX)
	} else {
		logger.Info("Server started at http://%s", addrX)
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
func (app *Application) Cache() *Cache {
	if app.cache == nil {
		app.cache = newCache(app)
	}

	return app.cache
}

// Cron ...
func (app *Application) Cron() *Cron {
	if app.cron == nil {
		app.cron = newCron()
	}

	return app.cron
}

// Queue ...
func (app *Application) Queue() *Queue {
	if app.queue == nil {
		app.queue = newQueue()
	}

	return app.queue
}

// Debug ...
func (app *Application) Debug() *Debug {
	if app.cache == nil {
		app.debug = newDebug(app)
	}

	return app.debug
}

// H is a shortcut for map[string]interface{}
type H map[string]interface{}
