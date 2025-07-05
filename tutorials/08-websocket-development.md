# Tutorial 08: WebSocket Development

## 📖 Overview

Learn to build real-time applications with WebSockets in Zoox. This tutorial covers WebSocket setup, connection management, message handling, and building interactive real-time features.

## 🎯 Learning Objectives

- Set up WebSocket connections
- Manage client connections and rooms
- Handle real-time messaging
- Build interactive applications
- Implement connection pooling and scaling

## 📋 Prerequisites

- Completed [Tutorial 01: Getting Started](./01-getting-started.md)
- Basic understanding of WebSockets
- Familiarity with JavaScript and HTML

## 🚀 Getting Started

### Basic WebSocket Setup

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/go-zoox/zoox"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins in development
    },
}

func main() {
    app := zoox.New()
    
    // WebSocket endpoint
    app.Get("/ws", func(ctx *zoox.Context) {
        conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
        if err != nil {
            log.Printf("WebSocket upgrade error: %v", err)
            return
        }
        defer conn.Close()
        
        // Handle messages
        for {
            messageType, message, err := conn.ReadMessage()
            if err != nil {
                log.Printf("Read error: %v", err)
                break
            }
            
            log.Printf("Received: %s", message)
            
            // Echo the message back
            if err := conn.WriteMessage(messageType, message); err != nil {
                log.Printf("Write error: %v", err)
                break
            }
        }
    })
    
    // Serve WebSocket client
    app.Get("/", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>WebSocket Test</title>
        </head>
        <body>
            <div id="messages"></div>
            <input type="text" id="messageInput" placeholder="Enter message">
            <button onclick="sendMessage()">Send</button>
            
            <script>
                const ws = new WebSocket('ws://localhost:8080/ws');
                const messages = document.getElementById('messages');
                const input = document.getElementById('messageInput');
                
                ws.onmessage = function(event) {
                    const div = document.createElement('div');
                    div.textContent = 'Received: ' + event.data;
                    messages.appendChild(div);
                };
                
                function sendMessage() {
                    if (input.value) {
                        ws.send(input.value);
                        input.value = '';
                    }
                }
                
                input.addEventListener('keypress', function(e) {
                    if (e.key === 'Enter') {
                        sendMessage();
                    }
                });
            </script>
        </body>
        </html>
        `
        ctx.HTML(200, html, nil)
    })
    
    app.Listen(":8080")
}
```

### Advanced WebSocket Hub

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"
    
    "github.com/go-zoox/zoox"
    "github.com/gorilla/websocket"
)

type Message struct {
    Type    string      `json:"type"`
    Content interface{} `json:"content"`
    From    string      `json:"from,omitempty"`
    To      string      `json:"to,omitempty"`
    Room    string      `json:"room,omitempty"`
    Time    time.Time   `json:"time"`
}

type Client struct {
    ID       string
    Conn     *websocket.Conn
    Hub      *Hub
    Send     chan Message
    Rooms    map[string]bool
    UserInfo map[string]interface{}
}

type Room struct {
    ID      string
    Clients map[*Client]bool
    History []Message
}

type Hub struct {
    clients    map[*Client]bool
    rooms      map[string]*Room
    register   chan *Client
    unregister chan *Client
    broadcast  chan Message
    mutex      sync.RWMutex
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        rooms:      make(map[string]*Room),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan Message),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.registerClient(client)
            
        case client := <-h.unregister:
            h.unregisterClient(client)
            
        case message := <-h.broadcast:
            h.handleMessage(message)
        }
    }
}

func (h *Hub) registerClient(client *Client) {
    h.mutex.Lock()
    defer h.mutex.Unlock()
    
    h.clients[client] = true
    log.Printf("Client %s connected", client.ID)
    
    // Send welcome message
    welcome := Message{
        Type:    "welcome",
        Content: map[string]interface{}{
            "id":      client.ID,
            "message": "Connected to WebSocket server",
        },
        Time: time.Now(),
    }
    
    select {
    case client.Send <- welcome:
    default:
        close(client.Send)
        delete(h.clients, client)
    }
}

func (h *Hub) unregisterClient(client *Client) {
    h.mutex.Lock()
    defer h.mutex.Unlock()
    
    if _, ok := h.clients[client]; ok {
        // Remove from all rooms
        for roomID := range client.Rooms {
            if room, exists := h.rooms[roomID]; exists {
                delete(room.Clients, client)
                
                // Notify room about user leaving
                leaveMsg := Message{
                    Type: "user_left",
                    Content: map[string]interface{}{
                        "user_id": client.ID,
                        "room_id": roomID,
                    },
                    Time: time.Now(),
                }
                h.broadcastToRoom(roomID, leaveMsg)
            }
        }
        
        delete(h.clients, client)
        close(client.Send)
        log.Printf("Client %s disconnected", client.ID)
    }
}

