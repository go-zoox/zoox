# Tutorial 01: Getting Started with Zoox

Welcome to your first Zoox tutorial! In this tutorial, you'll learn the fundamentals of the Zoox Go web framework by building your first web application.

## 🎯 Learning Objectives

By the end of this tutorial, you will:
- Understand what Zoox is and its key features
- Set up a new Zoox project
- Create your first HTTP routes
- Handle different HTTP methods
- Serve static content
- Understand the request-response lifecycle

## ⏱️ Estimated Time: 30 minutes

## 📋 Prerequisites

- Go 1.19 or higher installed
- Basic knowledge of Go programming
- Text editor or IDE
- Terminal/command line access

## 🚀 What is Zoox?

Zoox is a modern, fast, and feature-rich web framework for Go that provides:

- **High Performance** - Built for speed and efficiency
- **Rich Middleware** - Comprehensive middleware ecosystem
- **Easy to Use** - Simple and intuitive API
- **Production Ready** - Built-in features for production deployment
- **Extensible** - Flexible architecture for custom needs

## 📝 Step 1: Project Setup

### 1.1 Create a New Directory

```bash
mkdir my-first-zoox-app
cd my-first-zoox-app
```

### 1.2 Initialize Go Module

```bash
go mod init my-first-zoox-app
```

### 1.3 Install Zoox

```bash
go get github.com/go-zoox/zoox
```

## 🏗️ Step 2: Your First Zoox Application

### 2.1 Create main.go

Create a file named `main.go` with the following content:

```go
package main

import (
    "net/http"
    
    "github.com/go-zoox/zoox"
)

func main() {
    // Create a new Zoox application
    app := zoox.Default()
    
    // Define a simple route
    app.Get("/", func(ctx *zoox.Context) {
        ctx.String(http.StatusOK, "Hello, Zoox!")
    })
    
    // Start the server on port 8080
    app.Run(":8080")
}
```

### 2.2 Run Your Application

```bash
go run main.go
```

You should see output similar to:
```
[ZOOX] Listening and serving HTTP on :8080
```

### 2.3 Test Your Application

Open your browser and navigate to `http://localhost:8080`

You should see: **Hello, Zoox!**

🎉 **Congratulations!** You've created your first Zoox application!

## 🛣️ Step 3: Adding More Routes

Let's add more routes to understand different response types:

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/go-zoox/zoox"
)

func main() {
    app := zoox.Default()
    
    // String response
    app.Get("/", func(ctx *zoox.Context) {
        ctx.String(http.StatusOK, "Hello, Zoox!")
    })
    
    // JSON response
    app.Get("/json", func(ctx *zoox.Context) {
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "message": "Hello from Zoox!",
            "status":  "success",
            "time":    time.Now(),
        })
    })
    
    // HTML response
    app.Get("/html", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Zoox Tutorial</title>
        </head>
        <body>
            <h1>Welcome to Zoox!</h1>
            <p>This is an HTML response from your Zoox application.</p>
            <a href="/">Go back to home</a>
        </body>
        </html>
        `
        ctx.HTML(http.StatusOK, html)
    })
    
    app.Run(":8080")
}
```

### 3.1 Test the New Routes

- `http://localhost:8080/` - String response
- `http://localhost:8080/json` - JSON response  
- `http://localhost:8080/html` - HTML response

## 📊 Step 4: Working with HTTP Methods

Zoox supports all standard HTTP methods. Let's add some examples:

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/go-zoox/zoox"
)

func main() {
    app := zoox.Default()
    
    // GET route
    app.Get("/users", func(ctx *zoox.Context) {
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "users": []string{"Alice", "Bob", "Charlie"},
        })
    })
    
    // POST route
    app.Post("/users", func(ctx *zoox.Context) {
        ctx.JSON(http.StatusCreated, map[string]interface{}{
            "message": "User created successfully",
            "user_id": 123,
            "created_at": time.Now(),
        })
    })
    
    // PUT route
    app.Put("/users/:id", func(ctx *zoox.Context) {
        userID := ctx.Param("id")
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "message": "User updated successfully",
            "user_id": userID,
            "updated_at": time.Now(),
        })
    })
    
    // DELETE route
    app.Delete("/users/:id", func(ctx *zoox.Context) {
        userID := ctx.Param("id")
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "message": "User deleted successfully",
            "user_id": userID,
            "deleted_at": time.Now(),
        })
    })
    
    // Catch-all route for undefined endpoints
    app.NoRoute(func(ctx *zoox.Context) {
        ctx.JSON(http.StatusNotFound, map[string]interface{}{
            "error": "Route not found",
            "path":  ctx.Path,
        })
    })
    
    app.Run(":8080")
}
```

### 4.1 Test HTTP Methods

Using curl or a tool like Postman:

```bash
# GET
curl http://localhost:8080/users

