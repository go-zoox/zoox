# Middleware Basics in Zoox Framework

Learn how to use and create middleware in Zoox to add cross-cutting concerns like authentication, logging, and error handling to your applications.

## 📋 Prerequisites

### Required Knowledge
- Completed [03-request-response-handling](./03-request-response-handling.md)
- Understanding of HTTP request/response cycle
- Basic Go function concepts

### Software Requirements
- Go 1.19 or higher
- Zoox framework installed

## 🎯 Learning Objectives

By the end of this tutorial, you will:
- ✅ Understand what middleware is and how it works
- ✅ Use built-in middleware effectively
- ✅ Create custom middleware functions
- ✅ Chain middleware for complex processing
- ✅ Handle middleware errors and recovery
- ✅ Apply middleware at different levels (global, group, route)

## 📖 Tutorial Content

### Step 1: Understanding Middleware

Middleware are functions that execute during the request-response cycle. They can:
- Execute code before the request reaches the handler
- Execute code after the handler completes
- Modify the request or response
- Terminate the request early

```go
package main

import (
	"log"
	"time"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	// Simple middleware example
	app.Use(func(ctx *zoox.Context) {
		log.Printf("Before handler - %s %s", ctx.Method, ctx.Path)
		
		// Continue to next middleware/handler
		ctx.Next()
		
		log.Printf("After handler - %s %s", ctx.Method, ctx.Path)
	})

	// Timing middleware
	app.Use(func(ctx *zoox.Context) {
		start := time.Now()
		
		ctx.Next()
		
		duration := time.Since(start)
		log.Printf("Request took %v", duration)
	})

	// Routes
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"message": "Hello World"})
	})

	app.Get("/users", func(ctx *zoox.Context) {
		time.Sleep(100 * time.Millisecond) // Simulate processing
		ctx.JSON(200, zoox.H{"users": []string{"John", "Jane"}})
	})

	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}
```

### Step 2: Built-in Middleware

Zoox provides several built-in middleware for common use cases:

```go
package main

import (
	"log"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.Default()

	// Logger middleware
	app.Use(middleware.Logger())

	// Recovery middleware (handles panics)
	app.Use(middleware.Recovery())

	// Request ID middleware
	app.Use(middleware.RequestID())

	// CORS middleware
	app.Use(middleware.CORS())

	// Gzip compression
	app.Use(middleware.Gzip())

	// Rate limiting
	app.Use(middleware.RateLimit(100)) // 100 requests per minute

	// Basic authentication
	app.Use(middleware.BasicAuth("Protected Area", map[string]string{
		"admin": "secret",
		"user":  "password",
	}))

	// Routes
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message":    "Middleware demo",
			"request_id": ctx.Header().Get("X-Request-ID"),
		})
	})

	app.Get("/panic", func(ctx *zoox.Context) {
		panic("Test panic - should be caught by recovery middleware")
	})

	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}
```

### Step 3: Custom Middleware

Create your own middleware for specific needs:

