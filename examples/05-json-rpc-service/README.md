# JSON-RPC Service Example

This example demonstrates how to implement a JSON-RPC 2.0 service using the Zoox framework. It showcases service-oriented architecture, method registration, error handling, and provides both programmatic and interactive testing interfaces.

## Features

### JSON-RPC 2.0 Compliance
- **Standard protocol** implementation following JSON-RPC 2.0 specification
- **Request/Response handling** with proper ID correlation
- **Batch request support** for multiple operations
- **Error codes and messages** following RPC standards

### Service Architecture
- **Modular service design** with separate service classes
- **Method registration** with automatic discovery
- **Parameter validation** and type checking
- **Result serialization** with proper JSON formatting

### Available Services
- **Math Service** - Basic mathematical operations
- **User Service** - User management operations  
- **System Service** - System information and utilities
- **Echo Service** - Testing and debugging utilities

### Advanced Features
- **Custom error handling** with detailed error information
- **Method introspection** and service discovery
- **Interactive testing interface** with HTML/JavaScript client
- **Logging and monitoring** for RPC calls

## Quick Start

1. **Run the JSON-RPC server:**
   ```bash
   cd examples/05-json-rpc-service
   go mod tidy
   go run main.go
   ```

2. **Test with the interactive interface:**
   - Open browser to `http://localhost:8080`
   - Use the web interface to test RPC methods
   - Try different services and parameters

3. **Test with curl:**
   ```bash
   # Basic math operation
   curl -X POST http://localhost:8080/rpc \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "method": "Math.Add",
       "params": [5, 3],
       "id": 1
     }'
   
   # Create a user
   curl -X POST http://localhost:8080/rpc \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0", 
       "method": "User.Create",
       "params": {"name": "John", "email": "john@example.com"},
       "id": 2
     }'
   ```

## Available RPC Methods

### Math Service (`Math.*`)

**Math.Add** - Add two numbers
```json
{
  "jsonrpc": "2.0",
  "method": "Math.Add", 
  "params": [10, 5],
  "id": 1
}
// Response: {"jsonrpc": "2.0", "result": 15, "id": 1}
```

**Math.Subtract** - Subtract two numbers
```json
{
  "jsonrpc": "2.0",
  "method": "Math.Subtract",
  "params": [10, 3],
  "id": 2  
}
// Response: {"jsonrpc": "2.0", "result": 7, "id": 2}
```

**Math.Multiply** - Multiply two numbers
```json
{
  "jsonrpc": "2.0",
  "method": "Math.Multiply",
  "params": [4, 6],
  "id": 3
}
```

**Math.Divide** - Divide two numbers (with error handling)
```json
{
  "jsonrpc": "2.0",
  "method": "Math.Divide", 
  "params": [10, 2],
  "id": 4
}
// Error case (division by zero):
// {"jsonrpc": "2.0", "error": {"code": -32602, "message": "Division by zero"}, "id": 4}
```

**Math.Power** - Calculate power
```json
{
  "jsonrpc": "2.0",
  "method": "Math.Power",
  "params": [2, 8],
  "id": 5
}
```

### User Service (`User.*`)

**User.Create** - Create a new user
```json
{
  "jsonrpc": "2.0",
  "method": "User.Create",
  "params": {
    "name": "Alice Smith",
    "email": "alice@example.com",
    "age": 30
  },
  "id": 10
}
```

**User.Get** - Get user by ID
```json
{
  "jsonrpc": "2.0",
  "method": "User.Get", 
  "params": [1],
  "id": 11
}
```

**User.List** - List all users
```json
{
  "jsonrpc": "2.0",
  "method": "User.List",
  "params": [],
  "id": 12
}
```

**User.Update** - Update user information
```json
{
  "jsonrpc": "2.0",
  "method": "User.Update",
  "params": {
    "id": 1,
    "name": "Alice Johnson", 
    "email": "alice.johnson@example.com"
  },
  "id": 13
}
```

**User.Delete** - Delete user by ID
```json
{
  "jsonrpc": "2.0",
  "method": "User.Delete",
  "params": [1],
  "id": 14
}
```

### System Service (`System.*`)

**System.Info** - Get system information
```json
{
  "jsonrpc": "2.0",
  "method": "System.Info",
  "params": [],
  "id": 20
}
```

**System.Time** - Get current server time
```json
{
  "jsonrpc": "2.0",
  "method": "System.Time",
  "params": [],
  "id": 21
}
```

**System.Methods** - List all available methods
```json
{
  "jsonrpc": "2.0", 
  "method": "System.Methods",
  "params": [],
  "id": 22
}
```

### Echo Service (`Echo.*`)

**Echo.Message** - Echo back a message
```json
{
  "jsonrpc": "2.0",
  "method": "Echo.Message",
  "params": ["Hello, JSON-RPC!"],
  "id": 30
}
```

**Echo.Delay** - Echo with artificial delay (for testing)
```json
{
  "jsonrpc": "2.0",
  "method": "Echo.Delay", 
  "params": {"message": "Delayed response", "seconds": 2},
  "id": 31
}
```

## Batch Requests

Send multiple RPC calls in a single HTTP request:

```json
[
  {
    "jsonrpc": "2.0",
    "method": "Math.Add",
    "params": [1, 2],
    "id": 1
  },
  {
    "jsonrpc": "2.0", 
    "method": "Math.Multiply",
    "params": [3, 4],
    "id": 2
  },
  {
    "jsonrpc": "2.0",
    "method": "User.List", 
    "params": [],
    "id": 3
  }
]
```