# POST
curl -X POST http://localhost:8080/users

# PUT
curl -X PUT http://localhost:8080/users/123

# DELETE
curl -X DELETE http://localhost:8080/users/123

# Test 404
curl http://localhost:8080/nonexistent
```

## 🔄 Step 5: Understanding the Context

The `zoox.Context` is the heart of every request. It provides access to:

- **Request data** - Headers, parameters, body
- **Response methods** - JSON, HTML, String, etc.
- **Route parameters** - URL path parameters
- **Query parameters** - URL query strings

Let's explore these features:

```go
package main

import (
    "net/http"
    
    "github.com/go-zoox/zoox"
)

func main() {
    app := zoox.Default()
    
    // Route with path parameter
    app.Get("/hello/:name", func(ctx *zoox.Context) {
        name := ctx.Param("name")
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "message": "Hello, " + name + "!",
            "path_param": name,
        })
    })
    
    // Route with query parameters
    app.Get("/search", func(ctx *zoox.Context) {
        query := ctx.Query().Get("q", "")
        page := ctx.Query().Get("page", "1")
        
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "search_query": query,
            "page": page,
            "results": []string{"Result 1", "Result 2", "Result 3"},
        })
    })
    
    // Route showing request headers
    app.Get("/headers", func(ctx *zoox.Context) {
        userAgent := ctx.Header("User-Agent")
        contentType := ctx.Header("Content-Type")
        
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "user_agent": userAgent,
            "content_type": contentType,
            "method": ctx.Method,
            "path": ctx.Path,
        })
    })
    
    app.Run(":8080")
}
```

### 5.1 Test Context Features

```bash
# Path parameter
curl http://localhost:8080/hello/Alice

# Query parameters  
curl "http://localhost:8080/search?q=zoox&page=2"

# Headers
curl -H "User-Agent: MyApp/1.0" http://localhost:8080/headers
```

## 📄 Step 6: Adding Basic Middleware

Middleware functions run before your route handlers. Let's add some basic logging:

```go
package main

import (
    "fmt"
    "net/http"
    "time"
    
    "github.com/go-zoox/zoox"
)

func main() {
    app := zoox.Default()
    
    // Custom middleware
    app.Use(func(ctx *zoox.Context) {
        start := time.Now()
        
        // Process request
        ctx.Next()
        
        // Log after request
        duration := time.Since(start)
        fmt.Printf("[%s] %s %s - %v\n", 
            time.Now().Format("2006-01-02 15:04:05"),
            ctx.Method, 
            ctx.Path, 
            duration,
        )
    })
    
    app.Get("/", func(ctx *zoox.Context) {
        // Simulate some work
        time.Sleep(100 * time.Millisecond)
        ctx.String(http.StatusOK, "Hello, Zoox!")
    })
    
    app.Get("/fast", func(ctx *zoox.Context) {
        ctx.String(http.StatusOK, "Fast response!")
    })
    
    app.Run(":8080")
}
```

## 🏁 Step 7: Complete Example

Here's a complete example that demonstrates everything we've learned:

```go
package main

import (
    "fmt"
    "net/http"
    "time"
    
    "github.com/go-zoox/zoox"
)

