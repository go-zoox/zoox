package zoox

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"time"

	"github.com/go-zoox/core-utils/strings"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/headers"
	"github.com/go-zoox/proxy"
	"github.com/go-zoox/zoox/components/application/websocket"
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
func (g *RouterGroup) Group(prefix string, cb ...GroupFunc) *RouterGroup {
	newGroup := newRouterGroup(g.app, g.prefix+prefix)
	newGroup.parent = g
	g.app.groups = append(g.app.groups, newGroup)

	for _, fn := range cb {
		fn(newGroup)
	}

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

// ProxyConfig defines the proxy config
type ProxyConfig struct {
	// internal proxy config
	proxy.SingleHostConfig

	// context proxy config
	OnRequestWithContext  func(ctx *Context) error
	OnResponseWithContext func(ctx *Context) error
}

// Proxy defines the method to proxy the request to the backend service.
//
// Example:
//
//	// default no rewrites
//	app.Proxy("/httpbin", "https://httpbin.org")
//
//	// custom rewrites
//	app.Proxy("/api/v1/tasks", "http://zmicro.services.tasks:8080", func (cfg *ProxyConfig) {
//		cfg.Rewrites = rewriter.Rewriters{
//	    {From: "/api/v1/tasks/(.*)", To: "/$1"},
//	  }
//	}))
func (g *RouterGroup) Proxy(path, target string, options ...func(cfg *ProxyConfig)) *RouterGroup {
	cfg := &ProxyConfig{}
	for _, option := range options {
		option(cfg)
	}

	handler := WrapH(proxy.NewSingleHost(target, &cfg.SingleHostConfig))

	g.Use(func(ctx *Context) {
		if strings.StartsWith(ctx.Path, path) {
			if cfg.OnRequestWithContext != nil {
				if err := cfg.OnRequestWithContext(ctx); err != nil {
					ctx.Logger.Errorf("proxy error: %s", err)
					ctx.Fail(err, 500, "proxy on request with context error")
					return
				}
			}

			handler(ctx)

			if cfg.OnResponseWithContext != nil {
				if err := cfg.OnResponseWithContext(ctx); err != nil {
					ctx.Logger.Errorf("proxy error: %s", err)
					ctx.Fail(err, 500, "proxy on response with context error")
					return
				}
			}
			return
		}

		ctx.Next()
	})

	return g
}

// WebSocket defines the method to add websocket route
// WebSocket defines the method to add websocket route
func (g *RouterGroup) WebSocket(path string, handler WsHandlerFunc, middlewares ...HandlerFunc) *RouterGroup {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	handleFunc := append(middlewares, func(ctx *Context) {
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
				go client.OnDisconnect()
			}

			conn.Close()
		}()

		// ctx.Logger.Info("ws connected")
		if client.OnConnect != nil {
			go client.OnConnect()
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
					go client.OnError(err)
				} else {
					if e, ok := err.(*gowebsocket.CloseError); ok {
						switch e.Code {
						case gowebsocket.CloseGoingAway:
							// @TODO
							// user auto leave, for example, close browser or go other page
							// we should not log as an error, it is very common.
							// action => ignored.
							// ctx.Logger.Warnf("read err: %s (type: %d)", err, mt)
						case gowebsocket.CloseAbnormalClosure:
							// @TODO
							// user close conn, we should not log as an error, it is very common.
							// action => ignored.
						default:
							ctx.Logger.Errorf("read err: %s (code: %d, type: %d)", err, e.Code, mt)
						}
					}

					// else {
					// 	ctx.Logger.Errorf("read err: %s (type: %d)", err, mt)
					// }
				}

				return
			}

			go func(mt int, message []byte) {
				defer func() {
					if err := recover(); err != nil {
						switch v := err.(type) {
						case error:
							if client.OnError != nil {
								go client.OnError(v)
							} else {
								ctx.Logger.Errorf("[onmessage] panic: %s", err)
							}
						case string:
							if client.OnError != nil {
								go client.OnError(errors.New(v))
							} else {
								ctx.Logger.Errorf("[onmessage] panic: %s", err)
							}
						default:
							ctx.Logger.Errorf("[onmessage] panic: %v", err)
						}
					}
				}()

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
			}(mt, message)
		}
	})

	g.addRoute(http.MethodGet, path, handleFunc...)

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

// JSONRPC defines the method to add jsonrpc route
func (g *RouterGroup) JSONRPC(path string, handler JSONRPCHandlerFunc) *RouterGroup {
	handler(g.app.JSONRPCRegistry())

	g.addRoute(http.MethodPost, path, func(ctx *Context) {
		request, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		defer ctx.Request.Body.Close()

		response, err := ctx.App.JSONRPCRegistry().Invoke(ctx.Context(), request)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.Status(http.StatusOK)
		ctx.Write(response)
	})

	return g
}

