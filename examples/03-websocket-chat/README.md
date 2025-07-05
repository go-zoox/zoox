# WebSocket Chat Example

This example demonstrates real-time bidirectional communication using WebSockets with the Zoox framework. It implements a complete chat application with user management, message broadcasting, and connection handling.

## Features

### Real-Time Communication
- **WebSocket connections** with automatic reconnection
- **Message broadcasting** to all connected clients
- **User presence** tracking (join/leave notifications)
- **Connection management** with graceful cleanup

### Chat Functionality
- **Public chat room** for all users
- **User nicknames** with validation
- **Message history** for new connections
- **Typing indicators** (optional feature)
- **User list** with online status

### Technical Features
- **Concurrent connection handling** using goroutines
- **Thread-safe message broadcasting** with mutexes
- **JSON message protocol** for structured communication
- **Error handling** for connection failures
- **Graceful shutdown** with proper cleanup

## Quick Start

1. **Run the chat server:**
   ```bash
   cd examples/03-websocket-chat
   go mod tidy
   go run main.go
   ```

2. **Open the chat interface:**
   - Open your browser to `http://localhost:8080`
   - Enter a nickname to join the chat
   - Start chatting with other users!

3. **Test with multiple clients:**
   - Open multiple browser tabs/windows
   - Use different nicknames
   - See real-time message synchronization

## API Endpoints

### WebSocket Connection
- **Endpoint:** `ws://localhost:8080/ws`
- **Protocol:** JSON-based message protocol
- **Heartbeat:** Automatic ping/pong for connection health

### HTTP Endpoints
- **GET /** - Chat web interface
- **GET /health** - Health check endpoint
- **GET /stats** - Connection statistics

## Message Protocol

### Client to Server Messages

**Join Chat:**
```json
{
  "type": "join",
  "nickname": "john_doe",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Send Message:**
```json
{
  "type": "message",
  "content": "Hello everyone!",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Leave Chat:**
```json
{
  "type": "leave",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Server to Client Messages

**User Joined:**
```json
{
  "type": "user_joined",
  "nickname": "john_doe",
  "timestamp": "2024-01-01T12:00:00Z",
  "user_count": 5
}
```

**Chat Message:**
```json
{
  "type": "message",
  "nickname": "john_doe",
  "content": "Hello everyone!",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**User Left:**
```json
{
  "type": "user_left",
  "nickname": "john_doe", 
  "timestamp": "2024-01-01T12:00:00Z",
  "user_count": 4
}
```

**System Message:**
```json
{
  "type": "system",
  "content": "Welcome to the chat!",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Architecture

### Connection Management
```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    mutex      sync.RWMutex
}
```

### Client Structure  
```go
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte
    nickname string
    joinTime time.Time
}
```

### Message Flow
1. **Client connects** → WebSocket upgrade → Register with hub
2. **Client sends message** → JSON parsing → Broadcast to all clients
3. **Client disconnects** → Cleanup → Notify other clients

## Testing the Chat

### Basic Functionality
1. **Single User Test:**
   ```bash
   # Open browser to localhost:8080
   # Enter nickname "Alice"
   # Send message "Hello world"
   # Verify message appears
   ```

2. **Multi-User Test:**
   ```bash
   # Open 3 browser tabs
   # Join as "Alice", "Bob", "Charlie"
   # Send messages from each user
   # Verify all users see all messages
   ```

### WebSocket Testing with Tools

**Using wscat (if installed):**
```bash
# Install wscat: npm install -g wscat
wscat -c ws://localhost:8080/ws

# Send join message
{"type":"join","nickname":"test_user"}

# Send chat message  
{"type":"message","content":"Hello from wscat!"}
```

**Using curl for HTTP endpoints:**
```bash
curl http://localhost:8080/health
curl http://localhost:8080/stats
```

## Features Demonstrated

### 1. WebSocket Basics
- **Connection upgrade** from HTTP to WebSocket
- **Bidirectional communication** between client and server
- **Message handling** with JSON protocol

### 2. Concurrency Patterns
- **Goroutine per connection** for concurrent client handling
- **Channel communication** between goroutines
- **Mutex protection** for shared data structures

### 3. Real-Time Features
- **Instant message delivery** to all connected clients
- **User presence notifications** (join/leave events)
- **Connection state management** and cleanup

### 4. Error Handling
- **Connection failure recovery** with automatic cleanup
- **Invalid message handling** with error responses
- **Graceful degradation** when clients disconnect unexpectedly

## Learning Objectives

After exploring this example, you will understand:

1. **WebSocket Protocol**
   - How to upgrade HTTP connections to WebSocket
   - Message framing and protocol design
   - Connection lifecycle management

2. **Real-Time Communication Patterns**
   - Message broadcasting architectures
   - Client state synchronization
   - Event-driven programming

3. **Go Concurrency**
   - Goroutines for concurrent connections
   - Channel communication patterns
   - Thread-safe data access with mutexes

4. **Production Considerations**
   - Connection scaling strategies
   - Memory management for long-lived connections
   - Error handling and recovery patterns

## Extending the Example

### Add Private Messaging
```go
// Add room/channel support
type Room struct {
    name    string
    clients map[*Client]bool
}
```

### Add Message Persistence
```go
// Store messages in database
type Message struct {
    ID        int       `json:"id"`
    Nickname  string    `json:"nickname"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

### Add Authentication
```go
// Require JWT token for WebSocket connections
func authenticateWebSocket(r *http.Request) (*User, error) {
    token := r.Header.Get("Authorization")
    return validateJWT(token)
}
```

## Performance Considerations

### Connection Limits
- **Default limit:** 1000 concurrent connections
- **Memory usage:** ~8KB per connection
- **CPU usage:** Minimal when idle

### Scaling Strategies
- **Horizontal scaling** with Redis pub/sub
- **Load balancing** with sticky sessions
- **Connection pooling** for database operations

## Troubleshooting

**WebSocket connection fails:**
```bash
# Check if server is running
curl http://localhost:8080/health

# Check browser console for errors
# Verify WebSocket URL is correct
```

**Messages not appearing:**
```bash
# Check browser network tab
# Verify JSON message format
# Check server logs for errors
```

**High memory usage:**
```bash
# Monitor with: go tool pprof http://localhost:8080/debug/pprof/heap
# Check for connection leaks
# Verify proper cleanup on disconnect
```

## Next Steps

- Explore the **Production API** example for authentication patterns
- Check the **Middleware Showcase** for security and monitoring
- Review the **Basic Server** example for REST API fundamentals 