func main() {
    // Create application with default middleware (Logger, Recovery)
    app := zoox.Default()
    
    // Custom middleware for request timing
    app.Use(func(ctx *zoox.Context) {
        start := time.Now()
        ctx.Next()
        duration := time.Since(start)
        ctx.Header("X-Response-Time", duration.String())
    })
    
    // Home page
    app.Get("/", func(ctx *zoox.Context) {
        ctx.HTML(http.StatusOK, `
        <!DOCTYPE html>
        <html>
        <head>
            <title>My First Zoox App</title>
            <style>body { font-family: Arial; margin: 40px; }</style>
        </head>
        <body>
            <h1>🚀 Welcome to My First Zoox App!</h1>
            <h2>Available Endpoints:</h2>
            <ul>
                <li><a href="/api/status">GET /api/status</a> - API status</li>
                <li><a href="/api/time">GET /api/time</a> - Current time</li>
                <li><a href="/hello/YourName">GET /hello/:name</a> - Personalized greeting</li>
                <li><a href="/search?q=test">GET /search</a> - Search with query params</li>
            </ul>
        </body>
        </html>
        `)
    })
    
    // API routes group
    api := app.Group("/api")
    
    api.Get("/status", func(ctx *zoox.Context) {
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "status": "healthy",
            "service": "my-first-zoox-app",
            "version": "1.0.0",
            "timestamp": time.Now(),
        })
    })
    
    api.Get("/time", func(ctx *zoox.Context) {
        now := time.Now()
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "current_time": now,
            "unix_timestamp": now.Unix(),
            "formatted": now.Format("2006-01-02 15:04:05"),
            "timezone": now.Location().String(),
        })
    })
    
    // Personalized greeting
    app.Get("/hello/:name", func(ctx *zoox.Context) {
        name := ctx.Param("name")
        greeting := ctx.Query().Get("greeting", "Hello")
        
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "message": fmt.Sprintf("%s, %s!", greeting, name),
            "name": name,
            "greeting": greeting,
            "timestamp": time.Now(),
        })
    })
    
    // Search endpoint
    app.Get("/search", func(ctx *zoox.Context) {
        query := ctx.Query().Get("q", "")
        if query == "" {
            ctx.JSON(http.StatusBadRequest, map[string]interface{}{
                "error": "Query parameter 'q' is required",
            })
            return
        }
        
        // Mock search results
        results := []map[string]interface{}{
            {"id": 1, "title": "Result 1 for " + query, "score": 0.95},
            {"id": 2, "title": "Result 2 for " + query, "score": 0.87},
            {"id": 3, "title": "Result 3 for " + query, "score": 0.75},
        }
        
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "query": query,
            "results": results,
            "total": len(results),
            "search_time": "0.05s",
        })
    })
    
    // 404 handler
    app.NoRoute(func(ctx *zoox.Context) {
        ctx.JSON(http.StatusNotFound, map[string]interface{}{
            "error": "Endpoint not found",
            "path": ctx.Path,
            "method": ctx.Method,
            "available_endpoints": []string{
                "GET /",
                "GET /api/status", 
                "GET /api/time",
                "GET /hello/:name",
                "GET /search?q=term",
            },
        })
    })
    
    // Start server
    fmt.Println("🚀 Starting Zoox application...")
    fmt.Println("📍 Server running on http://localhost:8080")
    fmt.Println("🌐 Open your browser to http://localhost:8080")
    
    app.Run(":8080")
}
```

## ✅ Testing Your Complete Application

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Test in browser:**
   - Visit `http://localhost:8080`
   - Try the different endpoints listed

3. **Test with curl:**
   ```bash
   curl http://localhost:8080/api/status
   curl http://localhost:8080/hello/Alice?greeting=Hi
   curl "http://localhost:8080/search?q=zoox"
   ```

## 🎯 What You've Learned

In this tutorial, you've learned:

✅ **Basic Application Setup** - Creating and configuring a Zoox app  
✅ **HTTP Routing** - Handling different HTTP methods and routes  
✅ **Response Types** - Sending String, JSON, and HTML responses  
✅ **Context Usage** - Accessing request data and setting responses  
✅ **Path Parameters** - Extracting values from URL paths  
✅ **Query Parameters** - Reading URL query strings  
✅ **Middleware** - Adding custom middleware for cross-cutting concerns  
✅ **Route Groups** - Organizing related routes  
✅ **Error Handling** - Managing 404 and other error responses  

## 🚀 Next Steps

Now that you understand the basics, you're ready to move on to:

- **[Tutorial 02: Routing Fundamentals](../02-routing-fundamentals/)** - Advanced routing patterns
- **[Tutorial 03: Request & Response Handling](../03-request-response-handling/)** - Data validation and processing
- **[Tutorial 04: Middleware Basics](../04-middleware-basics/)** - Building custom middleware

## 💡 Practice Exercises

Try these exercises to reinforce your learning:

1. **Personal API** - Create an API that returns information about yourself
2. **Calculator Service** - Build a simple calculator with endpoints for math operations
3. **Todo API** - Create basic CRUD endpoints for a todo list
4. **Weather Mock** - Build a mock weather API with different cities

## 🤝 Need Help?

- Check the [main documentation](../../DOCUMENTATION.md)
- Look at [examples](../../examples/) for more code samples
- Join our community discussions
- Ask questions on Stack Overflow with the `zoox-framework` tag

---

🎉 **Congratulations on completing Tutorial 01!** You're now ready to build web applications with Zoox! 