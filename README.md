# Zoox - A Lightweight, High Performance Go Web Framework

[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-zoox/zoox)](https://pkg.go.dev/github.com/go-zoox/zoox)
[![Build Status](https://github.com/go-zoox/zoox/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/go-zoox/zoox/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-zoox/zoox)](https://goreportcard.com/report/github.com/go-zoox/zoox)
[![Coverage Status](https://coveralls.io/repos/github/go-zoox/zoox/badge.svg?branch=master)](https://coveralls.io/github/go-zoox/zoox?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/go-zoox/zoox.svg)](https://github.com/go-zoox/zoox/issues)
[![Release](https://img.shields.io/github/tag/go-zoox/zoox.svg?label=Release)](https://github.com/go-zoox/zoox/tags)
[![Go Version](https://img.shields.io/badge/Go-1.22.1+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)

Zoox is a modern, lightweight, and high-performance web framework for Go. It provides a comprehensive set of features for building robust web applications, APIs, and microservices with excellent developer experience.

## ✨ Features

### 🚀 Core Features
- **Lightweight & Fast**: Minimal overhead with high performance
- **Type-Safe**: Full Go type safety with excellent IDE support
- **Middleware Support**: Rich ecosystem of built-in middleware
- **Router**: Fast trie-based router with parameter support
- **Context**: Enhanced HTTP context with utilities
- **Templates**: Built-in template engine with custom functions

### 🔧 Built-in Components
- **Cache**: Redis and in-memory caching support
- **Session**: Secure session management
- **JWT**: JSON Web Token authentication
- **CORS**: Cross-Origin Resource Sharing
- **Rate Limiting**: Request rate limiting
- **WebSocket**: Real-time communication
- **JSON-RPC**: JSON-RPC server support
- **Pub/Sub**: Event-driven messaging
- **Message Queue**: Asynchronous message processing
- **Cron Jobs**: Scheduled task execution
- **Job Queue**: Background job processing
- **i18n**: Internationalization support
- **Logger**: Structured logging
- **Monitoring**: Prometheus metrics
- **Debug**: Development debugging tools

### 🛡️ Security & Performance
- **Helmet**: Security headers middleware
- **Gzip**: Response compression
- **Body Limit**: Request size limiting
- **Timeout**: Request timeout handling
- **Recovery**: Panic recovery
- **Real IP**: Client IP detection
- **Request ID**: Request tracing
- **Sentry**: Error tracking integration

## 📦 Installation

```bash
go get github.com/go-zoox/zoox
```

## 🚀 Quick Start

### Basic Example

```go
package main

import "github.com/go-zoox/zoox"

func main() {
    app := zoox.Default()

    app.Get("/", func(ctx *zoox.Context) {
        ctx.JSON(zoox.H{
            "message": "Hello, Zoox!",
            "version": zoox.Version,
        })
    })

    app.Get("/users/:id", func(ctx *zoox.Context) {
        id := ctx.Param("id")
        ctx.JSON(zoox.H{
            "id":   id,
            "name": "John Doe",
        })
    })

    app.Run(":8080")
}
```

### Advanced Example with Middleware

```go
package main

import (
    "github.com/go-zoox/zoox"
    "github.com/go-zoox/zoox/middleware"
)

func main() {
    app := zoox.New()

    // Global middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    app.Use(middleware.CORS())
    app.Use(middleware.Gzip())

    // API routes
    api := app.Group("/api/v1")
    api.Use(middleware.Jwt())

    api.Get("/users", func(ctx *zoox.Context) {
        ctx.JSON(zoox.H{
            "users": []zoox.H{
                {"id": 1, "name": "Alice"},
                {"id": 2, "name": "Bob"},
            },
        })
    })

    api.Post("/users", func(ctx *zoox.Context) {
        var user struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }
        
        if err := ctx.BindJSON(&user); err != nil {
            ctx.Error(400, "Invalid JSON")
            return
        }

        ctx.JSON(zoox.H{
            "message": "User created",
            "user":    user,
        })
    })

    app.Run(":8080")
}
```

## 🛠️ Development Tools

Install the Zoox CLI for enhanced development experience:

```bash
go install github.com/go-zoox/zoox/cmd/zoox@latest
```

### CLI Commands

```bash
# Start development server with hot reload
zoox dev

# Build application for production
zoox build

# Run tests
zoox test

# Generate API documentation
zoox docs
```

## 📚 Documentation

📖 **Full documentation is available at:** [https://go-zoox.github.io/zoox/](https://go-zoox.github.io/zoox/)

The documentation includes:
- 🚀 [Getting Started](https://go-zoox.github.io/zoox/getting-started/installation) - Installation and quick start guide
- 📖 [Guides](https://go-zoox.github.io/zoox/guides/routing) - Routing, middleware, context, templates, and configuration
- 🔧 [Components](https://go-zoox.github.io/zoox/components/cache) - Cache, session, JWT, and more
- 🛡️ [Middleware](https://go-zoox.github.io/zoox/middleware/overview) - Authentication, security, performance, and monitoring
- 🚀 [Advanced](https://go-zoox.github.io/zoox/advanced/websocket) - WebSocket, JSON-RPC, proxy, cron jobs
- 📚 [API Reference](https://go-zoox.github.io/zoox/api-reference/application) - Complete API documentation
- 💡 [Examples](https://go-zoox.github.io/zoox/examples/rest-api) - RESTful API, WebSocket, Static Files, JSON-RPC, API Gateway, and microservice examples
- 🎯 [Best Practices](https://go-zoox.github.io/zoox/best-practices) - Development guidelines and recommendations

### Middleware

Zoox provides a rich set of middleware for common web application needs:

```go
// Authentication
app.Use(middleware.Jwt())
app.Use(middleware.BasicAuth("Protected Area", map[string]string{
    "admin": "password",
}))
app.Use(middleware.BearerToken("token"))

// Security
app.Use(middleware.Helmet(nil))
app.Use(middleware.CORS())
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
    Period: time.Minute,
    Limit:  100,
}))

// Performance
app.Use(middleware.Gzip())
app.Use(middleware.CacheControl(&middleware.CacheControlConfig{
    Paths:  []string{".*"},
    MaxAge: time.Hour,
}))

// Monitoring
app.Use(middleware.Prometheus())
app.Use(middleware.Logger())
app.Use(middleware.RequestID())

// Development
app.Use(middleware.PProf())
```

### Context Utilities

```go
func handler(ctx *zoox.Context) {
    // Request data
    body := ctx.Body()
    query := ctx.Query().Get("page")
    param := ctx.Param().Get("id")
    header := ctx.Header().Get("Authorization")
    
    // Form data
    form := ctx.Form().Get("name")
    file, fileHeader, err := ctx.File("upload")
    
    // JSON handling
    var data map[string]interface{}
    ctx.BindJSON(&data)
    
    // Response
    ctx.JSON(200, zoox.H{"status": "success"})
    ctx.HTML(200, "template.html", data)
    ctx.RenderStatic("/static/", "static/")
    
    // Status codes
    ctx.Status(201)
    ctx.Error(400, "Bad Request")
}
```

### Database Integration

```go
// Cache example
cache := app.Cache()
cache.Set("key", "value", time.Hour)
value := cache.Get("key")

// Session example
session := ctx.Session()
session.Set("user_id", 123)
userID := session.Get("user_id")
```

### WebSocket Support

```go
server, err := app.WebSocket("/ws")
if err != nil {
    log.Fatal(err)
}

server.OnMessage(func(message []byte) {
    server.WriteText("Echo: " + string(message))
})
```

### Scheduled Tasks

```go
cron := app.Cron()
cron.AddJob("daily-cleanup", "0 0 * * *", func() error {
    // Daily task at midnight
    log.Println("Running daily cleanup")
    return nil
})
```

## 🔧 Configuration

Zoox supports flexible configuration through environment variables and config files:

```go
app := zoox.New()

// Environment-based configuration
app.Config.Protocol = "https"
app.Config.Host = "0.0.0.0"
app.Config.Port = 8443
app.Config.SecretKey = "your-secret-key"

// Redis configuration
app.Config.Redis.Host = "localhost"
app.Config.Redis.Port = 6379
app.Config.Redis.Password = "password"

// Session configuration
app.Config.Session.MaxAge = time.Hour
```

## 🧪 Testing

```go
func TestUserAPI(t *testing.T) {
    app := zoox.New()
    
    app.Get("/users/:id", func(ctx *zoox.Context) {
        id := ctx.Param().Get("id")
        ctx.JSON(200, zoox.H{"id": id.String()})
    })

    req := httptest.NewRequest("GET", "/users/123", nil)
    w := httptest.NewRecorder()
    
    app.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), `"id":"123"`)
}
```

## 📊 Performance

Zoox is designed for high performance:

- **Fast Router**: Trie-based routing with O(1) lookup
- **Minimal Memory**: Low memory footprint
- **Concurrent Safe**: Thread-safe design
- **Zero Allocations**: Optimized for minimal GC pressure

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
git clone https://github.com/go-zoox/zoox.git
cd zoox
go mod download
go test ./...
```

## 📄 License

Zoox is released under the [MIT License](./LICENSE).

## 🙏 Acknowledgments

- Inspired by modern web frameworks
- Built with Go's standard library
- Community-driven development

## 📞 Support

- 📖 **Documentation**: [https://go-zoox.github.io/zoox/](https://go-zoox.github.io/zoox/)
- 🐛 **Issues**: [GitHub Issues](https://github.com/go-zoox/zoox/issues)
- 💬 **Discord**: [Join our community](https://discord.gg/zoox)
- 📧 **Email**: [support@zoox.dev](mailto:support@zoox.dev)

---

**Made with ❤️ by the Zoox Team**
