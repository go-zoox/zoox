package zoox

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	ctx  *Context
	conn *websocket.Conn

	OnConnect       func()
	OnDisconnect    func()
	OnMessage       func(typ int, msg []byte)
	OnTextMessage   func(msg []byte)
	OnBinaryMessage func(msg []byte)
	OnError         func(err error)
}

func newWebSocket(ctx *Context, conn *websocket.Conn) *WebSocketClient {
	return &WebSocketClient{
		ctx:  ctx,
		conn: conn,
	}
}

func (c *WebSocketClient) Disconnect() {
	c.conn.Close()
}

func (c *WebSocketClient) Write(typ int, msg []byte) error {
	return c.conn.WriteMessage(typ, msg)
}

func (c *WebSocketClient) WriteText(msg string) error {
	return c.Write(websocket.TextMessage, []byte(msg))
}

func (c *WebSocketClient) WriteBinary(msg []byte) error {
	return c.Write(websocket.BinaryMessage, msg)
}

func (c *WebSocketClient) WriteJSON(msg interface{}) error {
	if bytes, err := json.Marshal(msg); err != nil {
		return err
	} else {
		return c.Write(websocket.TextMessage, bytes)
	}
}
