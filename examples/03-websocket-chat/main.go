package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-zoox/zoox"
	"github.com/gorilla/websocket"
)

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Room      string    `json:"room"`
}

// Client represents a WebSocket client
type Client struct {
	ID       string
	Username string
	Room     string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
}

// Hub manages WebSocket connections and message broadcasting
type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
	messages   []Message
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		messages:   make([]Message, 0),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
			
		case client := <-h.unregister:
			h.unregisterClient(client)
			
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Add client to main registry
	h.clients[client] = true
	
	// Add client to room
	if h.rooms[client.Room] == nil {
		h.rooms[client.Room] = make(map[*Client]bool)
	}
	h.rooms[client.Room][client] = true
	
	// Send welcome message
	welcomeMsg := Message{
		ID:        fmt.Sprintf("welcome_%d", time.Now().UnixNano()),
		Type:      "system",
		Username:  "System",
		Content:   fmt.Sprintf("%s joined the room", client.Username),
		Timestamp: time.Now(),
		Room:      client.Room,
	}
	
	h.messages = append(h.messages, welcomeMsg)
	
	// Broadcast join notification
	msgData, _ := json.Marshal(welcomeMsg)
	h.broadcastToRoom(client.Room, msgData)
	
	// Send recent messages to new client
	h.sendRecentMessages(client)
	
	// Send updated user list
	h.broadcastUserList(client.Room)
	
	log.Printf("Client %s (%s) connected to room %s", client.Username, client.ID, client.Room)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if _, ok := h.clients[client]; ok {
		// Remove from main registry
		delete(h.clients, client)
		
		// Remove from room
		if roomClients, ok := h.rooms[client.Room]; ok {
			delete(roomClients, client)
			if len(roomClients) == 0 {
				delete(h.rooms, client.Room)
			}
		}
		
		close(client.Send)
		
		// Send leave message
		leaveMsg := Message{
			ID:        fmt.Sprintf("leave_%d", time.Now().UnixNano()),
			Type:      "system",
			Username:  "System",
			Content:   fmt.Sprintf("%s left the room", client.Username),
			Timestamp: time.Now(),
			Room:      client.Room,
		}
		
		h.messages = append(h.messages, leaveMsg)
		
		// Broadcast leave notification
		msgData, _ := json.Marshal(leaveMsg)
		h.broadcastToRoom(client.Room, msgData)
		
		// Send updated user list
		h.broadcastUserList(client.Room)
		
		log.Printf("Client %s (%s) disconnected from room %s", client.Username, client.ID, client.Room)
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}
	
	h.mutex.Lock()
	h.messages = append(h.messages, msg)
	h.mutex.Unlock()
	
	h.broadcastToRoom(msg.Room, message)
}

func (h *Hub) broadcastToRoom(room string, message []byte) {
	h.mutex.RLock()
	roomClients := h.rooms[room]
	h.mutex.RUnlock()
	
	for client := range roomClients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, client)
			delete(roomClients, client)
		}
	}
}

func (h *Hub) sendRecentMessages(client *Client) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	// Send last 50 messages from the room
	roomMessages := make([]Message, 0)
	for _, msg := range h.messages {
		if msg.Room == client.Room {
			roomMessages = append(roomMessages, msg)
		}
	}
	
	// Keep only the last 50 messages
	if len(roomMessages) > 50 {
		roomMessages = roomMessages[len(roomMessages)-50:]
	}
	
	for _, msg := range roomMessages {
		msgData, _ := json.Marshal(msg)
		select {
		case client.Send <- msgData:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) broadcastUserList(room string) {
	h.mutex.RLock()
	roomClients := h.rooms[room]
	h.mutex.RUnlock()
	
	users := make([]string, 0)
	for client := range roomClients {
		users = append(users, client.Username)
	}
	
	userListMsg := Message{
		ID:        fmt.Sprintf("userlist_%d", time.Now().UnixNano()),
		Type:      "userlist",
		Username:  "System",
		Content:   "",
		Timestamp: time.Now(),
		Room:      room,
	}
	
	// Add users list to content as JSON
	usersData, _ := json.Marshal(users)
	userListMsg.Content = string(usersData)
	
	msgData, _ := json.Marshal(userListMsg)
	h.broadcastToRoom(room, msgData)
}

func (h *Hub) getRoomStats() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_clients": len(h.clients),
		"total_rooms":   len(h.rooms),
		"total_messages": len(h.messages),
		"rooms": make(map[string]int),
	}
	
	for room, clients := range h.rooms {
		stats["rooms"].(map[string]int)[room] = len(clients)
	}
	
	return stats
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

