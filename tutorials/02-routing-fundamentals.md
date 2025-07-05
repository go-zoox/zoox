# Routing Fundamentals in Zoox Framework

Learn how to master routing in Zoox, from basic route definitions to advanced patterns and organization strategies.

## 📋 Prerequisites

### Required Knowledge
- Completed [01-getting-started](./01-getting-started.md)
- Basic understanding of HTTP methods
- Familiarity with URL patterns

### Software Requirements
- Go 1.19 or higher
- Zoox framework installed

## 🎯 Learning Objectives

By the end of this tutorial, you will:
- ✅ Understand different HTTP methods and their usage
- ✅ Master route parameters and wildcards
- ✅ Organize routes using route groups
- ✅ Apply middleware to specific routes
- ✅ Handle route conflicts and precedence
- ✅ Implement dynamic route registration

## 📖 Tutorial Content

### Step 1: HTTP Methods Overview

Zoox supports all standard HTTP methods. Let's create a comprehensive example:

```go
package main

import (
	"log"
	"strconv"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	// GET - Retrieve data
	app.Get("/users", getUsersHandler)
	app.Get("/users/:id", getUserHandler)

	// POST - Create new resource
	app.Post("/users", createUserHandler)

	// PUT - Update entire resource
	app.Put("/users/:id", updateUserHandler)

	// PATCH - Partial update
	app.Patch("/users/:id", patchUserHandler)

	// DELETE - Remove resource
	app.Delete("/users/:id", deleteUserHandler)

	// HEAD - Same as GET but only headers
	app.Head("/users", headUsersHandler)

	// OPTIONS - Check allowed methods
	app.Options("/users", optionsUsersHandler)

	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}

// Handler implementations
func getUsersHandler(ctx *zoox.Context) {
	users := []zoox.H{
		{"id": 1, "name": "John Doe", "email": "john@example.com"},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
	}
	ctx.JSON(200, zoox.H{
		"users": users,
		"total": len(users),
	})
}

func getUserHandler(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(400, zoox.H{"error": "Invalid user ID"})
		return
	}

	ctx.JSON(200, zoox.H{
		"user": zoox.H{
			"id":    userID,
			"name":  "John Doe",
			"email": "john@example.com",
		},
	})
}

func createUserHandler(ctx *zoox.Context) {
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(400, zoox.H{"error": "Invalid request body"})
		return
	}

	ctx.JSON(201, zoox.H{
		"message": "User created successfully",
		"user": zoox.H{
			"id":    3,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func updateUserHandler(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := ctx.BindJSON(&user); err != nil {
		ctx.JSON(400, zoox.H{"error": "Invalid request body"})
		return
	}

	ctx.JSON(200, zoox.H{
		"message": "User updated successfully",
		"user": zoox.H{
			"id":    id,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func patchUserHandler(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	var updates map[string]interface{}

	if err := ctx.BindJSON(&updates); err != nil {
		ctx.JSON(400, zoox.H{"error": "Invalid request body"})
		return
	}

	ctx.JSON(200, zoox.H{
		"message": "User partially updated",
		"id":      id,
		"updates": updates,
	})
}

func deleteUserHandler(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	ctx.JSON(200, zoox.H{
		"message": "User deleted successfully",
		"id":      id,
	})
}

func headUsersHandler(ctx *zoox.Context) {
	ctx.Header().Set("X-Total-Count", "2")
	ctx.Status(200)
}

func optionsUsersHandler(ctx *zoox.Context) {
	ctx.Header().Set("Allow", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
	ctx.Status(200)
}
```

### Step 2: Route Parameters

Zoox supports various types of route parameters:

```go
package main

import (
	"log"
	"strings"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	// Single parameter
	app.Get("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		ctx.JSON(200, zoox.H{"user_id": id})
	})

	// Multiple parameters
	app.Get("/users/:id/posts/:post_id", func(ctx *zoox.Context) {
		userID := ctx.Param().Get("id")
		postID := ctx.Param().Get("post_id")
		ctx.JSON(200, zoox.H{
			"user_id": userID,
			"post_id": postID,
		})
	})

	// Optional parameters with default values
	app.Get("/search/:query", func(ctx *zoox.Context) {
		query := ctx.Param().Get("query")
		page := ctx.Query().Get("page", "1")
		limit := ctx.Query().Get("limit", "10")
		
		ctx.JSON(200, zoox.H{
			"query": query,
			"page":  page,
			"limit": limit,
		})
	})

	// Wildcard parameters
	app.Get("/files/*path", func(ctx *zoox.Context) {
		path := ctx.Param().Get("path")
		ctx.JSON(200, zoox.H{
			"file_path": path,
			"segments":  strings.Split(path, "/"),
		})
	})

	// Parameter validation
	app.Get("/validate/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		
		// Simple validation
		if len(id) < 1 {
			ctx.JSON(400, zoox.H{"error": "ID is required"})
			return
		}
		
		// Numeric validation
		if !isNumeric(id) {
			ctx.JSON(400, zoox.H{"error": "ID must be numeric"})
			return
		}
		
		ctx.JSON(200, zoox.H{"valid_id": id})
	})

	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
```