// Use adds a middleware to the group
func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *RouterGroup) createStaticHandler(absolutePath string, fs http.FileSystem) HandlerFunc {
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	// fix mime types
	var builtinMimeTypesLower = map[string]string{
		".html": "text/html; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".js":   "application/javascript",
		// ".ts":    "application/typescript",
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".json":  "application/json; charset=utf-8",
		".txt":   "text/plain; charset=utf-8",
		".csv":   "text/csv; charset=utf-8",
		".htm":   "text/html; charset=utf-8",
		".jpg":   "image/jpeg",
		".png":   "image/png",
		".svg":   "image/svg+xml",
		".gif":   "image/gif",
		".ico":   "image/x-icon",
		".webp":  "image/webp",
		".avif":  "image/avif",
		".bmp":   "image/x-ms-bmp",
		".wasm":  "application/wasm",
		".pdf":   "application/pdf",
		".xml":   "text/xml; charset=utf-8",
		".tar":   "application/x-tar",
		".gz":    "application/gzip",
		".zip":   "application/zip",
		".7z":    "application/x-7z-compressed",
		".rar":   "application/vnd.rar",
		".bz2":   "application/x-bzip2",
		".xz":    "application/x-xz",
		".exe":   "application/octet-stream",
		".deb":   "application/octet-stream",
		".apk":   "application/vnd.android.package-archive",
		".dmg":   "application/octet-stream",
		".iso":   "application/octet-stream",
		".img":   "application/octet-stream",
		".msi":   "application/octet-stream",
		".jar":   "application/java-archive",
		".war":   "application/java-archive",
		".ear":   "application/java-archive",
		".doc":   "application/msword",
		".ps":    "application/postscript",
		".ai":    "application/postscript",
		".eps":   "application/postscript",
		".xls":   "application/vnd.ms-excel",
		".ppt":   "application/vnd.ms-powerpoint",
		".rtf":   "application/rtf",
		".m3u8":  "application/vnd.apple.mpegurl",
		".kml":   "application/vnd.google-earth.kml+xml",
		".kmz":   "application/vnd.google-earth.kmz",
		".odg":   "application/vnd.oasis.opendocument.graphics",
		".odp":   "application/vnd.oasis.opendocument.presentation",
		".ods":   "application/vnd.oasis.opendocument.spreadsheet",
		".odt":   "application/vnd.oasis.opendocument.text",
		".pptx":  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".xlsx":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		// audio
		".mp3": "audio/mpeg",
		".ogg": "audio/ogg",
		".m4a": "audio/x-m4a",
		".ra":  "audio/x-realaudio",
		// video
		".mp4":  "video/mp4",
		".mpeg": "video/mpeg",
		".mpg":  "video/mpeg",
		".mov":  "video/quicktime",
		".webm": "video/webm",
		".flv":  "video/x-flv",
		".m4v":  "video/x-m4v",
		".mng":  "video/x-mng",
		".asx":  "video/x-ms-asf",
		".asf":  "video/x-ms-asf",
		".wmv":  "video/x-ms-wmv",
		".avi":  "video/x-msvideo",
		// ".ts":   "video/mp2t",
		".3gpp": "video/3gpp",
		".3gp":  "video/3gpp",
	}

	for k, v := range builtinMimeTypesLower {
		if err := mime.AddExtensionType(k, v); err != nil {
			panic(fmt.Errorf("failed to register mime type(%s): %s", k, err))
		}
	}

	return func(ctx *Context) {
		// file := ctx.Param().Get("filepath")
		// key := fmt.Sprintf("static_fs:%s", file)
		// if ok := ctx.Cache().Has(key); !ok {
		// 	// Check if file exists and/or is not a directory
		// 	f, err := fs.Open(file.String())
		// 	if err != nil {
		// 		// ctx.Status(http.StatusNotFound)
		// 		ctx.handlers = append(ctx.handlers, ctx.App.notfound)

		// 		ctx.Next()
		// 		return
		// 	}
		// 	f.Close()

		// 	ctx.Cache().Set(key, true, 24*time.Hour)
		// }

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
func (g *RouterGroup) Static(relativePath string, root string, options ...*StaticOptions) {
	var opts *StaticOptions
	if len(options) > 0 {
		opts = options[0]
	}

	if !strings.StartsWith(relativePath, "/") {
		root = fs.JoinCurrentDir(relativePath)
	}
	absolutePath := path.Join(g.prefix, relativePath)
	handler := g.createStaticHandler(absolutePath, http.Dir(root))

	g.Use(func(ctx *Context) {
		if ctx.Method != http.MethodGet && ctx.Method != http.MethodHead {
			ctx.Next()
			return
		}

		if !strings.StartsWith(ctx.Path, absolutePath) {
			ctx.Next()
			return
		}

		if opts != nil {
			if opts.Suffix != "" {
				ctx.Request.URL.Path = ctx.Request.URL.Path + opts.Suffix
				ctx.Request.URL.RawPath = ctx.Request.URL.RawPath + opts.Suffix
			}

			if opts.MaxAge > 0 {
				ctx.Set(headers.CacheControl, fmt.Sprintf("max-age=%d", int64(opts.MaxAge.Seconds())))
			}
		}

		handler(ctx)
	})
}

// StaticFS defines the method to serve static files
func (g *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) {
	handler := g.createStaticHandler(relativePath, fs)
	pathX := path.Join(relativePath, "/*filepath")

	//
	g.Get(pathX, handler)
	g.Head(pathX, handler)
}