// readPump handles reading from the WebSocket connection
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
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Set message metadata
		msg.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
		msg.Username = c.Username
		msg.Timestamp = time.Now()
		msg.Room = c.Room
		
		if msg.Type == "" {
			msg.Type = "message"
		}
		
		msgData, _ := json.Marshal(msg)
		c.Hub.broadcast <- msgData
	}
}

// writePump handles writing to the WebSocket connection
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
			
			c.Conn.WriteMessage(websocket.TextMessage, message)
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func main() {
	app := zoox.Default()
	
	// Create and start WebSocket hub
	hub := NewHub()
	go hub.Run()
	
	// Serve the chat interface
	app.Get("/", func(ctx *zoox.Context) {
		ctx.HTML(http.StatusOK, chatHTML)
	})
	
	// WebSocket endpoint
	app.Get("/ws", func(ctx *zoox.Context) {
		username := ctx.Query().Get("username", "Anonymous")
		room := ctx.Query().Get("room", "general")
		
		if username == "" {
			ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Username is required",
			})
			return
		}
		
		// Upgrade connection to WebSocket
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		
		// Create new client
		client := &Client{
			ID:       fmt.Sprintf("client_%d", time.Now().UnixNano()),
			Username: username,
			Room:     room,
			Conn:     conn,
			Hub:      hub,
			Send:     make(chan []byte, 256),
		}
		
		// Register client
		client.Hub.register <- client
		
		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump()
	})
	
	// API endpoints
	api := app.Group("/api")
	
	// Get chat statistics
	api.Get("/stats", func(ctx *zoox.Context) {
		stats := hub.getRoomStats()
		ctx.JSON(http.StatusOK, stats)
	})
	
	// Get room list
	api.Get("/rooms", func(ctx *zoox.Context) {
		hub.mutex.RLock()
		rooms := make([]map[string]interface{}, 0)
		for room, clients := range hub.rooms {
			users := make([]string, 0)
			for client := range clients {
				users = append(users, client.Username)
			}
			rooms = append(rooms, map[string]interface{}{
				"name":       room,
				"user_count": len(clients),
				"users":      users,
			})
		}
		hub.mutex.RUnlock()
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"rooms": rooms,
		})
	})
	
	// Get messages for a room
	api.Get("/rooms/:room/messages", func(ctx *zoox.Context) {
		room := ctx.Param("room")
		
		hub.mutex.RLock()
		roomMessages := make([]Message, 0)
		for _, msg := range hub.messages {
			if msg.Room == room {
				roomMessages = append(roomMessages, msg)
			}
		}
		hub.mutex.RUnlock()
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"room":     room,
			"messages": roomMessages,
		})
	})
	
	fmt.Println("🚀 WebSocket Chat starting...")
	fmt.Println("📍 Server running on http://localhost:8080")
	fmt.Println("💬 Chat Interface: http://localhost:8080")
	fmt.Println("📊 Chat Statistics: http://localhost:8080/api/stats")
	fmt.Println("🏠 Room List: http://localhost:8080/api/rooms")
	
	app.Run(":8080")
}

