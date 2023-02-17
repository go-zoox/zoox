package zoox

import (
	"fmt"
	"mime"
	"net/http"
	"path"
	"time"

	"github.com/go-zoox/zoox/components/context/websocket"
	gowebsocket "github.com/gorilla/websocket"
)

var anyMethods = []string{
	http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
	http.MethodHead, http.MethodOptions, http.MethodConnect,
	http.MethodTrace,
}

// RouterGroup is a group of routes.
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	app         *Application
}

func newRouterGroup(app *Application, prefix string) *RouterGroup {
	return &RouterGroup{
		app:    app,
		prefix: prefix,
	}
}

// Group defines a new router group
func (g *RouterGroup) Group(prefix string) *RouterGroup {
	newGroup := newRouterGroup(g.app, g.prefix+prefix)
	newGroup.parent = g
	g.app.groups = append(g.app.groups, newGroup)
	return newGroup
}

func (g *RouterGroup) addRoute(method string, path string, handler ...HandlerFunc) {
	pathX := g.prefix + path
	g.app.router.addRoute(method, pathX, handler...)
}

// Get defines the method to add GET request
func (g *RouterGroup) Get(path string, handler ...HandlerFunc) *RouterGroup {
	g.addRoute(http.MethodGet, path, handler...)
	return g
}

// Post defines the method to add POST request
func (g *RouterGroup) Post(path string, handler ...HandlerFunc) *RouterGroup {
	g.addRoute(http.MethodPost, path, handler...)
	return g
}

// Put defines the method to add PUT request
func (g *RouterGroup) Put(path string, handler ...HandlerFunc) *RouterGroup {
	g.addRoute(http.MethodPut, path, handler...)
	return g
}

// Patch defines the method to add PATCH request
func (g *RouterGroup) Patch(path string, handler ...HandlerFunc) *RouterGroup {
	g.addRoute(http.MethodPatch, path, handler...)
	return g
}

// Delete defines the method to add DELETE request
func (g *RouterGroup) Delete(path string, handler ...HandlerFunc) *RouterGroup {
	g.addRoute(http.MethodDelete, path, handler...)
	return g
}

// Head defines the method to add HEAD request
func (g *RouterGroup) Head(path string, handler ...HandlerFunc) *RouterGroup {
	g.addRoute(http.MethodHead, path, handler...)
	return g
}

// Options defines the method to add OPTIONS request
func (g *RouterGroup) Options(path string, handler ...HandlerFunc) *RouterGroup {
	g.addRoute(http.MethodOptions, path, handler...)
	return g
}

// Connect defines the method to add CONNECT request
func (g *RouterGroup) Connect(path string, handler ...HandlerFunc) *RouterGroup {
	g.addRoute(http.MethodConnect, path, handler...)
	return g
}

// Any defines all request methods (anyMethods)
func (g *RouterGroup) Any(path string, handler ...HandlerFunc) *RouterGroup {
	for _, method := range anyMethods {
		g.addRoute(method, path, handler...)
	}
	return g
}

// WebSocket defines the method to add websocket route
func (g *RouterGroup) WebSocket(path string, handler WsHandlerFunc) *RouterGroup {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	g.addRoute(http.MethodGet, path, func(ctx *Context) {
		ctx.Status(200)

		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.Logger.Errorf("ws error: %s", err)
			return
		}
		defer conn.Close()

		client := websocket.New(conn)
		handler(ctx, client)

		defer func() {
			if client.OnDisconnect != nil {
				client.OnDisconnect()
			}

			conn.Close()
		}()

		// ctx.Logger.Info("ws connected")
		if client.OnConnect != nil {
			client.OnConnect()
		}

		for {
			mt, message, err := conn.ReadMessage()
			// if mt == -1 {
			// 	ctx.Logger.Info("xxx disconnected: %s", message)
			// 	if client.OnDisconnect != nil {
			// 		client.OnDisconnect()
			// 	}
			// } else if err != nil {
			// 	// ctx.Logger.Info("read err: %s %d", err, mt)

			// 	if client.OnError != nil {
			// 		client.OnError(err)
			// 	}
			// 	return
			// }

			if err != nil {
				if client.OnError != nil {
					client.OnError(err)
				} else {
					if e, ok := err.(*gowebsocket.CloseError); ok {
						switch e.Code {
						case gowebsocket.CloseGoingAway:
							// @TODO
							// user auto leave, for example, close browser or go other page
							// we should not log as an error, it is very common.
							// action => ignored.
							// ctx.Logger.Warnf("read err: %s (type: %d)", err, mt)
						default:
							ctx.Logger.Errorf("read err: %s (type: %d)", err, mt)
						}
					} else {
						ctx.Logger.Errorf("read err: %s (type: %d)", err, mt)
					}
				}

				return
			}

			switch mt {
			case websocket.TextMessage:
				if client.OnTextMessage != nil {
					client.OnTextMessage(message)
				}
			case websocket.BinaryMessage:
				if client.OnBinaryMessage != nil {
					client.OnBinaryMessage(message)
				}
			default:
				ctx.Logger.Warn("unknown message type: %d", mt)
			}

			if client.OnMessage != nil {
				client.OnMessage(mt, message)
			}
		}
	})

	return g
}

// WebSocketGorilla defines the method to add websocket route
func (g *RouterGroup) WebSocketGorilla(path string, handler WsGorillaHandlerFunc) *RouterGroup {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	g.addRoute(http.MethodGet, path, func(ctx *Context) {
		ctx.Status(200)

		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.Logger.Errorf("ws error: %s", err)
			return
		}
		defer conn.Close()

		handler(ctx, conn)
	})

	return g
}

// Use adds a middleware to the group
func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(g.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	// fix mime types
	var builtinMimeTypesLower = map[string]string{
		".css":  "text/css; charset=utf-8",
		".gif":  "image/gif",
		".htm":  "text/html; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".jpg":  "image/jpeg",
		".js":   "application/javascript",
		".wasm": "application/wasm",
		".pdf":  "application/pdf",
		".png":  "image/png",
		".svg":  "image/svg+xml",
		".xml":  "text/xml; charset=utf-8",
	}

	for k, v := range builtinMimeTypesLower {
		if err := mime.AddExtensionType(k, v); err != nil {
			panic(fmt.Errorf("failed to register mime type(%s): %s", k, err))
		}
	}

	return func(ctx *Context) {
		file := ctx.Param().Get("filepath")
		// Check if file exists and/or is not a directory
		f, err := fs.Open(file.String())
		if err != nil {
			// ctx.Status(http.StatusNotFound)
			ctx.handlers = append(ctx.handlers, ctx.App.notfound)

			ctx.Next()
			return
		}
		f.Close()

		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

// StaticOptions is the options for static method
type StaticOptions struct {
	Gzip         bool
	Md5          bool
	CacheControl string
	MaxAge       time.Duration
	Index        bool
	Suffix       string
}

// Static defines the method to serve static files
func (g *RouterGroup) Static(relativePath string, root string, options ...StaticOptions) {
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	pathX := path.Join(relativePath, "/*filepath")

	//
	g.Get(pathX, handler)
}

// StaticFS defines the method to serve static files
func (g *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) {
	handler := g.createStaticHandler(relativePath, fs)
	pathX := path.Join(relativePath, "/*filepath")

	//
	g.Get(pathX, handler)
	g.Head(pathX, handler)
}
