# 实时应用示例

这是一个使用 WebSocket 构建的实时聊天应用示例。

## 项目结构

```
real-time-app/
├── main.go
├── go.mod
├── handlers/
│   └── chat.go
└── models/
    └── room.go
```

## 完整代码

### main.go

```go
package main

import (
	"sync"
	
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
	"real-time-app/handlers"
)

func main() {
	app := zoox.New()
	
	// 全局中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	
	// 静态文件
	app.Static("/static", "./public")
	
	// WebSocket 路由
	app.Get("/ws", handlers.HandleWebSocket)
	
	// 首页
	app.Get("/", func(ctx *zoox.Context) {
		ctx.RenderHTML("./public/index.html")
	})
	
	app.Run(":8080")
}
```

### models/room.go

```go
package models

import (
	"sync"
	
	websocket "github.com/go-zoox/websocket/server"
)

// ChatRoom 聊天室
type ChatRoom struct {
	clients map[*websocket.Server]string // client -> username
	mutex   sync.RWMutex
}

var room = &ChatRoom{
	clients: make(map[*websocket.Server]string),
}

// AddClient 添加客户端
func (r *ChatRoom) AddClient(client *websocket.Server, username string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.clients[client] = username
}

// RemoveClient 移除客户端
func (r *ChatRoom) RemoveClient(client *websocket.Server) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.clients, client)
}

// Broadcast 广播消息
func (r *ChatRoom) Broadcast(message string, exclude *websocket.Server) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	for client := range r.clients {
		if client != exclude {
			client.WriteText(message)
		}
	}
}

// GetUsername 获取用户名
func (r *ChatRoom) GetUsername(client *websocket.Server) string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.clients[client]
}

// GetRoom 获取聊天室实例
func GetRoom() *ChatRoom {
	return room
}
```

### handlers/chat.go

```go
package handlers

import (
	"encoding/json"
	
	"github.com/go-zoox/zoox"
	"real-time-app/models"
)

// Message 消息结构
type Message struct {
	Type     string `json:"type"`     // message, join, leave
	Username string `json:"username"`
	Content  string `json:"content"`
	Time     string `json:"time"`
}

// HandleWebSocket 处理 WebSocket 连接
func HandleWebSocket(ctx *zoox.Context) {
	server, err := ctx.App.WebSocket("/ws")
	if err != nil {
		ctx.Error(500, "Failed to create WebSocket")
		return
	}
	
	room := models.GetRoom()
	
	// 获取用户名（从查询参数）
	username := ctx.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}
	
	// 连接建立
	server.OnConnect(func() {
		room.AddClient(server, username)
		
		// 通知其他用户
		msg := Message{
			Type:     "join",
			Username: username,
			Content:  username + " joined the chat",
		}
		sendMessage(room, msg, server)
		
		// 发送欢迎消息
		welcome := Message{
			Type:     "system",
			Content:  "Welcome to the chat room!",
		}
		data, _ := json.Marshal(welcome)
		server.WriteText(string(data))
	})
	
	// 接收消息
	server.OnText(func(message string) {
		var msg Message
		if err := json.Unmarshal([]byte(message), &msg); err != nil {
			return
		}
		
		msg.Type = "message"
		msg.Username = username
		
		// 广播消息
		sendMessage(room, msg, server)
	})
	
	// 连接断开
	server.OnDisconnect(func() {
		username := room.GetUsername(server)
		room.RemoveClient(server)
		
		// 通知其他用户
		msg := Message{
			Type:     "leave",
			Username: username,
			Content:  username + " left the chat",
		}
		sendMessage(room, msg, nil)
	})
	
	// 错误处理
	server.OnError(func(err error) {
		ctx.Logger.Errorf("WebSocket error: %v", err)
	})
}

// sendMessage 发送消息
func sendMessage(room *models.ChatRoom, msg Message, exclude *websocket.Server) {
	data, _ := json.Marshal(msg)
	room.Broadcast(string(data), exclude)
}
```

### public/index.html

```html
<!DOCTYPE html>
<html>
<head>
	<title>Real-time Chat</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			max-width: 800px;
			margin: 0 auto;
			padding: 20px;
		}
		#messages {
			border: 1px solid #ccc;
			height: 400px;
			overflow-y: auto;
			padding: 10px;
			margin-bottom: 10px;
		}
		.message {
			margin-bottom: 10px;
		}
		.message.system {
			color: #666;
			font-style: italic;
		}
		.message.join, .message.leave {
			color: #999;
		}
		#input {
			display: flex;
			gap: 10px;
		}
		#message {
			flex: 1;
			padding: 10px;
		}
		#send {
			padding: 10px 20px;
		}
	</style>
</head>
<body>
	<h1>Real-time Chat</h1>
	<div id="messages"></div>
	<div id="input">
		<input type="text" id="message" placeholder="Type a message..." />
		<button id="send">Send</button>
	</div>
	
	<script>
		const username = prompt("Enter your username:") || "Anonymous";
		const ws = new WebSocket(`ws://localhost:8080/ws?username=${encodeURIComponent(username)}`);
		
		const messagesDiv = document.getElementById("messages");
		const messageInput = document.getElementById("message");
		const sendButton = document.getElementById("send");
		
		function addMessage(msg) {
			const div = document.createElement("div");
			div.className = `message ${msg.type}`;
			
			if (msg.type === "system") {
				div.textContent = msg.content;
			} else if (msg.type === "join" || msg.type === "leave") {
				div.textContent = msg.content;
			} else {
				div.innerHTML = `<strong>${msg.username}:</strong> ${msg.content}`;
			}
			
			messagesDiv.appendChild(div);
			messagesDiv.scrollTop = messagesDiv.scrollHeight;
		}
		
		ws.onmessage = function(event) {
			const msg = JSON.parse(event.data);
			addMessage(msg);
		};
		
		ws.onerror = function(error) {
			console.error("WebSocket error:", error);
		};
		
		ws.onclose = function() {
			addMessage({type: "system", content: "Connection closed"});
		};
		
		function sendMessage() {
			const content = messageInput.value.trim();
			if (content === "") return;
			
			const msg = {
				type: "message",
				content: content
			};
			
			ws.send(JSON.stringify(msg));
			messageInput.value = "";
		}
		
		sendButton.addEventListener("click", sendMessage);
		messageInput.addEventListener("keypress", function(e) {
			if (e.key === "Enter") {
				sendMessage();
			}
		});
	</script>
</body>
</html>
```

## 运行和测试

### 启动服务器

```bash
go mod init real-time-app
go get github.com/go-zoox/zoox
go get github.com/go-zoox/websocket
go run main.go
```

### 测试

1. 打开浏览器访问 `http://localhost:8080`
2. 输入用户名
3. 打开多个浏览器标签页模拟多个用户
4. 发送消息测试实时通信

## 特性说明

1. **实时通信** - 使用 WebSocket 实现实时消息传递
2. **用户管理** - 跟踪在线用户
3. **消息广播** - 向所有客户端广播消息
4. **连接管理** - 处理连接建立和断开
5. **错误处理** - 完善的错误处理机制

## 扩展建议

1. **房间功能** - 支持多个聊天室
2. **私聊功能** - 支持用户之间的私聊
3. **消息历史** - 保存和显示历史消息
4. **用户认证** - 添加用户认证功能
5. **文件传输** - 支持文件传输功能

## 下一步

- 🏗️ 查看 [微服务示例](microservice.md) - 微服务架构
- 📡 学习 [WebSocket 文档](../advanced/websocket.md) - WebSocket 详细说明
- 📚 阅读 [最佳实践](../best-practices.md) - 开发建议

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
