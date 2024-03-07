package zoox

import (
	"net/http"
	"strings"

	"github.com/go-zoox/headers"
	"github.com/go-zoox/logger"

	websocket "github.com/go-zoox/websocket/server"
)

const headerConnectionValueUpgrade = "upgrade"
const headerUpgradeValueWebSocket = "websocket"

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

	// handleFunc := append(opt.Middlewares, func(ctx *Context) {
	// 	ctx.Status(200)

	// 	opt.Server.ServeHTTP(ctx.Writer, ctx.Request)
	// })

	// g.addRoute(http.MethodGet, path, handleFunc...)

	logger.Info("[router] register: WS %s", path)
	g.Use(func(ctx *Context) {
		// ignore@1: method != get
		if ctx.Method != http.MethodGet {
			ctx.Next()
			return
		}

		// ignore@2: path != path
		if ctx.Path != path {
			ctx.Next()
			return
		}

		// ignore@2: connection != Upgrade
		connection := ctx.Header().Get(headers.Connection)
		if connection == "" {
			ctx.Next()
			return
		}
		if strings.ToLower(connection) != headerConnectionValueUpgrade {
			ctx.Next()
			return
		}

		// ignore@3: upgrade != websocket
		upgrade := ctx.Header().Get(headers.Upgrade)
		if upgrade == "" {
			ctx.Next()
			return
		}
		if strings.ToLower(upgrade) != headerUpgradeValueWebSocket {
			ctx.Next()
			return
		}

		opt.Server.ServeHTTP(ctx.Writer, ctx.Request)
	})

	return opt.Server, nil
}
