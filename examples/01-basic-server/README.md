# Basic Server Example

A comprehensive REST API demonstrating fundamental Zoox framework concepts including routing, middleware, JSON handling, and basic authentication.

## Features

- ✅ **RESTful API Design** - Complete CRUD operations for user management
- ✅ **Route Groups** - Organized public and protected endpoints
- ✅ **Middleware** - Authentication and logging middleware
- ✅ **JSON Handling** - Request/response JSON serialization
- ✅ **Error Handling** - Structured error responses
- ✅ **Basic Authentication** - Token-based authentication demo
- ✅ **API Documentation** - Self-documenting endpoints

## What You'll Learn

- Setting up a basic Zoox server
- Creating REST endpoints with proper HTTP methods
- Working with JSON requests and responses
- Using route groups for API organization
- Implementing basic authentication middleware
- Handling errors gracefully
- Adding logging and monitoring

## Quick Start

1. **Navigate to this example:**
   ```bash
   cd examples/01-basic-server
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the server:**
   ```bash
   go run main.go
   ```

4. **Test the API:**
   ```bash
   # Check server health
   curl http://localhost:8080/health
   
   # Get all users (public)
   curl http://localhost:8080/api/v1/users
   
   # Get user by ID (public)
   curl http://localhost:8080/api/v1/users/1
   ```

## API Endpoints

### Public Endpoints

#### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": 1640995200,
  "version": "1.0.0"
}
```

#### Get All Users
```http
GET /api/v1/users
```

**Response:**
```json
{
  "users": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-01T12:00:00Z"
    }
  ],
  "count": 1
}
```

#### Get User by ID
```http
GET /api/v1/users/:id
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

### Protected Endpoints

All protected endpoints require the `Authorization` header:
```
Authorization: Bearer demo-token
```

#### Create User
```http
POST /api/v1/users
Content-Type: application/json
Authorization: Bearer demo-token

{
  "name": "Jane Smith",
  "email": "jane@example.com"
}
```

**Response:**
```json
{
  "message": "User created successfully",
  "user": {
    "id": 3,
    "name": "Jane Smith",
    "email": "jane@example.com",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

#### Update User
```http
PUT /api/v1/users/:id
Content-Type: application/json
Authorization: Bearer demo-token

{
  "name": "Jane Doe",
  "email": "jane.doe@example.com"
}
```

#### Delete User
```http
DELETE /api/v1/users/:id
Authorization: Bearer demo-token
```

### API Documentation
```http
GET /api/docs
```

Returns a complete API reference in JSON format.

## Testing with cURL

### Basic Tests
```bash
# Health check
curl -i http://localhost:8080/health

# Get all users
curl -i http://localhost:8080/api/v1/users

# Get specific user
curl -i http://localhost:8080/api/v1/users/1

# API documentation
curl -i http://localhost:8080/api/docs
```

### Protected Endpoint Tests
```bash
# Create a new user
curl -i -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -d '{"name":"Test User","email":"test@example.com"}'

# Update user
curl -i -X PUT http://localhost:8080/api/v1/users/3 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -d '{"name":"Updated User","email":"updated@example.com"}'

# Delete user
curl -i -X DELETE http://localhost:8080/api/v1/users/3 \
  -H "Authorization: Bearer demo-token"
```

### Error Scenarios
```bash
# Missing authentication
curl -i -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com"}'

# Invalid user ID
curl -i http://localhost:8080/api/v1/users/invalid

# Missing required fields
curl -i -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -d '{"name":"","email":""}'
```

## Code Structure

### Main Components

1. **User Model** - Represents user data structure
2. **UserStore** - In-memory storage with thread-safe operations
3. **Middleware** - Authentication and logging
4. **Route Handlers** - CRUD operation endpoints
5. **Error Handling** - Structured error responses

### Key Concepts Demonstrated

#### Route Groups
```go
// Public API routes
public := app.Group("/api/v1")
public.Get("/users", getUsersHandler)

// Protected API routes (require authentication)
protected := app.Group("/api/v1", authMiddleware)
protected.Post("/users", createUserHandler)
```

#### Middleware Usage
```go
// Custom authentication middleware
authMiddleware := func(ctx *zoox.Context) {
    auth := ctx.Header("Authorization")
    if auth != "Bearer demo-token" {
        ctx.JSON(http.StatusUnauthorized, ErrorResponse{...})
        return
    }
    ctx.Next()
}
```

#### JSON Binding
```go
var req CreateUserRequest
if err := ctx.BindJSON(&req); err != nil {
    // Handle error
}
```

#### Error Responses
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}
```

## Configuration

### Server Settings
- **Port:** 8080 (configurable in `app.Run(":8080")`)
- **Authentication:** Bearer token (`demo-token`)
- **Data Storage:** In-memory (not persistent)

### Sample Data
The server starts with two sample users:
1. John Doe (john@example.com)
2. Jane Smith (jane@example.com)

## Security Notes

⚠️ **Important:** This example uses hardcoded authentication for demonstration purposes only. In production:

- Use proper JWT tokens or OAuth
- Implement password hashing
- Use secure token storage
- Add rate limiting
- Validate all inputs
- Use HTTPS

## Next Steps

After mastering this example, try:

1. **[02-middleware-showcase](../02-middleware-showcase/)** - Learn about Zoox's built-in middleware
2. **[03-websocket-chat](../03-websocket-chat/)** - Add real-time features
3. **[06-production-api](../06-production-api/)** - See production-ready patterns

## Troubleshooting

### Common Issues

**Port already in use:**
```bash
# Kill existing process
sudo lsof -ti:8080 | xargs kill -9

# Or use a different port
go run main.go -port=8081
```

**Authentication errors:**
```bash
# Ensure you're using the correct token
curl -H "Authorization: Bearer demo-token" ...
```

**JSON parsing errors:**
```bash
# Ensure Content-Type header is set
curl -H "Content-Type: application/json" ...
```

## Contributing

Found an issue or want to improve this example? Please:

1. Check existing issues
2. Create a detailed bug report
3. Submit a pull request with improvements

---

📚 **Learn More:** Check out the [main documentation](../../DOCUMENTATION.md) for detailed API reference and advanced features. 