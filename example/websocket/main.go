package main

import (
	"github.com/go-zoox/zoox"
	zd "github.com/go-zoox/zoox/default"
)

func main() {
	app := zd.Default()

	app.WebSocket("/ws", func(ctx *zoox.Context, client *zoox.WebSocketClient) {
		client.OnConnect = func() {
			ctx.Logger.Info("Connected")
		}

		client.OnDisconnect = func() {
			ctx.Logger.Info("Disconnected")
		}

		client.OnMessage = func(typ int, msg []byte) {
			ctx.Logger.Info("Message: %d %s", typ, msg)
			// client.Disconnect()

			client.Write(typ, msg)
		}
	})

	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "hello",
		})
	})

	app.Run(":8080")
}
