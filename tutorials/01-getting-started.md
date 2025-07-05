# Getting Started with Zoox Framework

Welcome to your first Zoox tutorial! This guide will walk you through setting up your development environment and creating your first Zoox application.

## 📋 Prerequisites

### Required Knowledge
- Basic understanding of Go programming language
- Familiarity with HTTP concepts
- Basic command-line usage

### Software Requirements
- **Go 1.19 or higher** - [Download Go](https://golang.org/dl/)
- **Git** - [Download Git](https://git-scm.com/downloads)
- **Code Editor** - VS Code, GoLand, or any text editor
- **Terminal/Command Prompt**

### System Check
Let's verify your system is ready:

```bash
# Check Go version
go version

# Check Git
git --version

# Check Go modules support
go env GOMOD
```

## 🎯 Learning Objectives

By the end of this tutorial, you will:
- ✅ Understand the Zoox framework architecture
- ✅ Create and run your first Zoox application
- ✅ Handle basic HTTP requests and responses
- ✅ Understand the request lifecycle
- ✅ Know how to structure a Zoox project

## 📖 Tutorial Content

### Step 1: Create Your Project

First, let's create a new directory for your Zoox project:

```bash
# Create project directory
mkdir my-first-zoox-app
cd my-first-zoox-app

# Initialize Go module
go mod init my-first-zoox-app

# Add Zoox dependency
go get github.com/go-zoox/zoox
```

### Step 2: Create Your First Server

Create a file named `main.go`:

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-zoox/zoox"
)

func main() {
	// Create a new Zoox application
	app := zoox.New()

	// Define a simple route
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(http.StatusOK, zoox.H{
			"message":   "Hello, Zoox!",
			"timestamp": time.Now().Format(time.RFC3339),
			"status":    "success",
		})
	})

	// Start the server
	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}
```

### Step 3: Run Your Application

```bash
# Run the application
go run main.go
```

You should see:
```
🚀 Server starting on http://localhost:8080
```

### Step 4: Test Your Application

Open your browser and navigate to `http://localhost:8080`, or use curl:

```bash
curl http://localhost:8080
```

You should see:
```json
{
  "message": "Hello, Zoox!",
  "timestamp": "2023-12-07T10:30:00Z",
  "status": "success"
}
```

🎉 **Congratulations!** You've just created your first Zoox application!

### Step 5: Understanding the Code

Let's break down what we just created:

```go
// 1. Import the Zoox framework
import "github.com/go-zoox/zoox"

// 2. Create a new application instance
app := zoox.New()

// 3. Define a route handler
app.Get("/", func(ctx *zoox.Context) {
    // ctx is the context object containing request/response data
    ctx.JSON(http.StatusOK, zoox.H{
        "message": "Hello, Zoox!",
    })
})

// 4. Start the server
app.Run(":8080")
```

#### Key Concepts:

1. **Application Instance**: `zoox.New()` creates a new Zoox application
2. **Route Handler**: `app.Get()` defines a GET route
3. **Context**: `ctx` provides access to request/response data
4. **JSON Response**: `ctx.JSON()` sends a JSON response
5. **Server Start**: `app.Run()` starts the HTTP server

### Step 6: Adding More Routes

Let's expand our application with more routes:

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()

	// Home route
	app.Get("/", homeHandler)

	// About route
	app.Get("/about", aboutHandler)

	// Health check route
	app.Get("/health", healthHandler)

	// User greeting route with parameter
	app.Get("/hello/:name", greetingHandler)

	// API route with JSON response
	app.Get("/api/status", apiStatusHandler)

	log.Println("🚀 Server starting on http://localhost:8080")
	log.Println("📋 Available routes:")
	log.Println("   GET  /")
	log.Println("   GET  /about")
	log.Println("   GET  /health")
	log.Println("   GET  /hello/:name")
	log.Println("   GET  /api/status")
	
	app.Run(":8080")
}

// Route handlers
func homeHandler(ctx *zoox.Context) {
	ctx.JSON(http.StatusOK, zoox.H{
		"message": "Welcome to Zoox Framework!",
		"version": "1.0.0",
		"routes": []string{
			"GET /",
			"GET /about",
			"GET /health",
			"GET /hello/:name",
			"GET /api/status",
		},
	})
}

func aboutHandler(ctx *zoox.Context) {
	ctx.JSON(http.StatusOK, zoox.H{
		"app":         "My First Zoox App",
		"description": "A simple web application built with Zoox framework",
		"author":      "Your Name",
		"created":     time.Now().Format("2006-01-02"),
	})
}

func healthHandler(ctx *zoox.Context) {
	ctx.JSON(http.StatusOK, zoox.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    "running",
	})
}

func greetingHandler(ctx *zoox.Context) {
	name := ctx.Param().Get("name")
	ctx.JSON(http.StatusOK, zoox.H{
		"greeting": "Hello, " + name + "!",
		"message":  "Welcome to Zoox Framework",
		"name":     name,
	})
}

func apiStatusHandler(ctx *zoox.Context) {
	ctx.JSON(http.StatusOK, zoox.H{
		"api": zoox.H{
			"version":     "v1.0.0",
			"status":      "operational",
			"endpoints":   5,
			"last_update": time.Now().Format(time.RFC3339),
		},
		"server": zoox.H{
			"framework": "Zoox",
			"go_version": "1.19+",
			"environment": "development",
		},
	})
}
```

### Step 7: Test All Routes

Now test all your routes:

```bash
# Home route
curl http://localhost:8080/

# About route
curl http://localhost:8080/about

# Health check
curl http://localhost:8080/health

# Greeting with parameter
curl http://localhost:8080/hello/John

