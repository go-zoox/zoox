package zoox

import (
	"fmt"
	"net/http"

	"github.com/go-zoox/core-utils/regexp"
	"github.com/go-zoox/core-utils/strings"
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

	logger.Info("[router] register: %8s %s", "WS", g.prefix+path)
	matchPath := func(requestPath string) (ok bool) {
		re := fmt.Sprintf("^%s$", g.prefix+path)
		if strings.Contains(re, ":") {
			re = strings.ReplaceAllFunc(re, ":\\w+", func(b []byte) []byte {
				return []byte("\\w+")
			})
		} else if strings.Contains(re, "{") {
			re = strings.ReplaceAllFunc(re, "{.*}", func(b []byte) []byte {
				return []byte("\\w+")
			})
		}

		return regexp.Match(re, requestPath)
	}

	g.Use(func(ctx *Context) {
		// ignore@1: method != get
		if ctx.Method != http.MethodGet {
			ctx.Next()
			return
		}

		// ignore@2: connection != Upgrade
		connection := ctx.Header().Get(headers.Connection)
		if connection == "" {
			ctx.Next()
			return
		} else if strings.ToLower(connection) != headerConnectionValueUpgrade {
			ctx.Next()
			return
		}

		// ignore@3: upgrade != websocket
		upgrade := ctx.Header().Get(headers.Upgrade)
		if upgrade == "" {
			ctx.Next()
			return
		} else if strings.ToLower(upgrade) != headerUpgradeValueWebSocket {
			ctx.Next()
			return
		}

		// ignore@4: path != path
		if ok := matchPath(ctx.Path); !ok {
			ctx.Next()
			return
		}

		// @TODO rewrites handlers
		//	=> ignore old handlers
		//	=> only use websocket handlers
		ctx.index = -1
		ctx.handlers = append(opt.Middlewares, func(ctx *Context) {
			opt.Server.ServeHTTP(ctx.Writer, ctx.Request)
		})

		ctx.Next()
	})

	return opt.Server, nil
}