func (h *Hub) handleMessage(message Message) {
    switch message.Type {
    case "join_room":
        h.handleJoinRoom(message)
    case "leave_room":
        h.handleLeaveRoom(message)
    case "room_message":
        h.handleRoomMessage(message)
    case "private_message":
        h.handlePrivateMessage(message)
    case "broadcast":
        h.handleBroadcast(message)
    }
}

func (h *Hub) handleJoinRoom(message Message) {
    roomID := message.Room
    if roomID == "" {
        return
    }
    
    h.mutex.Lock()
    defer h.mutex.Unlock()
    
    // Create room if it doesn't exist
    if _, exists := h.rooms[roomID]; !exists {
        h.rooms[roomID] = &Room{
            ID:      roomID,
            Clients: make(map[*Client]bool),
            History: make([]Message, 0),
        }
    }
    
    // Find client and add to room
    for client := range h.clients {
        if client.ID == message.From {
            h.rooms[roomID].Clients[client] = true
            client.Rooms[roomID] = true
            
            // Send room history
            history := Message{
                Type:    "room_history",
                Content: h.rooms[roomID].History,
                Room:    roomID,
                Time:    time.Now(),
            }
            
            select {
            case client.Send <- history:
            default:
            }
            
            // Notify room about new user
            joinMsg := Message{
                Type: "user_joined",
                Content: map[string]interface{}{
                    "user_id": client.ID,
                    "room_id": roomID,
                },
                Room: roomID,
                Time: time.Now(),
            }
            h.broadcastToRoom(roomID, joinMsg)
            break
        }
    }
}

func (h *Hub) handleRoomMessage(message Message) {
    roomID := message.Room
    if roomID == "" {
        return
    }
    
    h.mutex.Lock()
    defer h.mutex.Unlock()
    
    if room, exists := h.rooms[roomID]; exists {
        // Add to history
        room.History = append(room.History, message)
        
        // Keep only last 100 messages
        if len(room.History) > 100 {
            room.History = room.History[1:]
        }
        
        // Broadcast to room
        h.broadcastToRoom(roomID, message)
    }
}

func (h *Hub) broadcastToRoom(roomID string, message Message) {
    if room, exists := h.rooms[roomID]; exists {
        for client := range room.Clients {
            select {
            case client.Send <- message:
            default:
                close(client.Send)
                delete(room.Clients, client)
                delete(h.clients, client)
            }
        }
    }
}

func (h *Hub) handlePrivateMessage(message Message) {
    h.mutex.RLock()
    defer h.mutex.RUnlock()
    
    for client := range h.clients {
        if client.ID == message.To {
            select {
            case client.Send <- message:
            default:
            }
            break
        }
    }
}

func (h *Hub) handleBroadcast(message Message) {
    h.mutex.RLock()
    defer h.mutex.RUnlock()
    
    for client := range h.clients {
        select {
        case client.Send <- message:
        default:
            close(client.Send)
            delete(h.clients, client)
        }
    }
}

