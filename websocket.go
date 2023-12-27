package zoox

import (
	"net/http"

	websocket "github.com/go-zoox/websocket/server"
)

// WebSocketOption ...
type WebSocketOption struct {
	Server      websocket.Server
	Middlewares []HandlerFunc
}

// WebSocket defines the method to add websocket route
func (g *RouterGroup) WebSocket(path string, opts ...func(opt *WebSocketOption)) (websocket.Server, error) {
	opt := &WebSocketOption{}
	for _, o := range opts {
		o(opt)
	}

	if opt.Server == nil {
		server, err := websocket.New()
		if err != nil {
			return nil, err
		}
		opt.Server = server
	}

	handleFunc := append(opt.Middlewares, func(ctx *Context) {
		ctx.Status(200)

		opt.Server.ServeHTTP(ctx.Writer, ctx.Request)
	})

	g.addRoute(http.MethodGet, path, handleFunc...)

	return opt.Server, nil
}