```go
package main

import (
	"log"
	"strings"
	"time"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	// Custom logging middleware
	app.Use(customLogger())

	// API key authentication middleware
	app.Use(apiKeyAuth())

	// Request validation middleware
	app.Use(validateRequest())

	// Security headers middleware
	app.Use(securityHeaders())

	// Routes
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"message": "Protected endpoint"})
	})

	app.Post("/data", func(ctx *zoox.Context) {
		var data map[string]interface{}
		ctx.BindJSON(&data)
		ctx.JSON(200, zoox.H{"received": data})
	})

	log.Println("🚀 Server starting on http://localhost:8080")
	log.Println("📋 Use API key: X-API-Key: secret-key")
	app.Run(":8080")
}

// Custom logger middleware
func customLogger() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		start := time.Now()
		
		// Log request
		log.Printf("[%s] %s %s - %s",
			start.Format("2006-01-02 15:04:05"),
			ctx.Method,
			ctx.Path,
			ctx.IP(),
		)
		
		ctx.Next()
		
		// Log response
		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d - %v",
			time.Now().Format("2006-01-02 15:04:05"),
			ctx.Method,
			ctx.Path,
			ctx.Status(),
			duration,
		)
	}
}

// API key authentication middleware
func apiKeyAuth() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		apiKey := ctx.Header().Get("X-API-Key")
		
		if apiKey == "" {
			ctx.JSON(401, zoox.H{
				"error": "Missing API key",
				"message": "Please provide X-API-Key header",
			})
			ctx.Abort()
			return
		}
		
		// Validate API key (in real app, check against database)
		validKeys := []string{"secret-key", "another-key"}
		valid := false
		for _, key := range validKeys {
			if apiKey == key {
				valid = true
				break
			}
		}
		
		if !valid {
			ctx.JSON(401, zoox.H{
				"error": "Invalid API key",
			})
			ctx.Abort()
			return
		}
		
		// Store API key info for later use
		ctx.Set("api_key", apiKey)
		ctx.Next()
	}
}

// Request validation middleware
func validateRequest() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		// Only validate POST requests
		if ctx.Method != "POST" {
			ctx.Next()
			return
		}
		
		contentType := ctx.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			ctx.JSON(400, zoox.H{
				"error": "Invalid content type",
				"message": "Only application/json is supported",
			})
			ctx.Abort()
			return
		}
		
		ctx.Next()
	}
}

// Security headers middleware
func securityHeaders() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		// Set security headers
		ctx.Header().Set("X-Content-Type-Options", "nosniff")
		ctx.Header().Set("X-Frame-Options", "DENY")
		ctx.Header().Set("X-XSS-Protection", "1; mode=block")
		ctx.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		ctx.Next()
	}
}
```

### Step 4: Middleware Chaining and Order

The order of middleware matters. They execute in the order they're added:

```go
package main

import (
	"log"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	// Middleware execute in order
	app.Use(middleware1())
	app.Use(middleware2())
	app.Use(middleware3())

	// Group-level middleware
	api := app.Group("/api")
	{
		api.Use(apiMiddleware())
		
		api.Get("/users", func(ctx *zoox.Context) {
			ctx.JSON(200, zoox.H{"users": []string{"John", "Jane"}})
		})
		
		// Route-specific middleware
		api.Get("/admin", adminMiddleware(), func(ctx *zoox.Context) {
			ctx.JSON(200, zoox.H{"message": "Admin area"})
		})
	}

	// Routes without group middleware
	app.Get("/public", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"message": "Public endpoint"})
	})

	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}

func middleware1() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		log.Println("Middleware 1 - Before")
		ctx.Next()
		log.Println("Middleware 1 - After")
	}
}

func middleware2() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		log.Println("Middleware 2 - Before")
		ctx.Next()
		log.Println("Middleware 2 - After")
	}
}

func middleware3() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		log.Println("Middleware 3 - Before")
		ctx.Next()
		log.Println("Middleware 3 - After")
	}
}

func apiMiddleware() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		log.Println("API Middleware - Setting API headers")
		ctx.Header().Set("X-API-Version", "1.0")
		ctx.Next()
	}
}

func adminMiddleware() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		log.Println("Admin Middleware - Checking admin access")
		// Simulate admin check
		isAdmin := ctx.Header().Get("X-Admin") == "true"
		if !isAdmin {
			ctx.JSON(403, zoox.H{"error": "Admin access required"})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
```

### Step 5: Error Handling in Middleware

Proper error handling is crucial in middleware:

```go
package main

import (
	"errors"
	"log"
	"time"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	// Error handling middleware (should be first)
	app.Use(errorHandler())

	// Panic recovery middleware
	app.Use(panicRecovery())

	// Timeout middleware
	app.Use(timeoutMiddleware(5 * time.Second))

	// Validation middleware
	app.Use(validationMiddleware())

	// Routes
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"message": "Success"})
	})

	app.Get("/error", func(ctx *zoox.Context) {
		ctx.Error(errors.New("something went wrong"))
	})

	app.Get("/panic", func(ctx *zoox.Context) {
		panic("test panic")
	})

	app.Get("/slow", func(ctx *zoox.Context) {
		time.Sleep(10 * time.Second) // Will timeout
		ctx.JSON(200, zoox.H{"message": "Slow response"})
	})

	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}

// Error handling middleware
func errorHandler() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		ctx.Next()
		
		// Check for errors after all middleware/handlers
		if len(ctx.Errors) > 0 {
			err := ctx.Errors[0]
			log.Printf("Error occurred: %v", err)
			
			// Don't send response if already sent
			if ctx.IsAborted() {
				return
			}
			
			ctx.JSON(500, zoox.H{
				"error": "Internal server error",
				"message": err.Error(),
			})
		}
	}
}

// Panic recovery middleware
func panicRecovery() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v", r)
				
				if !ctx.IsAborted() {
					ctx.JSON(500, zoox.H{
						"error": "Internal server error",
						"message": "Server panic occurred",
					})
				}
			}
		}()
		
		ctx.Next()
	}
}

// Timeout middleware
func timeoutMiddleware(timeout time.Duration) func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		done := make(chan bool, 1)
		
		go func() {
			ctx.Next()
			done <- true
		}()
		
		select {
		case <-done:
			// Request completed normally
		case <-time.After(timeout):
			// Request timed out
			log.Printf("Request timeout: %s %s", ctx.Method, ctx.Path)
			if !ctx.IsAborted() {
				ctx.JSON(408, zoox.H{
					"error": "Request timeout",
					"timeout": timeout.String(),
				})
				ctx.Abort()
			}
		}
	}
}

// Validation middleware
func validationMiddleware() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		// Example: validate user agent
		userAgent := ctx.UserAgent().String()
		if userAgent == "" {
			ctx.JSON(400, zoox.H{
				"error": "Missing User-Agent header",
			})
			ctx.Abort()
			return
		}
		
		// Block certain user agents
		blocked := []string{"BadBot", "Crawler"}
		for _, bot := range blocked {
			if userAgent == bot {
				ctx.JSON(403, zoox.H{
					"error": "Blocked user agent",
				})
				ctx.Abort()
				return
			}
		}
		
		ctx.Next()
	}
}
```

### Step 6: Advanced Middleware Patterns

