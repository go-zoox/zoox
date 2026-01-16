# WebSocket 支持

Zoox 提供了完整的 WebSocket 支持，可以轻松构建实时应用。

## 基本用法

### 创建 WebSocket 路由

```go
package main

import (
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()
	
	// 创建 WebSocket 路由
	server, err := app.WebSocket("/ws")
	if err != nil {
		panic(err)
	}
	
	// 处理消息
	server.OnMessage(func(message []byte) {
		// 回显消息
		server.WriteText("Echo: " + string(message))
	})
	
	app.Run(":8080")
}
```

**说明**: WebSocket 实现参考 `websocket.go:24-108`。

## WebSocket 事件

### OnMessage - 接收消息

```go
server.OnMessage(func(message []byte) {
	// 处理文本或二进制消息
	fmt.Println("Received:", string(message))
})
```

### OnText - 接收文本消息

```go
server.OnText(func(message string) {
	// 处理文本消息
	fmt.Println("Text:", message)
})
```

### OnBinary - 接收二进制消息

```go
server.OnBinary(func(message []byte) {
	// 处理二进制消息
	fmt.Println("Binary:", len(message), "bytes")
})
```

### OnConnect - 连接建立

```go
server.OnConnect(func() {
	fmt.Println("Client connected")
})
```

### OnDisconnect - 连接断开

```go
server.OnDisconnect(func() {
	fmt.Println("Client disconnected")
})
```

### OnError - 错误处理

```go
server.OnError(func(err error) {
	fmt.Println("Error:", err)
})
```

## 发送消息

### 发送文本消息

```go
server.WriteText("Hello, Client!")
```

### 发送二进制消息

```go
server.WriteBinary([]byte("Binary data"))
```

### 发送 JSON 消息

```go
data := zoox.H{
	"type": "message",
	"content": "Hello",
}
server.WriteJSON(data)
```

## WebSocket 中间件

为 WebSocket 连接添加中间件：

```go
server, err := app.WebSocket("/ws", func(opt *zoox.WebSocketOption) {
	// 添加中间件
	opt.Middlewares = []zoox.HandlerFunc{
		func(ctx *zoox.Context) {
			// 验证连接
			token := ctx.Query().Get("token")
			if token == "" {
				ctx.Error(401, "Unauthorized")
				return
			}
			ctx.Next()
		},
	}
})
```

## 完整示例

### 聊天室示例

```go
package main

import (
	"sync"
	
	"github.com/go-zoox/zoox"
)

type ChatRoom struct {
	clients map[*zoox.WebSocketServer]bool
	mutex   sync.RWMutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients: make(map[*zoox.WebSocketServer]bool),
	}
}

func (room *ChatRoom) AddClient(server *zoox.WebSocketServer) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	room.clients[server] = true
}

func (room *ChatRoom) RemoveClient(server *zoox.WebSocketServer) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	delete(room.clients, server)
}

func (room *ChatRoom) Broadcast(message string) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	
	for client := range room.clients {
		client.WriteText(message)
	}
}

func main() {
	app := zoox.New()
	room := NewChatRoom()
	
	server, _ := app.WebSocket("/ws")
	
	server.OnConnect(func() {
		room.AddClient(server)
		room.Broadcast("User joined")
	})
	
	server.OnDisconnect(func() {
		room.RemoveClient(server)
		room.Broadcast("User left")
	})
	
	server.OnText(func(message string) {
		// 广播消息到所有客户端
		room.Broadcast(message)
	})
	
	app.Run(":8080")
}
```

### 客户端示例（JavaScript）

```html
<!DOCTYPE html>
<html>
<head>
	<title>WebSocket Chat</title>
</head>
<body>
	<div id="messages"></div>
	<input type="text" id="message" />
	<button onclick="send()">Send</button>
	
	<script>
		const ws = new WebSocket("ws://localhost:8080/ws");
		
		ws.onmessage = function(event) {
			const div = document.createElement("div");
			div.textContent = event.data;
			document.getElementById("messages").appendChild(div);
		};
		
		function send() {
			const input = document.getElementById("message");
			ws.send(input.value);
			input.value = "";
		}
	</script>
</body>
</html>
```

## 路由参数

WebSocket 路由也支持路由参数：

```go
server, err := app.WebSocket("/ws/:roomId")
if err != nil {
	panic(err)
}

server.OnConnect(func() {
	// 在中间件中获取参数
	// roomId := ctx.Param().Get("roomId")
})
```

## 认证

在 WebSocket 连接时进行认证：

```go
server, err := app.WebSocket("/ws", func(opt *zoox.WebSocketOption) {
	opt.Middlewares = []zoox.HandlerFunc{
		func(ctx *zoox.Context) {
			// 从查询参数获取 token
			token := ctx.Query().Get("token")
			if token == "" {
				ctx.Error(401, "Unauthorized")
				return
			}
			
			// 验证 token
			// ...
			
			ctx.Next()
		},
	}
})
```

客户端连接：

```javascript
const token = "your-token";
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);
```

## 心跳检测

实现心跳检测保持连接：

```go
server.OnConnect(func() {
	// 启动心跳
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				server.WriteText(`{"type":"ping"}`)
			}
		}
	}()
})

server.OnText(func(message string) {
	// 处理心跳响应
	if message == `{"type":"pong"}` {
		// 心跳响应
		return
	}
	
	// 处理其他消息
})
```

## 错误处理

```go
server.OnError(func(err error) {
	log.Printf("WebSocket error: %v", err)
	// 处理错误
})
```

## 性能优化

### 1. 连接池管理

```go
type ConnectionManager struct {
	connections map[string]*zoox.WebSocketServer
	mutex       sync.RWMutex
}

func (cm *ConnectionManager) Add(id string, server *zoox.WebSocketServer) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.connections[id] = server
}

func (cm *ConnectionManager) Broadcast(message string) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	for _, server := range cm.connections {
		server.WriteText(message)
	}
}
```

### 2. 消息队列

对于高并发场景，使用消息队列：

```go
type MessageQueue struct {
	messages chan string
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		messages: make(chan string, 1000),
	}
}

func (mq *MessageQueue) Process(server *zoox.WebSocketServer) {
	for message := range mq.messages {
		server.WriteText(message)
	}
}
```

## 下一步

- 📡 学习 [Server-Sent Events (SSE)](sse.md) - 另一种实时通信方式
- 🔌 查看 [中间件使用](../guides/middleware.md) - 为 WebSocket 添加中间件
- 🚀 探索 [其他高级功能](jsonrpc.md) - JSON-RPC、代理等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
