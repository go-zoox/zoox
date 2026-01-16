# Best Practices

This document outlines best practices for developing applications with Zoox.

## Project Structure

Organize your project with a clear structure:

```
your-project/
├── main.go
├── go.mod
├── go.sum
├── config/
│   └── config.go
├── handlers/
│   ├── user.go
│   └── product.go
├── middleware/
│   └── auth.go
├── models/
│   └── user.go
└── utils/
    └── helpers.go
```

## Error Handling

Always handle errors properly:

```go
app.Get("/users/:id", func(ctx *zoox.Context) {
    id := ctx.Param("id")
    user, err := getUserByID(id)
    if err != nil {
        ctx.Error(404, "User not found", err)
        return
    }
    ctx.JSON(200, user)
})
```

## Middleware

Use middleware for cross-cutting concerns:

```go
// Authentication middleware
app.Use(func(ctx *zoox.Context) {
    token := ctx.Header("Authorization")
    if token == "" {
        ctx.Error(401, "Unauthorized", nil)
        return
    }
    // Validate token...
    ctx.Next()
})
```

## Configuration

Use environment variables for configuration:

```go
import (
    "os"
    "github.com/go-zoox/zoox"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    app := zoox.New()
    app.Run(":" + port)
}
```

## Logging

Use structured logging:

```go
app.Use(zoox.Logger())
```

## Security

Always use HTTPS in production and implement proper security measures:

```go
app.Use(zoox.Helmet())
app.Use(zoox.CORS())
```

## Testing

Write tests for your handlers:

```go
func TestUserHandler(t *testing.T) {
    app := zoox.New()
    app.Get("/users/:id", getUserHandler)
    
    req := httptest.NewRequest("GET", "/users/1", nil)
    w := httptest.NewRecorder()
    
    app.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
}
```