### Step 3: Route Groups

Route groups help organize related routes and apply common middleware:

```go
package main

import (
	"log"
	"time"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.Default()

	// Root level routes
	app.Get("/", homeHandler)
	app.Get("/health", healthHandler)

	// API v1 group
	v1 := app.Group("/api/v1")
	{
		v1.Use(middleware.Logger())
		v1.Use(middleware.RequestID())
		
		// Public routes
		v1.Get("/status", statusHandler)
		v1.Post("/register", registerHandler)
		v1.Post("/login", loginHandler)

		// Protected routes
		protected := v1.Group("/protected")
		{
			protected.Use(middleware.BasicAuth("Protected Area", map[string]string{
				"admin": "secret",
			}))
			
			protected.Get("/profile", profileHandler)
			protected.Get("/dashboard", dashboardHandler)
		}

		// User management routes
		users := v1.Group("/users")
		{
			users.Use(rateLimitMiddleware())
			
			users.Get("", getUsersHandler)
			users.Post("", createUserHandler)
			users.Get("/:id", getUserHandler)
			users.Put("/:id", updateUserHandler)
			users.Delete("/:id", deleteUserHandler)
		}
	}

	// API v2 group (future version)
	v2 := app.Group("/api/v2")
	{
		v2.Use(middleware.Logger())
		v2.Use(middleware.CORS())
		
		v2.Get("/status", func(ctx *zoox.Context) {
			ctx.JSON(200, zoox.H{
				"version": "2.0",
				"status":  "active",
			})
		})
	}

	// Admin routes
	admin := app.Group("/admin")
	{
		admin.Use(middleware.BasicAuth("Admin Area", map[string]string{
			"admin": "supersecret",
		}))
		
		admin.Get("/stats", adminStatsHandler)
		admin.Get("/logs", adminLogsHandler)
		admin.Post("/maintenance", maintenanceHandler)
	}

	log.Println("🚀 Server starting on http://localhost:8080")
	log.Println("📋 Available route groups:")
	log.Println("   /api/v1/* - API version 1")
	log.Println("   /api/v2/* - API version 2")
	log.Println("   /admin/*  - Admin panel")
	
	app.Run(":8080")
}

// Handlers
func homeHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"message": "Welcome to Zoox Routing Tutorial",
		"routes": []string{
			"/api/v1/status",
			"/api/v1/protected/profile",
			"/api/v1/users",
			"/admin/stats",
		},
	})
}

func healthHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func statusHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"api_version": "1.0",
		"status":      "operational",
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

func registerHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"message": "User registration endpoint"})
}

func loginHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"message": "User login endpoint"})
}

func profileHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"message": "User profile (protected)",
		"user":    "admin",
	})
}

func dashboardHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"message": "Dashboard (protected)",
		"stats":   zoox.H{"users": 100, "posts": 500},
	})
}

func adminStatsHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"message": "Admin statistics",
		"stats": zoox.H{
			"total_users":    1000,
			"active_users":   850,
			"total_requests": 50000,
		},
	})
}

func adminLogsHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"message": "System logs",
		"logs": []string{
			"2023-12-07 10:00:00 - Server started",
			"2023-12-07 10:01:00 - User login",
			"2023-12-07 10:02:00 - API request",
		},
	})
}

func maintenanceHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"message": "Maintenance mode toggled",
		"status":  "maintenance",
	})
}

// Custom middleware
func rateLimitMiddleware() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		// Simple rate limiting simulation
		ctx.Header().Set("X-RateLimit-Limit", "100")
		ctx.Header().Set("X-RateLimit-Remaining", "99")
		ctx.Next()
	}
}
```

### Step 4: Advanced Route Patterns

Let's explore more advanced routing patterns:

```go
package main

import (
	"log"
	"regexp"
	"strings"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	// Route with constraints
	app.Get("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		
		// Constraint: ID must be numeric
		if matched, _ := regexp.MatchString(`^\d+$`, id); !matched {
			ctx.JSON(400, zoox.H{"error": "ID must be numeric"})
			return
		}
		
		ctx.JSON(200, zoox.H{"user_id": id})
	})

	// Route with multiple wildcards
	app.Get("/api/:version/*endpoint", func(ctx *zoox.Context) {
		version := ctx.Param().Get("version")
		endpoint := ctx.Param().Get("endpoint")
		
		ctx.JSON(200, zoox.H{
			"api_version": version,
			"endpoint":    endpoint,
			"path_parts":  strings.Split(endpoint, "/"),
		})
	})

	// Route with optional segments
	app.Get("/search", searchHandler)
	app.Get("/search/:category", searchHandler)
	app.Get("/search/:category/:subcategory", searchHandler)

	// Route with file extensions
	app.Get("/files/:filename", func(ctx *zoox.Context) {
		filename := ctx.Param().Get("filename")
		
		// Extract extension
		parts := strings.Split(filename, ".")
		var extension string
		if len(parts) > 1 {
			extension = parts[len(parts)-1]
		}
		
		ctx.JSON(200, zoox.H{
			"filename":  filename,
			"extension": extension,
			"mime_type": getMimeType(extension),
		})
	})

	// Dynamic route registration
	registerDynamicRoutes(app)

	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}

func searchHandler(ctx *zoox.Context) {
	category := ctx.Param().Get("category")
	subcategory := ctx.Param().Get("subcategory")
	query := ctx.Query().Get("q", "")
	
	result := zoox.H{
		"query": query,
	}
	
	if category != "" {
		result["category"] = category
	}
	if subcategory != "" {
		result["subcategory"] = subcategory
	}
	
	ctx.JSON(200, result)
}

func getMimeType(extension string) string {
	mimeTypes := map[string]string{
		"txt":  "text/plain",
		"html": "text/html",
		"css":  "text/css",
		"js":   "application/javascript",
		"json": "application/json",
		"png":  "image/png",
		"jpg":  "image/jpeg",
		"pdf":  "application/pdf",
	}
	
	if mimeType, exists := mimeTypes[extension]; exists {
		return mimeType
	}
	return "application/octet-stream"
}

func registerDynamicRoutes(app *zoox.Application) {
	// Simulate dynamic route registration
	routes := []struct {
		method  string
		path    string
		handler func(*zoox.Context)
	}{
		{"GET", "/dynamic/route1", dynamicHandler1},
		{"POST", "/dynamic/route2", dynamicHandler2},
		{"PUT", "/dynamic/route3", dynamicHandler3},
	}
	
	for _, route := range routes {
		switch route.method {
		case "GET":
			app.Get(route.path, route.handler)
		case "POST":
			app.Post(route.path, route.handler)
		case "PUT":
			app.Put(route.path, route.handler)
		}
	}
}

func dynamicHandler1(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"message": "Dynamic route 1"})
}

func dynamicHandler2(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"message": "Dynamic route 2"})
}

func dynamicHandler3(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"message": "Dynamic route 3"})
}
```

## 🧪 Hands-on Exercise

### Exercise 1: Build a Blog API

Create a complete blog API with the following requirements:

1. **Route Structure:**
   - `/api/v1/posts` - List all posts
   - `/api/v1/posts/:id` - Get specific post
   - `/api/v1/posts/:id/comments` - Get post comments
   - `/api/v1/authors/:author_id/posts` - Get posts by author

2. **Route Groups:**
   - Public routes (no authentication)
   - Protected routes (require authentication)
   - Admin routes (require admin role)

3. **Parameters:**
   - Support pagination with query parameters
   - Validate numeric IDs
   - Handle optional sorting parameters

### Solution:

```go
package main

import (
	"log"
	"strconv"
	"time"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

type Post struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	AuthorID int       `json:"author_id"`
	Created  time.Time `json:"created"`
}

type Comment struct {
	ID      int       `json:"id"`
	PostID  int       `json:"post_id"`
	Content string    `json:"content"`
	Author  string    `json:"author"`
	Created time.Time `json:"created"`
}

func main() {
	app := zoox.Default()

	// Sample data
	posts := []Post{
		{1, "First Post", "Hello World", 1, time.Now()},
		{2, "Second Post", "Learning Zoox", 1, time.Now()},
		{3, "Third Post", "Advanced Routing", 2, time.Now()},
	}

	comments := []Comment{
		{1, 1, "Great post!", "User1", time.Now()},
		{2, 1, "Thanks for sharing", "User2", time.Now()},
		{3, 2, "Very helpful", "User3", time.Now()},
	}

	// API v1
	api := app.Group("/api/v1")
	{
		// Public routes
		api.Get("/posts", func(ctx *zoox.Context) {
			page, _ := strconv.Atoi(ctx.Query().Get("page", "1"))
			limit, _ := strconv.Atoi(ctx.Query().Get("limit", "10"))
			sort := ctx.Query().Get("sort", "created")
			
			ctx.JSON(200, zoox.H{
				"posts": posts,
				"pagination": zoox.H{
					"page":  page,
					"limit": limit,
					"total": len(posts),
				},
				"sort": sort,
			})
		})

		api.Get("/posts/:id", func(ctx *zoox.Context) {
			id, err := strconv.Atoi(ctx.Param().Get("id"))
			if err != nil {
				ctx.JSON(400, zoox.H{"error": "Invalid post ID"})
				return
			}

			for _, post := range posts {
				if post.ID == id {
					ctx.JSON(200, zoox.H{"post": post})
					return
				}
			}
			
			ctx.JSON(404, zoox.H{"error": "Post not found"})
		})

		api.Get("/posts/:id/comments", func(ctx *zoox.Context) {
			id, err := strconv.Atoi(ctx.Param().Get("id"))
			if err != nil {
				ctx.JSON(400, zoox.H{"error": "Invalid post ID"})
				return
			}

			var postComments []Comment
			for _, comment := range comments {
				if comment.PostID == id {
					postComments = append(postComments, comment)
				}
			}

			ctx.JSON(200, zoox.H{
				"comments": postComments,
				"total":    len(postComments),
			})
		})

		api.Get("/authors/:author_id/posts", func(ctx *zoox.Context) {
			authorID, err := strconv.Atoi(ctx.Param().Get("author_id"))
			if err != nil {
				ctx.JSON(400, zoox.H{"error": "Invalid author ID"})
				return
			}

			var authorPosts []Post
			for _, post := range posts {
				if post.AuthorID == authorID {
					authorPosts = append(authorPosts, post)
				}
			}

			ctx.JSON(200, zoox.H{
				"posts":     authorPosts,
				"author_id": authorID,
				"total":     len(authorPosts),
			})
		})

		// Protected routes
		protected := api.Group("/protected")
		{
			protected.Use(middleware.BasicAuth("Protected", map[string]string{
				"user": "password",
			}))

			protected.Post("/posts", func(ctx *zoox.Context) {
				ctx.JSON(201, zoox.H{"message": "Post created"})
			})

			protected.Put("/posts/:id", func(ctx *zoox.Context) {
				id := ctx.Param().Get("id")
				ctx.JSON(200, zoox.H{"message": "Post updated", "id": id})
			})
		}

		// Admin routes
		admin := api.Group("/admin")
		{
			admin.Use(middleware.BasicAuth("Admin", map[string]string{
				"admin": "secret",
			}))

			admin.Delete("/posts/:id", func(ctx *zoox.Context) {
				id := ctx.Param().Get("id")
				ctx.JSON(200, zoox.H{"message": "Post deleted", "id": id})
			})

			admin.Get("/stats", func(ctx *zoox.Context) {
				ctx.JSON(200, zoox.H{
					"total_posts":    len(posts),
					"total_comments": len(comments),
					"authors":        2,
				})
			})
		}
	}

	log.Println("🚀 Blog API Server starting on http://localhost:8080")
	log.Println("📋 Try these endpoints:")
	log.Println("   GET  /api/v1/posts")
	log.Println("   GET  /api/v1/posts/1")
	log.Println("   GET  /api/v1/posts/1/comments")
	log.Println("   GET  /api/v1/authors/1/posts")
	log.Println("   POST /api/v1/protected/posts (user:password)")
	log.Println("   GET  /api/v1/admin/stats (admin:secret)")
	
	app.Run(":8080")
}
```

## 📚 Key Takeaways

1. **HTTP Methods**: Use appropriate HTTP methods for different operations
2. **Route Parameters**: Use `:param` for single values and `*param` for wildcards
3. **Route Groups**: Organize related routes and apply common middleware
4. **Parameter Validation**: Always validate route parameters
5. **Route Precedence**: More specific routes should be defined before general ones
6. **Middleware Application**: Apply middleware at the appropriate level (global, group, or route)

## 📖 Additional Resources

- [Zoox Routing Documentation](../DOCUMENTATION.md#routing)
- [HTTP Methods Reference](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods)
- [URL Pattern Matching](https://tools.ietf.org/html/rfc3986)
- [Next Tutorial: Request & Response Handling](./03-request-response-handling.md)

## 🔗 What's Next?

In the next tutorial, we'll dive deep into request and response handling, learning how to:
- Parse different types of request data
- Handle form data and file uploads
- Send various response formats
- Implement proper error handling

Continue to [Tutorial 03: Request & Response Handling](./03-request-response-handling.md)! 