func (c *Client) readPump() {
    defer func() {
        c.Hub.unregister <- c
        c.Conn.Close()
    }()
    
    c.Conn.SetReadLimit(512)
    c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.Conn.SetPongHandler(func(string) error {
        c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        var message Message
        err := c.Conn.ReadJSON(&message)
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }
        
        message.From = c.ID
        message.Time = time.Now()
        c.Hub.broadcast <- message
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.Conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.Send:
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            if err := c.Conn.WriteJSON(message); err != nil {
                log.Printf("Write error: %v", err)
                return
            }
            
        case <-ticker.C:
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func main() {
    app := zoox.New()
    
    hub := NewHub()
    go hub.Run()
    
    var upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }
    
    // WebSocket endpoint
    app.Get("/ws", func(ctx *zoox.Context) {
        conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
        if err != nil {
            log.Printf("WebSocket upgrade error: %v", err)
            return
        }
        
        clientID := ctx.Query("id")
        if clientID == "" {
            clientID = generateClientID()
        }
        
        client := &Client{
            ID:       clientID,
            Conn:     conn,
            Hub:      hub,
            Send:     make(chan Message, 256),
            Rooms:    make(map[string]bool),
            UserInfo: make(map[string]interface{}),
        }
        
        hub.register <- client
        
        go client.writePump()
        go client.readPump()
    })
    
    // Chat application
    app.Get("/chat", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>WebSocket Chat</title>
            <style>
                body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
                .chat-container { max-width: 800px; margin: 0 auto; }
                .messages { height: 400px; overflow-y: auto; border: 1px solid #ccc; padding: 10px; margin-bottom: 10px; }
                .message { margin-bottom: 10px; padding: 5px; border-radius: 5px; }
                .message.own { background-color: #dcf8c6; text-align: right; }
                .message.other { background-color: #f1f1f1; }
                .input-area { display: flex; gap: 10px; }
                .input-area input { flex: 1; padding: 10px; }
                .input-area button { padding: 10px 20px; }
                .rooms { margin-bottom: 20px; }
                .room-btn { margin-right: 10px; padding: 5px 10px; }
            </style>
        </head>
        <body>
            <div class="chat-container">
                <h1>WebSocket Chat</h1>
                
                <div class="rooms">
                    <button class="room-btn" onclick="joinRoom('general')">General</button>
                    <button class="room-btn" onclick="joinRoom('tech')">Tech</button>
                    <button class="room-btn" onclick="joinRoom('random')">Random</button>
                </div>
                
                <div id="messages" class="messages"></div>
                
                <div class="input-area">
                    <input type="text" id="messageInput" placeholder="Type a message...">
                    <button onclick="sendMessage()">Send</button>
                </div>
            </div>
            
            <script>
                const clientId = 'user_' + Math.random().toString(36).substr(2, 9);
                const ws = new WebSocket('ws://localhost:8080/ws?id=' + clientId);
                const messages = document.getElementById('messages');
                const input = document.getElementById('messageInput');
                let currentRoom = '';
                
                ws.onopen = function() {
                    console.log('Connected to WebSocket server');
                };
                
                ws.onmessage = function(event) {
                    const message = JSON.parse(event.data);
                    displayMessage(message);
                };
                
                ws.onclose = function() {
                    console.log('Disconnected from WebSocket server');
                };
                
                function displayMessage(message) {
                    const div = document.createElement('div');
                    div.className = 'message ' + (message.from === clientId ? 'own' : 'other');
                    
                    let content = '';
                    switch(message.type) {
                        case 'welcome':
                            content = '✓ ' + message.content.message;
                            break;
                        case 'user_joined':
                            content = '👋 ' + message.content.user_id + ' joined the room';
                            break;
                        case 'user_left':
                            content = '👋 ' + message.content.user_id + ' left the room';
                            break;
                        case 'room_message':
                            content = message.from + ': ' + message.content;
                            break;
                        default:
                            content = JSON.stringify(message.content);
                    }
                    
                    div.innerHTML = content;
                    messages.appendChild(div);
                    messages.scrollTop = messages.scrollHeight;
                }
                
                function joinRoom(roomId) {
                    currentRoom = roomId;
                    const message = {
                        type: 'join_room',
                        room: roomId
                    };
                    ws.send(JSON.stringify(message));
                    
                    // Clear messages
                    messages.innerHTML = '';
                    
                    // Update UI
                    document.querySelectorAll('.room-btn').forEach(btn => {
                        btn.style.backgroundColor = btn.textContent.toLowerCase() === roomId ? '#007bff' : '';
                        btn.style.color = btn.textContent.toLowerCase() === roomId ? 'white' : '';
                    });
                }
                
                function sendMessage() {
                    if (input.value && currentRoom) {
                        const message = {
                            type: 'room_message',
                            content: input.value,
                            room: currentRoom
                        };
                        ws.send(JSON.stringify(message));
                        input.value = '';
                    }
                }
                
                input.addEventListener('keypress', function(e) {
                    if (e.key === 'Enter') {
                        sendMessage();
                    }
                });
                
                // Auto-join general room
                setTimeout(() => joinRoom('general'), 1000);
            </script>
        </body>
        </html>
        `
        ctx.HTML(200, html, nil)
    })
    
    log.Println("WebSocket server starting on :8080")
    log.Println("Chat: http://localhost:8080/chat")
    
    app.Listen(":8080")
}

func generateClientID() string {
    return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
```

## 🎯 Hands-on Exercise

Create a real-time collaborative whiteboard application using WebSockets.

## 📚 Key Takeaways

1. **Connection Management**: Handle client connections and disconnections gracefully
2. **Message Routing**: Implement room-based and private messaging
3. **Real-time Features**: Build interactive applications with instant updates
4. **Scalability**: Design for multiple concurrent connections
5. **Error Handling**: Implement proper error handling and reconnection logic

## 🎯 Next Steps

- Learn [Tutorial 09: JSON-RPC Services](./09-json-rpc-services.md)
- Explore [Tutorial 10: Authentication & Authorization](./10-authentication-authorization.md)
- Study [Tutorial 12: Caching Strategies](./12-caching-strategies.md)

---

**Congratulations!** You've mastered WebSocket development in Zoox and can now build real-time applications! 