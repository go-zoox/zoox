package zoox

import (
	"net/http"

	"github.com/go-errors/errors"
	"github.com/go-zoox/zoox/components/application/websocket"
	gowebsocket "github.com/gorilla/websocket"
)

type WebSocketOption struct {
	Middlewares []HandlerFunc
}

// WebSocket defines the method to add websocket route
func (g *RouterGroup) WebSocket(path string, handler WsHandlerFunc, opts ...func(opt *WebSocketOption)) *RouterGroup {
	opt := &WebSocketOption{}
	for _, f := range opts {
		f(opt)
	}

	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	handleFunc := append(opt.Middlewares, func(ctx *Context) {
		ctx.Status(200)

		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.Logger.Errorf("ws error: %s", err)
			return
		}
		defer conn.Close()

		client := websocket.New(conn)
		defer func() {
			if client.OnDisconnect != nil {
				go client.OnDisconnect()
			}
		}()

		handler(ctx, client)

		// ctx.Logger.Info("ws connected")
		if client.OnConnect != nil {
			go client.OnConnect()
		}

		for {
			mt, message, err := conn.ReadMessage()
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
