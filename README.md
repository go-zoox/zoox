# Zoox Framework - Documentation and Examples

A comprehensive collection of documentation, examples, and tutorials for the Zoox Go web framework.

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

### Run the Application

```bash
go run main.go
```

Visit `http://localhost:8080` to see the result.

## 📚 Learning Resources

### 🎓 Tutorials
- **[Getting Started Tutorial](./tutorials/01-getting-started.md)** - Your first Zoox application
- **[Complete Tutorial Series](./tutorials/README.md)** - 18 comprehensive tutorials from beginner to production
- **[Learning Paths](./tutorials/README.md#learning-paths)** - Structured learning paths for different goals

### 💡 Examples
- **[Examples Gallery](./examples/README.md)** - 5 complete, runnable examples
- **[Basic Server](./examples/01-basic-server/)** - REST API with authentication
- **[Middleware Showcase](./examples/02-middleware-showcase/)** - All built-in middleware
- **[WebSocket Chat](./examples/03-websocket-chat/)** - Real-time chat application
- **[File Upload System](./examples/04-file-upload-download/)** - Complete file management
- **[JSON-RPC Service](./examples/05-json-rpc-service/)** - Professional RPC implementation

### 📖 Documentation
- **[DOCUMENTATION.md](./DOCUMENTATION.md)** - Complete framework documentation
- **[API Reference](./DOCUMENTATION.md#api-reference)** - Full API documentation
- **[Best Practices](./DOCUMENTATION.md#best-practices)** - Production-ready patterns

## 🎯 Key Features

### 🚀 High Performance
- Efficient radix tree-based routing
- Middleware caching optimization
- Zero-allocation path parameter parsing

### 🔧 Rich Middleware
- Logging, Recovery, CORS
- Authentication, Rate Limiting, Caching
- Monitoring integration (Prometheus, Sentry)

### 🌐 Multi-Protocol Support
- HTTP/HTTPS
- WebSocket
- JSON-RPC
- Reverse Proxy

### 📦 Component Architecture
- Caching system
- Message queues
- Scheduled tasks
- Internationalization support

## 🛤️ Learning Paths

### 🔰 Beginner Path
1. **[Getting Started](./tutorials/01-getting-started.md)** - Installation and first app
2. **[Basic Server Example](./examples/01-basic-server/)** - Complete REST API
3. **[Routing Fundamentals](./tutorials/02-routing-fundamentals.md)** - Master routing
4. **[Middleware Basics](./tutorials/04-middleware-basics.md)** - Understanding middleware

### 🚀 Intermediate Path
1. **[WebSocket Development](./tutorials/08-websocket-development.md)** - Real-time features
2. **[WebSocket Chat Example](./examples/03-websocket-chat/)** - Working chat app
3. **[File Upload System](./examples/04-file-upload-download/)** - File management
4. **[JSON-RPC Services](./tutorials/09-json-rpc-services.md)** - RPC patterns

### 🏗️ Advanced Path
1. **[Authentication & Authorization](./tutorials/10-authentication-authorization.md)** - Security
2. **[Database Integration](./tutorials/11-database-integration.md)** - Data layer
3. **[Microservices Architecture](./tutorials/18-microservices-architecture.md)** - Scaling
4. **[Production Deployment](./tutorials/17-deployment-strategies.md)** - Going live

## 🎨 Example Applications

### 🔰 Beginner Examples

#### [01-basic-server](./examples/01-basic-server/)
Complete REST API with CRUD operations, authentication, and middleware.

```go
app := zoox.Default()

// User routes
app.Get("/users", getUsersHandler)
app.Post("/users", createUserHandler)

// Protected routes
protected := app.Group("/api/v1/protected")
protected.Use(middleware.BasicAuth("admin", "secret123"))
protected.Get("/dashboard", dashboardHandler)
```

#### [02-middleware-showcase](./examples/02-middleware-showcase/)
Comprehensive demonstration of all built-in middleware.

```go
// Security middleware
app.Use(middleware.CORS())
app.Use(middleware.Helmet())
app.Use(middleware.RateLimit())

// Performance middleware
app.Use(middleware.Gzip())
app.Use(middleware.Cache())
```

### 🚀 Intermediate Examples

#### [03-websocket-chat](./examples/03-websocket-chat/)
Real-time chat application with WebSocket support.

```go
app.WebSocket("/ws", func(ctx *zoox.Context) {
	conn := ctx.WebSocket()
	
	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		
		chatRoom.Broadcast(msg)
	}
})
```

#### [04-file-upload-download](./examples/04-file-upload-download/)
Complete file management system with upload, download, and validation.

```go
app.Post("/upload/single", func(ctx *zoox.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(400, zoox.H{"error": "No file uploaded"})
		return
	}
	
	// Validate and save file
	filename := generateUniqueFilename(file.Filename)
	ctx.SaveFile(file, filepath.Join(uploadDir, filename))
	
	ctx.JSON(200, zoox.H{
		"message": "File uploaded successfully",
		"filename": filename,
	})
})
```

#### [05-json-rpc-service](./examples/05-json-rpc-service/)
Professional JSON-RPC implementation with multiple services.

```go
type MathService struct{}

func (m *MathService) Add(ctx context.Context, args *AddArgs, reply *AddReply) error {
	reply.Result = args.A + args.B
	return nil
}

// Register service
app.JSONRPC("/rpc/math", &MathService{})
```

## 📋 Documentation Structure

```
📚 Documentation Modules
├── 🚀 Quick Start - Installation and first application
├── 🏗️ Core Concepts - Application and Context
├── 🛣️ Routing System - Basic routing, parameters, route groups
├── 🔧 Middleware - Built-in and custom middleware
├── 📥 Request Handling - Data retrieval, binding, file uploads
├── 📤 Response Handling - Various response types and error handling
├── 🎨 Template Engine - Template setup and rendering
├── 📁 Static Files - File serving and caching
├── 🔌 WebSocket - Real-time communication
├── 🌐 JSON-RPC - Remote procedure calls
├── 🔄 Proxy Service - Reverse proxy configuration
├── 📦 Advanced Components - Cache, queues, scheduled tasks
├── ⚙️ Configuration - Environment variables and app configuration
├── 🚀 Deployment Guide - Development and production deployment
├── 💡 Best Practices - Project structure and development standards
└── 📋 API Reference - Complete API documentation
```

## 🎯 Features Covered

### Basic Features
- ✅ Basic routing setup
- ✅ Middleware usage
- ✅ Parameter handling
- ✅ Form and JSON data processing
- ✅ Health check endpoints

### Advanced Features
- ✅ WebSocket real-time communication
- ✅ JSON-RPC services
- ✅ Proxy service configuration
- ✅ Caching system
- ✅ Scheduled tasks
- ✅ File upload and download

### Production Features
- ✅ Authentication and authorization
- ✅ Rate limiting and security
- ✅ Monitoring and logging
- ✅ Error handling and recovery
- ✅ Performance optimization
- ✅ Deployment strategies

## 🤝 Contributing

We welcome contributions to code, documentation, or examples! Please check the [Contributing Guide](CONTRIBUTING.md).

### How to Contribute Examples

1. **Create a new example directory** under `examples/`
2. **Include comprehensive documentation** in README.md
3. **Add interactive features** where possible
4. **Follow Go best practices** and coding standards
5. **Test thoroughly** before submitting

### How to Contribute Tutorials

1. **Follow the tutorial format** in `tutorials/README.md`
2. **Include hands-on exercises** with solutions
3. **Provide clear explanations** and code examples
4. **Test all code examples** for accuracy

## 📄 License

This project is open-sourced under the MIT License.

## 🔗 Related Links

- [Zoox Official Repository](https://github.com/go-zoox/zoox)
- [Go Official Documentation](https://golang.org/doc/)
- [Web Development Best Practices](https://web.dev/)

---

**Start your Zoox journey today!** 🎉
