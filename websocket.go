package zoox

import (
	"encoding/json"

	"github.com/go-zoox/zoox/rpc/jsonrpc"
	"github.com/gorilla/websocket"
)

// WebSocketClient ...
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

// Disconnect ...
func (c *WebSocketClient) Disconnect() {
	c.conn.Close()
}

// Write ...
func (c *WebSocketClient) Write(typ int, msg []byte) error {
	return c.conn.WriteMessage(typ, msg)
}

// WriteText ...
func (c *WebSocketClient) WriteText(msg string) error {
	return c.Write(websocket.TextMessage, []byte(msg))
}

// WriteBytes ...
func (c *WebSocketClient) WriteBytes(msg []byte) error {
	return c.Write(websocket.TextMessage, msg)
}

// WriteBinary ...
func (c *WebSocketClient) WriteBinary(msg []byte) error {
	return c.Write(websocket.BinaryMessage, msg)
}

// WriteJSON ...
func (c *WebSocketClient) WriteJSON(msg interface{}) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.Write(websocket.TextMessage, bytes)
}

// CreateJSONRPC ...
func (c *WebSocketClient) CreateJSONRPC() jsonrpc.Server[any] {
	rpc := jsonrpc.NewServer[any]()

	onTextMessage := []func([]byte){}
	if c.OnTextMessage != nil {
		onTextMessage = append(onTextMessage, c.OnTextMessage)
	}

	onTextMessage = append(onTextMessage, func(msg []byte) {
		resp, err := rpc.Invoke(c, msg)
		if err != nil {
			if c.OnError != nil {
				c.OnError(err)
			} else {
				c.ctx.Logger.Error("ws error: %s", err)
			}
		}

		c.WriteBytes(resp)
	})

	c.OnTextMessage = func(msg []byte) {
		for _, f := range onTextMessage {
			f(msg)
		}
	}

	return rpc
}