// HTML template for the chat interface
const chatHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Zoox WebSocket Chat</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: Arial, sans-serif; background: #f0f0f0; }
        .container { max-width: 800px; margin: 0 auto; background: white; min-height: 100vh; display: flex; flex-direction: column; }
        .header { background: #333; color: white; padding: 1rem; text-align: center; }
        .chat-area { flex: 1; display: flex; }
        .users { width: 200px; background: #f8f8f8; padding: 1rem; border-right: 1px solid #ddd; }
        .messages-container { flex: 1; display: flex; flex-direction: column; }
        .messages { flex: 1; padding: 1rem; max-height: 500px; overflow-y: auto; }
        .message { margin: 0.5rem 0; padding: 0.5rem; border-radius: 4px; }
        .message.user { background: #e3f2fd; }
        .message.system { background: #fff3e0; font-style: italic; }
        .message .username { font-weight: bold; color: #1976d2; }
        .message .timestamp { font-size: 0.8em; color: #666; }
        .input-area { padding: 1rem; border-top: 1px solid #ddd; }
        .login-form, .message-form { display: flex; gap: 0.5rem; }
        input[type="text"] { flex: 1; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }
        button { padding: 0.5rem 1rem; background: #1976d2; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #1565c0; }
        .status { padding: 0.5rem; text-align: center; background: #e8f5e8; }
        .users h3 { margin-bottom: 0.5rem; }
        .user-list { list-style: none; }
        .user-list li { padding: 0.25rem 0; }
        #loginArea { padding: 2rem; text-align: center; }
        #chatArea { display: none; }
        .room-info { background: #f5f5f5; padding: 0.5rem; text-align: center; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Zoox WebSocket Chat</h1>
        </div>
        
        <div id="loginArea">
            <h2>Join Chat</h2>
            <form class="login-form" onsubmit="connect(event)">
                <input type="text" id="usernameInput" placeholder="Enter your username" required>
                <input type="text" id="roomInput" placeholder="Room (default: general)" value="general">
                <button type="submit">Connect</button>
            </form>
        </div>
        
        <div id="chatArea">
            <div class="room-info">
                Room: <span id="currentRoom"></span> | Status: <span id="connectionStatus">Disconnected</span>
            </div>
            
            <div class="chat-area">
                <div class="users">
                    <h3>Online Users</h3>
                    <ul id="userList" class="user-list"></ul>
                </div>
                
                <div class="messages-container">
                    <div id="messages" class="messages"></div>
                    
                    <div class="input-area">
                        <form class="message-form" onsubmit="sendMessage(event)">
                            <input type="text" id="messageInput" placeholder="Type your message..." required>
                            <button type="submit">Send</button>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        let ws = null;
        let username = '';
        let room = '';
        
        function connect(event) {
            event.preventDefault();
            
            username = document.getElementById('usernameInput').value.trim();
            room = document.getElementById('roomInput').value.trim() || 'general';
            
            if (!username) {
                alert('Please enter a username');
                return;
            }
            
            const wsUrl = 'ws://localhost:8080/ws?username=' + encodeURIComponent(username) + '&room=' + encodeURIComponent(room);
            
            try {
                ws = new WebSocket(wsUrl);
                
                ws.onopen = function() {
                    document.getElementById('loginArea').style.display = 'none';
                    document.getElementById('chatArea').style.display = 'flex';
                    document.getElementById('currentRoom').textContent = room;
                    document.getElementById('connectionStatus').textContent = 'Connected';
                    document.getElementById('connectionStatus').style.color = 'green';
                };
                
                ws.onmessage = function(event) {
                    const message = JSON.parse(event.data);
                    handleMessage(message);
                };
                
                ws.onclose = function() {
                    document.getElementById('connectionStatus').textContent = 'Disconnected';
                    document.getElementById('connectionStatus').style.color = 'red';
                };
                
                ws.onerror = function(error) {
                    console.error('WebSocket error:', error);
                    alert('Connection error. Please try again.');
                };
                
            } catch (error) {
                console.error('WebSocket connection failed:', error);
                alert('Failed to connect. Please try again.');
            }
        }
        
        function sendMessage(event) {
            event.preventDefault();
            
            const input = document.getElementById('messageInput');
            const content = input.value.trim();
            
            if (!content || !ws) return;
            
            const message = {
                type: 'message',
                content: content
            };
            
            ws.send(JSON.stringify(message));
            input.value = '';
        }
        
        function handleMessage(message) {
            if (message.type === 'userlist') {
                updateUserList(JSON.parse(message.content));
            } else {
                displayMessage(message);
            }
        }
        
        function displayMessage(message) {
            const messagesDiv = document.getElementById('messages');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message ' + message.type;
            
            const timestamp = new Date(message.timestamp).toLocaleTimeString();
            
            messageDiv.innerHTML = 
                '<div class="username">' + escapeHtml(message.username) + '</div>' +
                '<div>' + escapeHtml(message.content) + '</div>' +
                '<div class="timestamp">' + timestamp + '</div>';
            
            messagesDiv.appendChild(messageDiv);
            messagesDiv.scrollTop = messagesDiv.scrollHeight;
        }
        
        function updateUserList(users) {
            const userList = document.getElementById('userList');
            userList.innerHTML = '';
            
            users.forEach(function(user) {
                const li = document.createElement('li');
                li.textContent = user;
                if (user === username) {
                    li.style.fontWeight = 'bold';
                }
                userList.appendChild(li);
            });
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        // Handle page unload
        window.addEventListener('beforeunload', function() {
            if (ws) {
                ws.close();
            }
        });
    </script>
</body>
</html>` 