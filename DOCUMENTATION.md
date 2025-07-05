# Zoox Framework Documentation

Zoox is a high-performance, feature-rich web framework for Go that combines simplicity with powerful capabilities. Built with modern web development practices in mind, Zoox provides everything you need to build robust web applications and APIs.

## 🚀 Quick Start

### Installation

```bash
go get github.com/go-zoox/zoox
```

### Hello World

```go
package main

import "github.com/go-zoox/zoox"

func main() {
    app := zoox.Default()
    
    app.Get("/", func(ctx *zoox.Context) {
        ctx.JSON(200, zoox.H{
            "message": "Hello, Zoox!",
        })
    })
    
    app.Run(":8080")
}
```

### 📚 Learning Resources

#### 🎓 Tutorials
- **[Getting Started Tutorial](./tutorials/01-getting-started.md)** - Your first Zoox application
- **[Complete Tutorial Series](./tutorials/README.md)** - Comprehensive learning path from beginner to advanced
- **[Interactive Examples](./examples/README.md)** - Hands-on examples with working code

#### 📖 Quick Links
- [Core Concepts](#core-concepts) - Understanding Zoox fundamentals
- [API Reference](#api-reference) - Complete API documentation
- [Examples Gallery](#examples-gallery) - Real-world examples
- [Best Practices](#best-practices) - Production-ready patterns

## 📖 Table of Contents

1. [Core Concepts](#core-concepts)
2. [Application](#application)
3. [Routing](#routing)
4. [Context](#context)
5. [Middleware](#middleware)
6. [Request Handling](#request-handling)
7. [Response Handling](#response-handling)
8. [Template Engine](#template-engine)
9. [Static Files](#static-files)
10. [WebSocket](#websocket)
11. [JSON-RPC](#json-rpc)
12. [Proxy](#proxy)
13. [Components](#components)
14. [Configuration](#configuration)
15. [Deployment](#deployment)
16. [Best Practices](#best-practices)
17. [Examples Gallery](#examples-gallery)
18. [API Reference](#api-reference)

## 🎯 Examples Gallery

### 🔰 Beginner Examples

#### [Basic Server](./examples/01-basic-server/)
Complete REST API with authentication, route groups, and error handling.

```go
app := zoox.Default()

// Basic routes
app.Get("/", homeHandler)
app.Get("/users", getUsersHandler)
app.Post("/users", createUserHandler)

// API group with middleware
api := app.Group("/api/v1")
api.Use(middleware.RequestID())
api.Use(middleware.Logger())
```

#### [Middleware Showcase](./examples/02-middleware-showcase/)
Comprehensive demonstration of all built-in middleware with custom implementations.

```go
// Global middleware
app.Use(middleware.Logger())
app.Use(middleware.Recovery())
app.Use(middleware.CORS())

// Route-specific middleware
app.Get("/protected", middleware.BasicAuth("admin", "secret"), protectedHandler)
```

### 🚀 Intermediate Examples

#### [WebSocket Chat](./examples/03-websocket-chat/)
Real-time chat application with connection management and message broadcasting.

```go
// WebSocket endpoint
app.WebSocket("/ws", func(ctx *zoox.Context) {
    conn := ctx.WebSocket()
    
    for {
        var msg Message
        if err := conn.ReadJSON(&msg); err != nil {
            break
        }
        
        // Broadcast to all clients
        chatRoom.Broadcast(msg)
    }
})
```

#### [File Upload System](./examples/04-file-upload-download/)
Complete file management system with upload, download, validation, and progress tracking.

```go
// Single file upload
app.Post("/upload/single", func(ctx *zoox.Context) {
    file, err := ctx.FormFile("file")
    if err != nil {
        ctx.JSON(400, zoox.H{"error": "No file uploaded"})
        return
    }
    
    // Validate and save file
    if err := validateFile(file); err != nil {
        ctx.JSON(400, zoox.H{"error": err.Error()})
        return
    }
    
    filename := generateUniqueFilename(file.Filename)
    ctx.SaveFile(file, filepath.Join(uploadDir, filename))
    
    ctx.JSON(200, zoox.H{
        "message": "File uploaded successfully",
        "filename": filename,
    })
})
```

#### [JSON-RPC Service](./examples/05-json-rpc-service/)
Professional JSON-RPC implementation with multiple services and interactive testing.

```go
// Math service
type MathService struct{}

func (m *MathService) Add(ctx context.Context, args *AddArgs, reply *AddReply) error {
    reply.Result = args.A + args.B
    return nil
}

// Register service
app.JSONRPC("/rpc/math", &MathService{})
```

### 🏗️ Advanced Examples

#### [Production API](./examples/06-production-api/)
Production-ready API with comprehensive security, monitoring, and deployment patterns.

```go
// Production-ready application structure
type App struct {
    config *Config
    router *zoox.Application
}

// Comprehensive middleware stack
func (app *App) setupMiddleware() {
    app.router.Use(middleware.Logger())
    app.router.Use(middleware.Recovery())
    app.router.Use(middleware.RequestID())
    app.router.Use(middleware.CORS())
    app.router.Use(middleware.Helmet())
    app.router.Use(middleware.RateLimit(100, time.Minute))
    app.router.Use(app.metricsMiddleware())
}

// Health check endpoint
func (app *App) healthHandler(ctx *zoox.Context) {
    health := HealthCheck{
        Status:    "healthy",
        Version:   "1.0.0",
        Timestamp: time.Now(),
        Services: map[string]string{
            "database": "connected",
            "cache":    "connected",
        },
    }
    ctx.JSON(200, health)
}
```

#### Database Integration *(Coming Soon)*
Complete database integration with ORM, migrations, and connection pooling.

#### Authentication System *(Coming Soon)*
Full authentication system with JWT, session management, and RBAC.

#### Microservices Architecture *(Coming Soon)*
Microservice implementation with service discovery and load balancing.

## 📚 Tutorial Series

### 🎯 Learning Paths

#### Path 1: Web Development Beginner
1. [Getting Started](./tutorials/01-getting-started.md) - Your first Zoox app
2. [Routing Fundamentals](./tutorials/02-routing-fundamentals.md) - Master routing
3. [Request/Response Handling](./tutorials/03-request-response-handling.md) - HTTP basics
4. [Middleware Basics](./tutorials/04-middleware-basics.md) - Understanding middleware
5. [Template Engine](./tutorials/06-template-engine.md) - HTML rendering
6. [Static Files](./tutorials/07-static-files-assets.md) - Serving assets

#### Path 2: API Development
1. [Getting Started](./tutorials/01-getting-started.md) - Foundation
2. [Routing Fundamentals](./tutorials/02-routing-fundamentals.md) - API routes
3. [JSON-RPC Services](./tutorials/09-json-rpc-services.md) - RPC APIs
4. [Authentication](./tutorials/10-authentication-authorization.md) - Security
5. [Database Integration](./tutorials/11-database-integration.md) - Data layer

#### Path 3: Real-time Applications
1. [Getting Started](./tutorials/01-getting-started.md) - Foundation
2. [WebSocket Development](./tutorials/08-websocket-development.md) - Real-time
3. [Caching Strategies](./tutorials/12-caching-strategies.md) - Performance
4. [Monitoring](./tutorials/13-monitoring-logging.md) - Observability

### 🚀 Quick Tutorial Access

- **⏱️ 5 Minutes**: [Hello World](./tutorials/01-getting-started.md#step-2-create-your-first-server)
- **⏱️ 15 Minutes**: [Basic REST API](./tutorials/01-getting-started.md#step-6-adding-more-routes)
- **⏱️ 30 Minutes**: [Complete Tutorial](./tutorials/01-getting-started.md)
- **⏱️ 2 Hours**: [Full Learning Path](./tutorials/README.md)

## 🤝 Contributing

We welcome contributions to improve Zoox! Here's how you can help:

### 📝 Documentation
- Improve existing documentation
- Add new examples and tutorials
- Translate documentation
- Fix typos and errors

### 💻 Code
- Fix bugs and issues
- Add new features
- Improve performance
- Write tests

### 🎯 Examples
- Create new example applications
- Improve existing examples
- Add interactive demos
- Document use cases

### 📋 Contribution Guidelines

1. **Fork the repository**
2. **Create a feature branch**
3. **Make your changes**
4. **Add tests** (if applicable)
5. **Update documentation**
6. **Submit a pull request**

For detailed contribution guidelines, see [CONTRIBUTING.md](./CONTRIBUTING.md).

## 📞 Support

### 🐛 Issues and Bugs
- [GitHub Issues](https://github.com/go-zoox/zoox/issues)
- [Bug Report Template](https://github.com/go-zoox/zoox/issues/new?template=bug_report.md)

### 💡 Feature Requests
- [Feature Request Template](https://github.com/go-zoox/zoox/issues/new?template=feature_request.md)
- [Discussions](https://github.com/go-zoox/zoox/discussions)

### 📚 Documentation
- [API Reference](#api-reference)
- [Examples](./examples/README.md)
- [Tutorials](./tutorials/README.md)

### 🌟 Community
- [GitHub Discussions](https://github.com/go-zoox/zoox/discussions)
- [Contributing Guide](./CONTRIBUTING.md)

---

**Ready to build amazing applications with Zoox? Start with our [Getting Started Tutorial](./tutorials/01-getting-started.md) or explore our [Examples Gallery](./examples/README.md)!** 🚀 