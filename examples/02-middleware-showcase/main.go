package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	// Create Zoox app
	app := zoox.Default()

	// ===================
	// GLOBAL MIDDLEWARE
	// ===================
	
	// Request ID middleware - adds unique ID to each request
	app.Use(middleware.RequestID())
	
	// Logger middleware - logs all requests
	app.Use(middleware.Logger())
	
	// Recovery middleware - handles panics gracefully
	app.Use(middleware.Recovery())
	
	// CORS middleware - handles cross-origin requests
	app.Use(middleware.CORS(&middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ===================
	// DEMONSTRATION ROUTES
	// ===================
	
	// Home page with middleware overview
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Zoox Middleware Showcase",
			"request_id": ctx.Header("X-Request-ID"),
			"middleware_demos": map[string]string{
				"GET /basic":           "Basic middleware demo",
				"GET /security":        "Security middleware demo",
				"GET /performance":     "Performance middleware demo", 
				"GET /auth/basic":      "Basic authentication",
				"GET /auth/bearer":     "Bearer token authentication",
				"GET /auth/jwt":        "JWT authentication",
				"GET /custom":          "Custom middleware demo",
				"GET /rate-limited":    "Rate limiting demo",
				"GET /cached":          "Caching demo",
				"GET /compressed":      "Compression demo",
				"POST /panic":          "Recovery middleware demo",
			},
		})
	})

	// ===================
	// BASIC MIDDLEWARE GROUP
	// ===================
	
	basic := app.Group("/basic")
	basic.Use(func(ctx *zoox.Context) {
		fmt.Printf("Basic middleware executed for: %s\n", ctx.Path)
		ctx.Set("demo", "basic middleware")
		ctx.Next()
	})
	
	basic.Get("/", func(ctx *zoox.Context) {
		demo := ctx.Get("demo")
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Basic middleware demonstration",
			"demo_value": demo,
			"request_id": ctx.Header("X-Request-ID"),
			"timestamp": time.Now().Unix(),
		})
	})

	// ===================
	// SECURITY MIDDLEWARE GROUP
	// ===================
	
	security := app.Group("/security")
	
	// Helmet middleware - adds security headers
	security.Use(middleware.Helmet(&middleware.HelmetConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "no-referrer",
		ContentSecurityPolicy: "default-src 'self'",
	}))
	
	security.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Security headers applied",
			"headers": map[string]string{
				"X-XSS-Protection":           ctx.Header("X-XSS-Protection"),
				"X-Content-Type-Options":     ctx.Header("X-Content-Type-Options"),
				"X-Frame-Options":            ctx.Header("X-Frame-Options"),
				"Referrer-Policy":            ctx.Header("Referrer-Policy"),
				"Content-Security-Policy":    ctx.Header("Content-Security-Policy"),
			},
		})
	})

	// ===================
	// PERFORMANCE MIDDLEWARE GROUP
	// ===================
	
	performance := app.Group("/performance")
	
	// Gzip compression middleware
	performance.Use(middleware.Gzip())
	
	// Static file serving with cache headers
	performance.Use(middleware.StaticCache(&middleware.StaticCacheConfig{
		MaxAge: 24 * time.Hour,
	}))
	
	performance.Get("/", func(ctx *zoox.Context) {
		// Large response to demonstrate compression
		largeData := make([]string, 1000)
		for i := range largeData {
			largeData[i] = fmt.Sprintf("This is line %d with some repeated content that compresses well", i)
		}
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Performance optimized response",
			"compression": "gzip applied",
			"cache_headers": "static cache headers added",
			"large_data": largeData,
		})
	})

	// ===================
	// AUTHENTICATION MIDDLEWARE GROUPS
	// ===================
	
	// Basic Authentication
	basicAuth := app.Group("/auth/basic")
	basicAuth.Use(middleware.BasicAuth("Restricted Area", map[string]string{
		"admin": "secret123",
		"user":  "password",
	}))
	
	basicAuth.Get("/", func(ctx *zoox.Context) {
		user := ctx.Get("user")
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Basic authentication successful",
			"user": user,
			"auth_type": "basic",
		})
	})
	
	// Bearer Token Authentication
	bearerAuth := app.Group("/auth/bearer")
	bearerAuth.Use(middleware.BearerAuth(func(token string) (interface{}, error) {
		// Simple token validation (in production, verify with database/JWT)
		validTokens := map[string]string{
			"demo-token": "demo-user",
			"admin-token": "admin-user",
		}
		
		if user, valid := validTokens[token]; valid {
			return user, nil
		}
		return nil, fmt.Errorf("invalid token")
	}))
	
	bearerAuth.Get("/", func(ctx *zoox.Context) {
		user := ctx.Get("user")
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Bearer token authentication successful",
			"user": user,
			"auth_type": "bearer",
		})
	})
	
	// JWT Authentication
	jwtAuth := app.Group("/auth/jwt")
	jwtAuth.Use(middleware.JWT(&middleware.JWTConfig{
		Secret: "your-secret-key",
		ContextKey: "jwt_user",
	}))
	
	jwtAuth.Get("/", func(ctx *zoox.Context) {
		jwtUser := ctx.Get("jwt_user")
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "JWT authentication successful",
			"user": jwtUser,
			"auth_type": "jwt",
		})
	})

	// ===================
	// CUSTOM MIDDLEWARE GROUP
	// ===================
	
	custom := app.Group("/custom")
	
	// Custom request timing middleware
	custom.Use(func(ctx *zoox.Context) {
		start := time.Now()
		ctx.Set("start_time", start)
		
		// Execute next handlers
		ctx.Next()
		
		// Calculate and log response time
		duration := time.Since(start)
		ctx.Header("X-Response-Time", duration.String())
		fmt.Printf("Request %s took %v\n", ctx.Path, duration)
	})
	
	// Custom user agent logging middleware
	custom.Use(func(ctx *zoox.Context) {
		userAgent := ctx.Header("User-Agent")
		fmt.Printf("User-Agent: %s\n", userAgent)
		ctx.Set("user_agent", userAgent)
		ctx.Next()
	})
	
	custom.Get("/", func(ctx *zoox.Context) {
		startTime := ctx.Get("start_time")
		userAgent := ctx.Get("user_agent")
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Custom middleware demonstration",
			"start_time": startTime,
			"user_agent": userAgent,
			"response_time": ctx.Header("X-Response-Time"),
		})
	})

	// ===================
	// RATE LIMITING MIDDLEWARE
	// ===================
	
	rateLimited := app.Group("/rate-limited")
	rateLimited.Use(middleware.RateLimit(&middleware.RateLimitConfig{
		Rate:     2,             // 2 requests
		Burst:    5,             // burst of 5
		Duration: time.Minute,   // per minute
	}))
	
	rateLimited.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Rate limiting applied",
			"rate": "2 requests per minute",
			"burst": 5,
			"timestamp": time.Now().Unix(),
		})
	})

	// ===================
	// CACHING MIDDLEWARE
	// ===================
	
	cached := app.Group("/cached")
	cached.Use(middleware.Cache(&middleware.CacheConfig{
		TTL: 30 * time.Second,
		Key: func(ctx *zoox.Context) string {
			return fmt.Sprintf("cache:%s:%s", ctx.Method, ctx.Path)
		},
	}))
	
	cached.Get("/", func(ctx *zoox.Context) {
		// Simulate expensive operation
		time.Sleep(100 * time.Millisecond)
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Cached response",
			"cache_ttl": "30 seconds",
			"generated_at": time.Now().Format(time.RFC3339),
			"expensive_calculation": time.Now().UnixNano(),
		})
	})

	// ===================
	// COMPRESSION MIDDLEWARE
	// ===================
	
	compressed := app.Group("/compressed")
	compressed.Use(middleware.Gzip())
	
	compressed.Get("/", func(ctx *zoox.Context) {
		// Generate large, compressible content
		content := ""
		for i := 0; i < 1000; i++ {
			content += fmt.Sprintf("This is repeated content line %d that should compress very well with gzip compression middleware. ", i)
		}
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Large response with gzip compression",
			"original_size_hint": "Very large without compression",
			"compressed_size_hint": "Much smaller with gzip",
			"content": content,
		})
	})

	// ===================
	// ERROR HANDLING / RECOVERY
	// ===================
	
	app.Post("/panic", func(ctx *zoox.Context) {
		// This will trigger the recovery middleware
		panic("Intentional panic to demonstrate recovery middleware")
	})

	// ===================
	// MIDDLEWARE DOCUMENTATION
	// ===================
	
	app.Get("/middleware/docs", func(ctx *zoox.Context) {
		docs := map[string]interface{}{
			"title": "Zoox Middleware Documentation",
			"global_middleware": []map[string]string{
				{"name": "RequestID", "description": "Adds unique request ID"},
				{"name": "Logger", "description": "Logs all requests"},
				{"name": "Recovery", "description": "Handles panics gracefully"},
				{"name": "CORS", "description": "Handles cross-origin requests"},
			},
			"security_middleware": []map[string]string{
				{"name": "Helmet", "description": "Adds security headers"},
				{"name": "BasicAuth", "description": "HTTP Basic authentication"},
				{"name": "BearerAuth", "description": "Bearer token authentication"},
				{"name": "JWT", "description": "JWT token authentication"},
			},
			"performance_middleware": []map[string]string{
				{"name": "Gzip", "description": "Response compression"},
				{"name": "StaticCache", "description": "Static file caching"},
				{"name": "RateLimit", "description": "Request rate limiting"},
				{"name": "Cache", "description": "Response caching"},
			},
			"custom_middleware": []map[string]string{
				{"name": "RequestTiming", "description": "Measures request duration"},
				{"name": "UserAgentLogging", "description": "Logs user agent strings"},
			},
		}
		
		ctx.JSON(http.StatusOK, docs)
	})

	// ===================
	// MIDDLEWARE TESTING ENDPOINTS
	// ===================
	
	testing := app.Group("/test")
	
	// Test endpoint for header inspection
	testing.Get("/headers", func(ctx *zoox.Context) {
		headers := make(map[string]string)
		ctx.Request.Header.Range(func(key, value []byte) bool {
			headers[string(key)] = string(value)
			return true
		})
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "All request headers",
			"headers": headers,
		})
	})
	
	// Test endpoint for response timing
	testing.Get("/slow", func(ctx *zoox.Context) {
		// Simulate slow operation
		time.Sleep(2 * time.Second)
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Slow response for testing custom timing middleware",
			"delay": "2 seconds",
		})
	})

	// Start server
	fmt.Println("🚀 Middleware Showcase starting...")
	fmt.Println("📍 Server running on http://localhost:8080")
	fmt.Println("📚 Middleware Documentation: http://localhost:8080/middleware/docs")
	fmt.Println("🔍 Demo Endpoints:")
	fmt.Println("  • Basic: http://localhost:8080/basic")
	fmt.Println("  • Security: http://localhost:8080/security")
	fmt.Println("  • Performance: http://localhost:8080/performance")
	fmt.Println("  • Auth Basic: http://localhost:8080/auth/basic (admin:secret123)")
	fmt.Println("  • Auth Bearer: http://localhost:8080/auth/bearer (Authorization: Bearer demo-token)")
	fmt.Println("  • Custom: http://localhost:8080/custom")
	fmt.Println("  • Rate Limited: http://localhost:8080/rate-limited")
	fmt.Println("  • Cached: http://localhost:8080/cached")
	fmt.Println("  • Compressed: http://localhost:8080/compressed")
	
	app.Run(":8080")
} 