## Error Handling

### Standard JSON-RPC Error Codes
- **-32700** Parse error (Invalid JSON)
- **-32600** Invalid Request
- **-32601** Method not found
- **-32602** Invalid params  
- **-32603** Internal error
- **-32000 to -32099** Server error (custom)

### Custom Error Examples

**Method not found:**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": "Method 'Math.Invalid' does not exist"
  },
  "id": 1
}
```

**Invalid parameters:**
```json
{
  "jsonrpc": "2.0", 
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Expected 2 parameters, got 1"
  },
  "id": 2
}
```

**Business logic error:**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "User not found", 
    "data": "No user with ID 999"
  },
  "id": 3
}
```

## Service Implementation

### Service Structure
```go
type MathService struct{}

func (s *MathService) Add(a, b float64) float64 {
    return a + b
}

func (s *MathService) Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

### Method Registration
```go
// Register services
rpc := NewRPCHandler()
rpc.RegisterService("Math", &MathService{})
rpc.RegisterService("User", &UserService{})
rpc.RegisterService("System", &SystemService{})
rpc.RegisterService("Echo", &EchoService{})
```

### Parameter Handling
```go
type CreateUserParams struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=0,max=150"`
}

func (s *UserService) Create(params CreateUserParams) (*User, error) {
    // Validation is automatic
    user := &User{
        ID:    generateID(),
        Name:  params.Name,
        Email: params.Email,
        Age:   params.Age,
    }
    return user, s.store.Save(user)
}
```

## Testing Strategies

### 1. Unit Testing RPC Methods
```go
func TestMathService_Add(t *testing.T) {
    service := &MathService{}
    result := service.Add(2, 3)
    assert.Equal(t, 5.0, result)
}

func TestMathService_Divide_ByZero(t *testing.T) {
    service := &MathService{}
    _, err := service.Divide(10, 0)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "division by zero")
}
```

### 2. Integration Testing
```go
func TestRPCHandler_MathAdd(t *testing.T) {
    handler := NewRPCHandler()
    handler.RegisterService("Math", &MathService{})
    
    request := RPCRequest{
        JSONRPC: "2.0",
        Method:  "Math.Add",
        Params:  []interface{}{5.0, 3.0},
        ID:      1,
    }
    
    response := handler.HandleRequest(request)
    assert.Equal(t, 8.0, response.Result)
}
```

### 3. Load Testing
```bash
# Using Apache Bench
ab -n 1000 -c 10 -p request.json -T application/json \
  http://localhost:8080/rpc

# Using curl in a loop
for i in {1..100}; do
  curl -X POST http://localhost:8080/rpc \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"Math.Add","params":[1,2],"id":'$i'}'
done
```

### 4. Error Testing
```bash
# Test invalid JSON
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"invalid": json}'

# Test method not found
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"Invalid.Method","id":1}'

# Test invalid parameters
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"Math.Add","params":["not","numbers"],"id":1}'
```

## Performance Considerations

### Request Processing
- **Single-threaded** method execution per request
- **Concurrent requests** handled by multiple goroutines
- **Memory pooling** for JSON marshaling/unmarshaling
- **Connection reuse** for HTTP keep-alive

### Optimization Strategies
- **Method caching** for reflection-based method lookup
- **Parameter pre-validation** to fail fast
- **Response compression** for large result sets
- **Connection pooling** for database operations

### Monitoring Metrics
- **Request count** per method
- **Response times** and percentiles
- **Error rates** by error type
- **Concurrent connections** and queue depth

## Learning Objectives

After working with this example, you will understand:

1. **JSON-RPC Protocol**
   - Request/response format and structure
   - Error handling and standard error codes
   - Batch processing and notifications

2. **Service-Oriented Architecture**
   - Service registration and discovery
   - Method routing and parameter binding
   - Interface design and contracts

3. **Go Reflection and Type System**
   - Dynamic method invocation
   - Parameter type conversion
   - Error handling and propagation

4. **API Design Patterns**
   - RPC vs REST trade-offs
   - Versioning strategies
   - Documentation and discoverability

## Production Considerations

### 1. Authentication & Authorization
```go
type AuthContext struct {
    UserID string `json:"user_id"`
    Roles  []string `json:"roles"`
}

func (s *UserService) GetProfile(ctx AuthContext, userID string) (*User, error) {
    if ctx.UserID != userID && !hasRole(ctx.Roles, "admin") {
        return nil, errors.New("access denied")
    }
    return s.store.GetUser(userID)
}
```

### 2. Rate Limiting
```go
func (h *RPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if !h.rateLimiter.Allow(r.RemoteAddr) {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    // Process request...
}
```

### 3. Logging and Monitoring
```go
func (h *RPCHandler) logRequest(req *RPCRequest, resp *RPCResponse, duration time.Duration) {
    log.WithFields(log.Fields{
        "method":   req.Method,
        "id":       req.ID,
        "duration": duration,
        "error":    resp.Error != nil,
    }).Info("RPC call completed")
}
```

## Next Steps

- Explore the **Production API** example for authentication patterns  
- Check the **WebSocket Chat** example for real-time RPC over WebSockets
- Review the **Middleware Showcase** for security and monitoring middleware
- Consider gRPC as an alternative to JSON-RPC for high-performance scenarios 