# API status
curl http://localhost:8080/api/status
```

### Step 8: Understanding Route Parameters

The greeting route demonstrates how to capture URL parameters:

```go
app.Get("/hello/:name", greetingHandler)

func greetingHandler(ctx *zoox.Context) {
    name := ctx.Param().Get("name")
    // Use the name parameter
}
```

- `:name` in the route captures the URL segment
- `ctx.Param().Get("name")` retrieves the parameter value

### Step 9: Project Structure

As your application grows, organize your code:

```
my-first-zoox-app/
├── main.go              # Application entry point
├── handlers/            # Route handlers
│   ├── home.go
│   ├── about.go
│   └── api.go
├── models/              # Data models
├── middleware/          # Custom middleware
├── static/              # Static files
├── templates/           # HTML templates
├── config/              # Configuration
└── go.mod              # Go module file
```

## 🧪 Hands-on Exercise

### Exercise 1: Create a Personal API

Create a personal information API with these endpoints:

1. `GET /me` - Return your personal information
2. `GET /me/skills` - Return your programming skills
3. `GET /me/projects` - Return your projects
4. `GET /me/contact` - Return your contact information

**Solution:**

```go
package main

import (
	"log"
	"net/http"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()

	// Personal API routes
	app.Get("/me", personalInfoHandler)
	app.Get("/me/skills", skillsHandler)
	app.Get("/me/projects", projectsHandler)
	app.Get("/me/contact", contactHandler)

	log.Println("🚀 Personal API server starting on http://localhost:8080")
	app.Run(":8080")
}

func personalInfoHandler(ctx *zoox.Context) {
	ctx.JSON(http.StatusOK, zoox.H{
		"name":        "Your Name",
		"title":       "Software Developer",
		"location":    "Your City, Country",
		"experience":  "X years",
		"bio":         "Passionate developer learning Zoox framework",
	})
}

func skillsHandler(ctx *zoox.Context) {
	ctx.JSON(http.StatusOK, zoox.H{
		"programming_languages": []string{"Go", "JavaScript", "Python"},
		"frameworks":           []string{"Zoox", "React", "Node.js"},
		"databases":            []string{"PostgreSQL", "MongoDB", "Redis"},
		"tools":                []string{"Docker", "Git", "VS Code"},
	})
}

func projectsHandler(ctx *zoox.Context) {
	ctx.JSON(http.StatusOK, zoox.H{
		"projects": []zoox.H{
			{
				"name":        "My First Zoox App",
				"description": "Learning Zoox framework basics",
				"status":      "in-progress",
				"tech_stack":  []string{"Go", "Zoox"},
			},
			{
				"name":        "Personal Portfolio",
				"description": "My personal website",
				"status":      "completed",
				"tech_stack":  []string{"HTML", "CSS", "JavaScript"},
			},
		},
	})
}

func contactHandler(ctx *zoox.Context) {
	ctx.JSON(http.StatusOK, zoox.H{
		"email":    "your.email@example.com",
		"github":   "https://github.com/yourusername",
		"linkedin": "https://linkedin.com/in/yourprofile",
		"website":  "https://yourwebsite.com",
	})
}
```

### Exercise 2: Add Error Handling

Enhance your application with proper error handling:

```go
func greetingHandler(ctx *zoox.Context) {
	name := ctx.Param().Get("name")
	
	// Validate input
	if name == "" {
		ctx.JSON(http.StatusBadRequest, zoox.H{
			"error": "Name parameter is required",
			"code":  "MISSING_NAME",
		})
		return
	}
	
	// Check for inappropriate content
	if len(name) > 50 {
		ctx.JSON(http.StatusBadRequest, zoox.H{
			"error": "Name too long (max 50 characters)",
			"code":  "NAME_TOO_LONG",
		})
		return
	}
	
	ctx.JSON(http.StatusOK, zoox.H{
		"greeting": "Hello, " + name + "!",
		"message":  "Welcome to Zoox Framework",
		"name":     name,
	})
}
```

## 📚 Additional Resources

### Documentation
- [Zoox Framework Documentation](../DOCUMENTATION.md)
- [Go HTTP Package](https://pkg.go.dev/net/http)
- [JSON in Go](https://blog.golang.org/json)

### Next Steps
- [Tutorial 02: Routing Fundamentals](./02-routing-fundamentals.md)
- [Tutorial 03: Request Response Handling](./03-request-response-handling.md)
- [Examples: Basic Server](../examples/01-basic-server/)

### Community Resources
- GitHub Repository: [go-zoox/zoox](https://github.com/go-zoox/zoox)
- Go Documentation: [golang.org/doc](https://golang.org/doc/)
- HTTP Status Codes: [httpstatuses.com](https://httpstatuses.com/)

## 🎯 Key Takeaways

1. **Zoox is Simple**: Creating a web server requires minimal code
2. **Context is Key**: The `ctx` parameter provides access to request/response data
3. **JSON Responses**: Use `ctx.JSON()` for API responses
4. **Route Parameters**: Capture URL segments with `:parameter` syntax
5. **Error Handling**: Always validate input and handle errors gracefully

## 🔄 What's Next?

Now that you've created your first Zoox application, you're ready to explore more advanced features:

1. **Routing**: Learn about advanced routing patterns
2. **Middleware**: Add functionality to your request pipeline
3. **Templates**: Render HTML pages
4. **Static Files**: Serve CSS, JavaScript, and images
5. **Database Integration**: Connect to databases

Continue with [Tutorial 02: Routing Fundamentals](./02-routing-fundamentals.md) to deepen your understanding of Zoox routing capabilities.

---

**🎉 Congratulations on completing your first Zoox tutorial!** You've taken the first step in your journey to becoming a Zoox developer. Keep practicing and exploring the framework's capabilities! 