```go
package main

import (
	"log"
	"sync"
	"time"

	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.Default()

	// Conditional middleware
	app.Use(conditionalMiddleware())

	// Caching middleware
	app.Use(cacheMiddleware())

	// Rate limiting with different limits per endpoint
	app.Use(advancedRateLimit())

	// Request/Response transformation
	app.Use(transformMiddleware())

	// Routes
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"message": "Hello World"})
	})

	app.Get("/cached", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "This response is cached",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	app.Get("/limited", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"message": "Rate limited endpoint"})
	})

	log.Println("🚀 Server starting on http://localhost:8080")
	app.Run(":8080")
}

// Conditional middleware - only applies to certain paths
func conditionalMiddleware() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		// Only apply to paths starting with /api
		if !strings.HasPrefix(ctx.Path, "/api") {
			ctx.Next()
			return
		}
		
		log.Println("API-specific middleware executed")
		ctx.Header().Set("X-API-Processed", "true")
		ctx.Next()
	}
}

// Simple caching middleware
func cacheMiddleware() func(*zoox.Context) {
	cache := make(map[string]cacheItem)
	var mu sync.RWMutex
	
	type cacheItem struct {
		data      interface{}
		timestamp time.Time
		ttl       time.Duration
	}
	
	return func(ctx *zoox.Context) {
		// Only cache GET requests
		if ctx.Method != "GET" {
			ctx.Next()
			return
		}
		
		key := ctx.Path
		mu.RLock()
		item, exists := cache[key]
		mu.RUnlock()
		
		// Check if cached and not expired
		if exists && time.Since(item.timestamp) < item.ttl {
			log.Printf("Cache hit for %s", key)
			ctx.Header().Set("X-Cache", "HIT")
			ctx.JSON(200, item.data)
			return
		}
		
		// Continue to handler
		ctx.Next()
		
		// Cache the response (simplified)
		if ctx.Status() == 200 {
			mu.Lock()
			cache[key] = cacheItem{
				data:      zoox.H{"cached": true, "original_time": time.Now().Format(time.RFC3339)},
				timestamp: time.Now(),
				ttl:       1 * time.Minute,
			}
			mu.Unlock()
			log.Printf("Cached response for %s", key)
		}
	}
}

// Advanced rate limiting with different limits per endpoint
func advancedRateLimit() func(*zoox.Context) {
	type rateLimiter struct {
		requests  int
		resetTime time.Time
		limit     int
	}
	
	limiters := make(map[string]*rateLimiter)
	var mu sync.RWMutex
	
	return func(ctx *zoox.Context) {
		// Different limits for different endpoints
		var limit int
		switch ctx.Path {
		case "/limited":
			limit = 5 // 5 requests per minute
		default:
			limit = 60 // 60 requests per minute
		}
		
		key := ctx.IP() + ":" + ctx.Path
		
		mu.Lock()
		limiter, exists := limiters[key]
		if !exists {
			limiter = &rateLimiter{
				requests:  0,
				resetTime: time.Now().Add(time.Minute),
				limit:     limit,
			}
			limiters[key] = limiter
		}
		
		// Reset if time window expired
		if time.Now().After(limiter.resetTime) {
			limiter.requests = 0
			limiter.resetTime = time.Now().Add(time.Minute)
		}
		
		limiter.requests++
		mu.Unlock()
		
		// Check if limit exceeded
		if limiter.requests > limit {
			ctx.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			ctx.Header().Set("X-RateLimit-Remaining", "0")
			ctx.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", limiter.resetTime.Unix()))
			
			ctx.JSON(429, zoox.H{
				"error": "Rate limit exceeded",
				"limit": limit,
			})
			ctx.Abort()
			return
		}
		
		// Set rate limit headers
		ctx.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		ctx.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-limiter.requests))
		
		ctx.Next()
	}
}

// Request/Response transformation middleware
func transformMiddleware() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		// Transform request
		if ctx.Method == "POST" {
			// Add timestamp to all POST requests
			ctx.Set("request_timestamp", time.Now().Format(time.RFC3339))
		}
		
		ctx.Next()
		
		// Transform response (add metadata)
		if ctx.Status() == 200 {
			ctx.Header().Set("X-Response-Time", time.Now().Format(time.RFC3339))
			ctx.Header().Set("X-Server", "Zoox-Tutorial")
		}
	}
}
```

## 🧪 Hands-on Exercise

### Exercise 1: Build a Complete Middleware Stack

Create a web application with the following middleware stack:

1. **Request logging** with unique request IDs
2. **Authentication** using JWT tokens
3. **Authorization** with role-based access control
4. **Rate limiting** with different limits per user role
5. **Response caching** for GET requests
6. **Error handling** with proper HTTP status codes

### Solution:

```go
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-zoox/zoox"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func main() {
	app := zoox.Default()

	// Middleware stack (order matters!)
	app.Use(requestLogger())
	app.Use(errorHandler())
	app.Use(authMiddleware())
	app.Use(roleBasedRateLimit())
	app.Use(cacheMiddleware())

	// Public routes (no auth required)
	app.Post("/login", loginHandler)

	// Protected routes
	api := app.Group("/api")
	{
		api.Use(requireAuth())
		
		api.Get("/profile", profileHandler)
		api.Get("/users", requireRole("admin"), usersHandler)
		api.Get("/data", dataHandler)
	}

	log.Println("🚀 Server starting on http://localhost:8080")
	log.Println("📋 Login with: POST /login {\"username\": \"admin\", \"password\": \"secret\"}")
	app.Run(":8080")
}

// Request logger with unique IDs
func requestLogger() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		// Generate unique request ID
		requestID := generateRequestID()
		ctx.Set("request_id", requestID)
		ctx.Header().Set("X-Request-ID", requestID)
		
		start := time.Now()
		log.Printf("[%s] %s %s %s - START", requestID, ctx.Method, ctx.Path, ctx.IP())
		
		ctx.Next()
		
		duration := time.Since(start)
		log.Printf("[%s] %s %s %s - %d - %v", requestID, ctx.Method, ctx.Path, ctx.IP(), ctx.Status(), duration)
	}
}

// Error handler
func errorHandler() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		ctx.Next()
		
		if len(ctx.Errors) > 0 {
			err := ctx.Errors[0]
			requestID := ctx.GetString("request_id")
			log.Printf("[%s] Error: %v", requestID, err)
			
			if !ctx.IsAborted() {
				ctx.JSON(500, zoox.H{
					"error":      "Internal server error",
					"request_id": requestID,
				})
			}
		}
	}
}

// Authentication middleware
func authMiddleware() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		// Skip auth for login endpoint
		if ctx.Path == "/login" {
			ctx.Next()
			return
		}
		
		ctx.Next()
	}
}

// Require authentication
func requireAuth() func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		token := ctx.Header().Get("Authorization")
		if token == "" {
			ctx.JSON(401, zoox.H{"error": "Missing authorization token"})
			ctx.Abort()
			return
		}
		
		// Remove "Bearer " prefix
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}
		
		// Validate token (simplified)
		user := validateToken(token)
		if user == nil {
			ctx.JSON(401, zoox.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}
		
		ctx.Set("user", user)
		ctx.Next()
	}
}

// Role-based access control
func requireRole(role string) func(*zoox.Context) {
	return func(ctx *zoox.Context) {
		user := ctx.Get("user").(*User)
		if user.Role != role {
			ctx.JSON(403, zoox.H{
				"error": "Insufficient permissions",
				"required_role": role,
				"user_role": user.Role,
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// Role-based rate limiting
func roleBasedRateLimit() func(*zoox.Context) {
	limiters := make(map[string]*rateLimiter)
	var mu sync.RWMutex
	
	type rateLimiter struct {
		requests  int
		resetTime time.Time
		limit     int
	}
	
	return func(ctx *zoox.Context) {
		// Skip for login
		if ctx.Path == "/login" {
			ctx.Next()
			return
		}
		
		user := ctx.Get("user")
		if user == nil {
			ctx.Next()
			return
		}
		
		u := user.(*User)
		
		// Different limits based on role
		var limit int
		switch u.Role {
		case "admin":
			limit = 1000 // 1000 requests per minute
		case "user":
			limit = 100 // 100 requests per minute
		default:
			limit = 10 // 10 requests per minute
		}
		
		key := u.ID
		
		mu.Lock()
		limiter, exists := limiters[key]
		if !exists {
			limiter = &rateLimiter{
				requests:  0,
				resetTime: time.Now().Add(time.Minute),
				limit:     limit,
			}
			limiters[key] = limiter
		}
		
		if time.Now().After(limiter.resetTime) {
			limiter.requests = 0
			limiter.resetTime = time.Now().Add(time.Minute)
		}
		
		limiter.requests++
		mu.Unlock()
		
		if limiter.requests > limit {
			ctx.JSON(429, zoox.H{
				"error": "Rate limit exceeded",
				"limit": limit,
				"role":  u.Role,
			})
			ctx.Abort()
			return
		}
		
		ctx.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		ctx.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-limiter.requests))
		
		ctx.Next()
	}
}

// Simple caching middleware
func cacheMiddleware() func(*zoox.Context) {
	cache := make(map[string]cacheItem)
	var mu sync.RWMutex
	
	type cacheItem struct {
		data      interface{}
		timestamp time.Time
		ttl       time.Duration
	}
	
	return func(ctx *zoox.Context) {
		if ctx.Method != "GET" {
			ctx.Next()
			return
		}
		
		key := ctx.Path
		mu.RLock()
		item, exists := cache[key]
		mu.RUnlock()
		
		if exists && time.Since(item.timestamp) < item.ttl {
			ctx.Header().Set("X-Cache", "HIT")
			ctx.JSON(200, item.data)
			return
		}
		
		ctx.Next()
		
		// Cache successful responses
		if ctx.Status() == 200 {
			mu.Lock()
			cache[key] = cacheItem{
				data:      zoox.H{"cached": true, "time": time.Now().Format(time.RFC3339)},
				timestamp: time.Now(),
				ttl:       30 * time.Second,
			}
			mu.Unlock()
		}
	}
}

// Handlers
func loginHandler(ctx *zoox.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := ctx.BindJSON(&credentials); err != nil {
		ctx.JSON(400, zoox.H{"error": "Invalid request"})
		return
	}
	
	// Validate credentials (simplified)
	if credentials.Username == "admin" && credentials.Password == "secret" {
		token := generateToken("admin")
		ctx.JSON(200, zoox.H{
			"token": token,
			"user": User{ID: "1", Name: "Admin", Role: "admin"},
		})
		return
	}
	
	if credentials.Username == "user" && credentials.Password == "password" {
		token := generateToken("user")
		ctx.JSON(200, zoox.H{
			"token": token,
			"user": User{ID: "2", Name: "User", Role: "user"},
		})
		return
	}
	
	ctx.JSON(401, zoox.H{"error": "Invalid credentials"})
}

func profileHandler(ctx *zoox.Context) {
	user := ctx.Get("user").(*User)
	ctx.JSON(200, zoox.H{"user": user})
}

func usersHandler(ctx *zoox.Context) {
	users := []User{
		{ID: "1", Name: "Admin", Role: "admin"},
		{ID: "2", Name: "User", Role: "user"},
	}
	ctx.JSON(200, zoox.H{"users": users})
}

func dataHandler(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"data": "This is cached data",
		"time": time.Now().Format(time.RFC3339),
	})
}

// Helper functions
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateToken(role string) string {
	// Simplified token generation
	return fmt.Sprintf("token_%s_%d", role, time.Now().Unix())
}

func validateToken(token string) *User {
	// Simplified token validation
	if strings.HasPrefix(token, "token_admin_") {
		return &User{ID: "1", Name: "Admin", Role: "admin"}
	}
	if strings.HasPrefix(token, "token_user_") {
		return &User{ID: "2", Name: "User", Role: "user"}
	}
	return nil
}
```

## 📚 Key Takeaways

1. **Middleware Order**: The order in which middleware is added matters
2. **ctx.Next()**: Always call `ctx.Next()` to continue the chain
3. **ctx.Abort()**: Use `ctx.Abort()` to stop processing
4. **Error Handling**: Implement proper error handling in middleware
5. **Conditional Logic**: Middleware can be conditional based on request properties
6. **State Management**: Use `ctx.Set()` and `ctx.Get()` to share data between middleware
7. **Performance**: Be mindful of middleware performance impact

## 📖 Additional Resources

- [Middleware Design Patterns](https://en.wikipedia.org/wiki/Middleware)
- [HTTP Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication)
- [Rate Limiting Strategies](https://en.wikipedia.org/wiki/Rate_limiting)
- [Next Tutorial: Authentication & Authorization](./05-authentication-authorization.md)

## 🔗 What's Next?

In the next tutorial, we'll dive deeper into authentication and authorization, learning how to:
- Implement JWT authentication
- Create role-based access control
- Handle session management
- Secure API endpoints

Continue to [Tutorial 05: Authentication & Authorization](./05-authentication-authorization.md)! 