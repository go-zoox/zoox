# Tutorial 05: Advanced Routing

## 📖 Overview

In this tutorial, we'll explore advanced routing techniques in Zoox that go beyond basic route definitions. You'll learn about dynamic route registration, route constraints, performance optimization, and advanced patterns for complex applications.

## 🎯 Learning Objectives

By the end of this tutorial, you will be able to:
- Implement dynamic route registration
- Use route constraints for validation
- Optimize routing performance
- Handle complex routing scenarios
- Build flexible and maintainable routing systems

## 📋 Prerequisites

- Completed [Tutorial 02: Routing Fundamentals](./02-routing-fundamentals.md)
- Basic understanding of HTTP methods and RESTful APIs
- Familiarity with Go programming

## 🚀 Getting Started

### Dynamic Route Registration

Dynamic route registration allows you to register routes at runtime based on configuration, database content, or other dynamic sources.

```go
package main

import (
    "fmt"
    "log"
    "strconv"
    "strings"
    
    "github.com/go-zoox/zoox"
)

// RouteConfig represents a dynamic route configuration
type RouteConfig struct {
    Method  string `json:"method"`
    Path    string `json:"path"`
    Handler string `json:"handler"`
    Params  map[string]string `json:"params"`
}

// DynamicRouter manages dynamic routes
type DynamicRouter struct {
    app    *zoox.Application
    routes map[string]RouteConfig
}

// NewDynamicRouter creates a new dynamic router
func NewDynamicRouter(app *zoox.Application) *DynamicRouter {
    return &DynamicRouter{
        app:    app,
        routes: make(map[string]RouteConfig),
    }
}

// RegisterRoute registers a route dynamically
func (dr *DynamicRouter) RegisterRoute(id string, config RouteConfig) error {
    handler := dr.createHandler(config)
    
    switch strings.ToUpper(config.Method) {
    case "GET":
        dr.app.Get(config.Path, handler)
    case "POST":
        dr.app.Post(config.Path, handler)
    case "PUT":
        dr.app.Put(config.Path, handler)
    case "DELETE":
        dr.app.Delete(config.Path, handler)
    default:
        return fmt.Errorf("unsupported method: %s", config.Method)
    }
    
    dr.routes[id] = config
    return nil
}

// createHandler creates a handler function based on configuration
func (dr *DynamicRouter) createHandler(config RouteConfig) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        switch config.Handler {
        case "echo":
            dr.echoHandler(ctx, config)
        case "static":
            dr.staticHandler(ctx, config)
        case "redirect":
            dr.redirectHandler(ctx, config)
        case "template":
            dr.templateHandler(ctx, config)
        default:
            ctx.JSON(500, map[string]interface{}{
                "error": "Unknown handler type",
                "handler": config.Handler,
            })
        }
    }
}

// Handler implementations
func (dr *DynamicRouter) echoHandler(ctx *zoox.Context, config RouteConfig) {
    response := map[string]interface{}{
        "message": "Dynamic route response",
        "route":   config.Path,
        "method":  config.Method,
        "params":  ctx.Params(),
        "query":   ctx.Query(),
    }
    
    if message, ok := config.Params["message"]; ok {
        response["message"] = message
    }
    
    ctx.JSON(200, response)
}

func (dr *DynamicRouter) staticHandler(ctx *zoox.Context, config RouteConfig) {
    if content, ok := config.Params["content"]; ok {
        ctx.String(200, content)
    } else {
        ctx.String(200, "Static content")
    }
}

func (dr *DynamicRouter) redirectHandler(ctx *zoox.Context, config RouteConfig) {
    if url, ok := config.Params["url"]; ok {
        ctx.Redirect(302, url)
    } else {
        ctx.String(400, "Redirect URL not specified")
    }
}

func (dr *DynamicRouter) templateHandler(ctx *zoox.Context, config RouteConfig) {
    template := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Dynamic Template</title>
    </head>
    <body>
        <h1>Dynamic Route: {{.Path}}</h1>
        <p>Method: {{.Method}}</p>
        <p>Handler: {{.Handler}}</p>
    </body>
    </html>
    `
    
    ctx.HTML(200, template, config)
}

func main() {
    app := zoox.New()
    
    // Create dynamic router
    dynamicRouter := NewDynamicRouter(app)
    
    // Register some dynamic routes
    routes := []struct {
        id     string
        config RouteConfig
    }{
        {
            id: "welcome",
            config: RouteConfig{
                Method:  "GET",
                Path:    "/welcome/:name",
                Handler: "echo",
                Params: map[string]string{
                    "message": "Welcome to our dynamic API!",
                },
            },
        },
        {
            id: "status",
            config: RouteConfig{
                Method:  "GET",
                Path:    "/status",
                Handler: "static",
                Params: map[string]string{
                    "content": "Service is running",
                },
            },
        },
        {
            id: "home_redirect",
            config: RouteConfig{
                Method:  "GET",
                Path:    "/home",
                Handler: "redirect",
                Params: map[string]string{
                    "url": "/welcome/guest",
                },
            },
        },
    }
    
    for _, route := range routes {
        if err := dynamicRouter.RegisterRoute(route.id, route.config); err != nil {
            log.Printf("Error registering route %s: %v", route.id, err)
        }
    }
    
    // Admin endpoint to manage dynamic routes
    app.Post("/admin/routes", func(ctx *zoox.Context) {
        var config RouteConfig
        if err := ctx.BindJSON(&config); err != nil {
            ctx.JSON(400, map[string]string{"error": err.Error()})
            return
        }
        
        id := ctx.Query("id")
        if id == "" {
            ctx.JSON(400, map[string]string{"error": "Route ID is required"})
            return
        }
        
        if err := dynamicRouter.RegisterRoute(id, config); err != nil {
            ctx.JSON(500, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.JSON(200, map[string]string{
            "message": "Route registered successfully",
            "id":      id,
        })
    })
    
    // List all dynamic routes
    app.Get("/admin/routes", func(ctx *zoox.Context) {
        ctx.JSON(200, dynamicRouter.routes)
    })
    
    app.Listen(":8080")
}
```

### Route Constraints

Route constraints allow you to validate route parameters before the handler is executed.

```go
package main

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
    
    "github.com/go-zoox/zoox"
)

// Constraint represents a route parameter constraint
type Constraint interface {
    Validate(value string) bool
    Name() string
}

// IntConstraint validates integer parameters
type IntConstraint struct {
    Min, Max int
}

func (c IntConstraint) Validate(value string) bool {
    if val, err := strconv.Atoi(value); err == nil {
        return val >= c.Min && val <= c.Max
    }
    return false
}

func (c IntConstraint) Name() string {
    return fmt.Sprintf("int(%d-%d)", c.Min, c.Max)
}

// RegexConstraint validates parameters using regex
type RegexConstraint struct {
    Pattern *regexp.Regexp
    Name_   string
}

func (c RegexConstraint) Validate(value string) bool {
    return c.Pattern.MatchString(value)
}

func (c RegexConstraint) Name() string {
    return c.Name_
}

// EnumConstraint validates parameters against a set of allowed values
type EnumConstraint struct {
    Values []string
}

func (c EnumConstraint) Validate(value string) bool {
    for _, v := range c.Values {
        if v == value {
            return true
        }
    }
    return false
}

func (c EnumConstraint) Name() string {
    return fmt.Sprintf("enum(%s)", strings.Join(c.Values, "|"))
}

// ConstraintValidator manages route constraints
type ConstraintValidator struct {
    constraints map[string]Constraint
}

func NewConstraintValidator() *ConstraintValidator {
    return &ConstraintValidator{
        constraints: make(map[string]Constraint),
    }
}

func (cv *ConstraintValidator) AddConstraint(param string, constraint Constraint) {
    cv.constraints[param] = constraint
}

func (cv *ConstraintValidator) ValidateMiddleware() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        params := ctx.Params()
        
        for param, constraint := range cv.constraints {
            if value, exists := params[param]; exists {
                if !constraint.Validate(value) {
                    ctx.JSON(400, map[string]interface{}{
                        "error": fmt.Sprintf("Invalid parameter '%s': %s", param, value),
                        "constraint": constraint.Name(),
                        "value": value,
                    })
                    return
                }
            }
        }
        
        ctx.Next()
    }
}

func main() {
    app := zoox.New()
    
    // Create constraint validator
    validator := NewConstraintValidator()
    
    // Define constraints
    validator.AddConstraint("id", IntConstraint{Min: 1, Max: 1000})
    validator.AddConstraint("category", EnumConstraint{Values: []string{"tech", "science", "art", "music"}})
    validator.AddConstraint("slug", RegexConstraint{
        Pattern: regexp.MustCompile(`^[a-z0-9-]+$`),
        Name_:   "slug",
    })
    
    // Routes with constraints
    api := app.Group("/api/v1")
    api.Use(validator.ValidateMiddleware())
    
    // User routes with ID constraint
    api.Get("/users/:id", func(ctx *zoox.Context) {
        id := ctx.Param("id")
        ctx.JSON(200, map[string]interface{}{
            "message": "User found",
            "id":      id,
        })
    })
    
    // Article routes with category and slug constraints
    api.Get("/articles/:category/:slug", func(ctx *zoox.Context) {
        category := ctx.Param("category")
        slug := ctx.Param("slug")
        
        ctx.JSON(200, map[string]interface{}{
            "message":  "Article found",
            "category": category,
            "slug":     slug,
        })
    })
    
    // Product routes with multiple constraints
    api.Get("/products/:category/:id", func(ctx *zoox.Context) {
        category := ctx.Param("category")
        id := ctx.Param("id")
        
        ctx.JSON(200, map[string]interface{}{
            "message":  "Product found",
            "category": category,
            "id":       id,
        })
    })
    
    // Constraint documentation endpoint
    app.Get("/constraints", func(ctx *zoox.Context) {
        constraints := make(map[string]string)
        for param, constraint := range validator.constraints {
            constraints[param] = constraint.Name()
        }
        
        ctx.JSON(200, map[string]interface{}{
            "constraints": constraints,
            "examples": map[string]interface{}{
                "valid_urls": []string{
                    "/api/v1/users/123",
                    "/api/v1/articles/tech/my-article",
                    "/api/v1/products/tech/456",
                },
                "invalid_urls": []string{
                    "/api/v1/users/0",      // ID too low
                    "/api/v1/users/1001",   // ID too high
                    "/api/v1/articles/invalid-category/my-article", // Invalid category
                    "/api/v1/articles/tech/My Article",             // Invalid slug
                },
            },
        })
    })
    
    app.Listen(":8080")
}
```

### Route Caching and Performance Optimization

```go
package main

import (
    "crypto/md5"
    "fmt"
    "sync"
    "time"
    
    "github.com/go-zoox/zoox"
)

// CacheEntry represents a cached route response
type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
    Headers   map[string]string
}

// RouteCache manages route response caching
type RouteCache struct {
    cache map[string]CacheEntry
    mutex sync.RWMutex
    ttl   time.Duration
}

func NewRouteCache(ttl time.Duration) *RouteCache {
    cache := &RouteCache{
        cache: make(map[string]CacheEntry),
        ttl:   ttl,
    }
    
    // Start cleanup goroutine
    go cache.cleanup()
    
    return cache
}

func (rc *RouteCache) cleanup() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        rc.mutex.Lock()
        now := time.Now()
        for key, entry := range rc.cache {
            if now.After(entry.ExpiresAt) {
                delete(rc.cache, key)
            }
        }
        rc.mutex.Unlock()
    }
}

func (rc *RouteCache) generateKey(method, path string, params map[string]string) string {
    key := fmt.Sprintf("%s:%s", method, path)
    for k, v := range params {
        key += fmt.Sprintf(":%s=%s", k, v)
    }
    return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}

func (rc *RouteCache) Get(method, path string, params map[string]string) (interface{}, bool) {
    key := rc.generateKey(method, path, params)
    
    rc.mutex.RLock()
    defer rc.mutex.RUnlock()
    
    entry, exists := rc.cache[key]
    if !exists || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    
    return entry.Data, true
}

func (rc *RouteCache) Set(method, path string, params map[string]string, data interface{}, headers map[string]string) {
    key := rc.generateKey(method, path, params)
    
    rc.mutex.Lock()
    defer rc.mutex.Unlock()
    
    rc.cache[key] = CacheEntry{
        Data:      data,
        ExpiresAt: time.Now().Add(rc.ttl),
        Headers:   headers,
    }
}

func (rc *RouteCache) CacheMiddleware() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        // Only cache GET requests
        if ctx.Method() != "GET" {
            ctx.Next()
            return
        }
        
        path := ctx.Request.URL.Path
        params := ctx.Params()
        
        // Check cache
        if data, found := rc.Get(ctx.Method(), path, params); found {
            ctx.JSON(200, data)
            return
        }
        
        // Create a response recorder
        originalWriter := ctx.Writer
        recorder := &ResponseRecorder{
            original: originalWriter,
            data:     make(map[string]interface{}),
        }
        ctx.Writer = recorder
        
        // Process request
        ctx.Next()
        
        // Cache the response if it was successful
        if recorder.statusCode >= 200 && recorder.statusCode < 300 {
            rc.Set(ctx.Method(), path, params, recorder.data, recorder.headers)
        }
        
        // Restore original writer
        ctx.Writer = originalWriter
    }
}

// ResponseRecorder captures response data for caching
type ResponseRecorder struct {
    original   zoox.ResponseWriter
    data       map[string]interface{}
    statusCode int
    headers    map[string]string
}

func (rr *ResponseRecorder) Header() map[string][]string {
    return rr.original.Header()
}

func (rr *ResponseRecorder) Write(data []byte) (int, error) {
    return rr.original.Write(data)
}

func (rr *ResponseRecorder) WriteHeader(statusCode int) {
    rr.statusCode = statusCode
    rr.original.WriteHeader(statusCode)
}

// Route Pool for performance optimization
type RoutePool struct {
    handlers sync.Map
    stats    map[string]*RouteStats
    mutex    sync.RWMutex
}

type RouteStats struct {
    Hits          int64
    TotalDuration time.Duration
    AvgDuration   time.Duration
    LastAccess    time.Time
}

func NewRoutePool() *RoutePool {
    return &RoutePool{
        stats: make(map[string]*RouteStats),
    }
}

func (rp *RoutePool) TrackRoute(path string, handler zoox.HandlerFunc) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        start := time.Now()
        
        // Execute handler
        handler(ctx)
        
        // Update statistics
        duration := time.Since(start)
        rp.updateStats(path, duration)
    }
}

func (rp *RoutePool) updateStats(path string, duration time.Duration) {
    rp.mutex.Lock()
    defer rp.mutex.Unlock()
    
    stats, exists := rp.stats[path]
    if !exists {
        stats = &RouteStats{}
        rp.stats[path] = stats
    }
    
    stats.Hits++
    stats.TotalDuration += duration
    stats.AvgDuration = time.Duration(int64(stats.TotalDuration) / stats.Hits)
    stats.LastAccess = time.Now()
}

func (rp *RoutePool) GetStats() map[string]*RouteStats {
    rp.mutex.RLock()
    defer rp.mutex.RUnlock()
    
    // Create a copy to avoid race conditions
    result := make(map[string]*RouteStats)
    for path, stats := range rp.stats {
        result[path] = &RouteStats{
            Hits:          stats.Hits,
            TotalDuration: stats.TotalDuration,
            AvgDuration:   stats.AvgDuration,
            LastAccess:    stats.LastAccess,
        }
    }
    
    return result
}

func main() {
    app := zoox.New()
    
    // Create cache and route pool
    cache := NewRouteCache(5 * time.Minute)
    pool := NewRoutePool()
    
    // Apply caching middleware
    app.Use(cache.CacheMiddleware())
    
    // Cached routes
    app.Get("/products", pool.TrackRoute("/products", func(ctx *zoox.Context) {
        // Simulate expensive operation
        time.Sleep(100 * time.Millisecond)
        
        products := []map[string]interface{}{
            {"id": 1, "name": "Laptop", "price": 999.99},
            {"id": 2, "name": "Mouse", "price": 29.99},
            {"id": 3, "name": "Keyboard", "price": 79.99},
        }
        
        ctx.JSON(200, map[string]interface{}{
            "products": products,
            "cached":   false,
            "timestamp": time.Now().Unix(),
        })
    }))
    
    app.Get("/users/:id", pool.TrackRoute("/users/:id", func(ctx *zoox.Context) {
        id := ctx.Param("id")
        
        // Simulate database lookup
        time.Sleep(50 * time.Millisecond)
        
        ctx.JSON(200, map[string]interface{}{
            "user": map[string]interface{}{
                "id":   id,
                "name": fmt.Sprintf("User %s", id),
                "email": fmt.Sprintf("user%s@example.com", id),
            },
            "cached":   false,
            "timestamp": time.Now().Unix(),
        })
    }))
    
    // Performance statistics endpoint
    app.Get("/admin/stats", func(ctx *zoox.Context) {
        stats := pool.GetStats()
        
        ctx.JSON(200, map[string]interface{}{
            "route_stats": stats,
            "cache_info": map[string]interface{}{
                "ttl_minutes": int(cache.ttl.Minutes()),
                "entries":     len(cache.cache),
            },
        })
    })
    
    // Cache management endpoints
    app.Delete("/admin/cache", func(ctx *zoox.Context) {
        cache.mutex.Lock()
        cache.cache = make(map[string]CacheEntry)
        cache.mutex.Unlock()
        
        ctx.JSON(200, map[string]string{
            "message": "Cache cleared successfully",
        })
    })
    
    app.Listen(":8080")
}
```

## 🎯 Hands-on Exercise

Create a dynamic content management system with the following features:

### Requirements

1. **Dynamic Route Registration**: Allow administrators to create custom routes
2. **Route Constraints**: Validate route parameters
3. **Performance Optimization**: Implement caching and route statistics
4. **Content Management**: CRUD operations for dynamic content

### Solution

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "time"
    
    "github.com/go-zoox/zoox"
)

// Content represents dynamic content
type Content struct {
    ID          int                    `json:"id"`
    Title       string                 `json:"title"`
    Body        string                 `json:"body"`
    Type        string                 `json:"type"`
    Status      string                 `json:"status"`
    Metadata    map[string]interface{} `json:"metadata"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// DynamicRoute represents a dynamic route configuration
type DynamicRoute struct {
    ID          string            `json:"id"`
    Method      string            `json:"method"`
    Path        string            `json:"path"`
    ContentType string            `json:"content_type"`
    Template    string            `json:"template"`
    Constraints map[string]string `json:"constraints"`
    CacheTTL    int               `json:"cache_ttl"`
    CreatedAt   time.Time         `json:"created_at"`
}

// CMS manages the dynamic content management system
type CMS struct {
    app           *zoox.Application
    contents      map[int]*Content
    routes        map[string]*DynamicRoute
    cache         *RouteCache
    pool          *RoutePool
    validator     *ConstraintValidator
    mutex         sync.RWMutex
    nextContentID int
}

func NewCMS(app *zoox.Application) *CMS {
    return &CMS{
        app:           app,
        contents:      make(map[int]*Content),
        routes:        make(map[string]*DynamicRoute),
        cache:         NewRouteCache(5 * time.Minute),
        pool:          NewRoutePool(),
        validator:     NewConstraintValidator(),
        nextContentID: 1,
    }
}

func (cms *CMS) Setup() {
    // Apply middleware
    cms.app.Use(cms.cache.CacheMiddleware())
    cms.app.Use(cms.validator.ValidateMiddleware())
    
    // Setup default constraints
    cms.validator.AddConstraint("id", IntConstraint{Min: 1, Max: 999999})
    cms.validator.AddConstraint("slug", RegexConstraint{
        Pattern: regexp.MustCompile(`^[a-z0-9-]+$`),
        Name_:   "slug",
    })
    
    // Setup API routes
    cms.setupContentAPI()
    cms.setupRouteAPI()
    cms.setupAdminAPI()
}

func (cms *CMS) setupContentAPI() {
    api := cms.app.Group("/api/content")
    
    // Create content
    api.Post("/", func(ctx *zoox.Context) {
        var content Content
        if err := ctx.BindJSON(&content); err != nil {
            ctx.JSON(400, map[string]string{"error": err.Error()})
            return
        }
        
        cms.mutex.Lock()
        content.ID = cms.nextContentID
        cms.nextContentID++
        content.CreatedAt = time.Now()
        content.UpdatedAt = time.Now()
        cms.contents[content.ID] = &content
        cms.mutex.Unlock()
        
        ctx.JSON(201, content)
    })
    
    // Get content
    api.Get("/:id", func(ctx *zoox.Context) {
        id, _ := strconv.Atoi(ctx.Param("id"))
        
        cms.mutex.RLock()
        content, exists := cms.contents[id]
        cms.mutex.RUnlock()
        
        if !exists {
            ctx.JSON(404, map[string]string{"error": "Content not found"})
            return
        }
        
        ctx.JSON(200, content)
    })
    
    // Update content
    api.Put("/:id", func(ctx *zoox.Context) {
        id, _ := strconv.Atoi(ctx.Param("id"))
        
        cms.mutex.Lock()
        content, exists := cms.contents[id]
        if !exists {
            cms.mutex.Unlock()
            ctx.JSON(404, map[string]string{"error": "Content not found"})
            return
        }
        
        var updates Content
        if err := ctx.BindJSON(&updates); err != nil {
            cms.mutex.Unlock()
            ctx.JSON(400, map[string]string{"error": err.Error()})
            return
        }
        
        content.Title = updates.Title
        content.Body = updates.Body
        content.Type = updates.Type
        content.Status = updates.Status
        content.Metadata = updates.Metadata
        content.UpdatedAt = time.Now()
        cms.mutex.Unlock()
        
        ctx.JSON(200, content)
    })
    
    // Delete content
    api.Delete("/:id", func(ctx *zoox.Context) {
        id, _ := strconv.Atoi(ctx.Param("id"))
        
        cms.mutex.Lock()
        delete(cms.contents, id)
        cms.mutex.Unlock()
        
        ctx.JSON(200, map[string]string{"message": "Content deleted"})
    })
    
    // List content
    api.Get("/", func(ctx *zoox.Context) {
        cms.mutex.RLock()
        contents := make([]*Content, 0, len(cms.contents))
        for _, content := range cms.contents {
            contents = append(contents, content)
        }
        cms.mutex.RUnlock()
        
        ctx.JSON(200, map[string]interface{}{
            "contents": contents,
            "total":    len(contents),
        })
    })
}

func (cms *CMS) setupRouteAPI() {
    api := cms.app.Group("/api/routes")
    
    // Create dynamic route
    api.Post("/", func(ctx *zoox.Context) {
        var route DynamicRoute
        if err := ctx.BindJSON(&route); err != nil {
            ctx.JSON(400, map[string]string{"error": err.Error()})
            return
        }
        
        route.CreatedAt = time.Now()
        
        if err := cms.registerDynamicRoute(&route); err != nil {
            ctx.JSON(500, map[string]string{"error": err.Error()})
            return
        }
        
        cms.mutex.Lock()
        cms.routes[route.ID] = &route
        cms.mutex.Unlock()
        
        ctx.JSON(201, route)
    })
    
    // List dynamic routes
    api.Get("/", func(ctx *zoox.Context) {
        cms.mutex.RLock()
        routes := make([]*DynamicRoute, 0, len(cms.routes))
        for _, route := range cms.routes {
            routes = append(routes, route)
        }
        cms.mutex.RUnlock()
        
        ctx.JSON(200, map[string]interface{}{
            "routes": routes,
            "total":  len(routes),
        })
    })
    
    // Delete dynamic route
    api.Delete("/:id", func(ctx *zoox.Context) {
        id := ctx.Param("id")
        
        cms.mutex.Lock()
        delete(cms.routes, id)
        cms.mutex.Unlock()
        
        ctx.JSON(200, map[string]string{"message": "Route deleted"})
    })
}

func (cms *CMS) setupAdminAPI() {
    admin := cms.app.Group("/admin")
    
    // Performance statistics
    admin.Get("/stats", func(ctx *zoox.Context) {
        stats := cms.pool.GetStats()
        
        ctx.JSON(200, map[string]interface{}{
            "route_stats": stats,
            "cache_info": map[string]interface{}{
                "ttl_minutes": int(cms.cache.ttl.Minutes()),
                "entries":     len(cms.cache.cache),
            },
            "content_count": len(cms.contents),
            "route_count":   len(cms.routes),
        })
    })
    
    // Clear cache
    admin.Delete("/cache", func(ctx *zoox.Context) {
        cms.cache.mutex.Lock()
        cms.cache.cache = make(map[string]CacheEntry)
        cms.cache.mutex.Unlock()
        
        ctx.JSON(200, map[string]string{"message": "Cache cleared"})
    })
    
    // Dashboard
    admin.Get("/dashboard", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>CMS Dashboard</title>
            <style>
                body { font-family: Arial, sans-serif; margin: 20px; }
                .card { border: 1px solid #ddd; padding: 20px; margin: 10px 0; border-radius: 5px; }
                .stats { display: flex; gap: 20px; }
                .stat { flex: 1; text-align: center; }
                button { padding: 10px 20px; margin: 5px; }
            </style>
        </head>
        <body>
            <h1>CMS Dashboard</h1>
            
            <div class="card">
                <h2>Statistics</h2>
                <div class="stats">
                    <div class="stat">
                        <h3>Contents</h3>
                        <p id="content-count">Loading...</p>
                    </div>
                    <div class="stat">
                        <h2>Routes</h2>
                        <p id="route-count">Loading...</p>
                    </div>
                    <div class="stat">
                        <h3>Cache Entries</h3>
                        <p id="cache-count">Loading...</p>
                    </div>
                </div>
            </div>
            
            <div class="card">
                <h2>Actions</h2>
                <button onclick="clearCache()">Clear Cache</button>
                <button onclick="refreshStats()">Refresh Stats</button>
            </div>
            
            <script>
                async function refreshStats() {
                    try {
                        const response = await fetch('/admin/stats');
                        const data = await response.json();
                        
                        document.getElementById('content-count').textContent = data.content_count;
                        document.getElementById('route-count').textContent = data.route_count;
                        document.getElementById('cache-count').textContent = data.cache_info.entries;
                    } catch (error) {
                        console.error('Error fetching stats:', error);
                    }
                }
                
                async function clearCache() {
                    try {
                        await fetch('/admin/cache', { method: 'DELETE' });
                        alert('Cache cleared successfully');
                        refreshStats();
                    } catch (error) {
                        console.error('Error clearing cache:', error);
                    }
                }
                
                // Load initial stats
                refreshStats();
            </script>
        </body>
        </html>
        `
        
        ctx.HTML(200, html, nil)
    })
}

func (cms *CMS) registerDynamicRoute(route *DynamicRoute) error {
    handler := cms.createDynamicHandler(route)
    
    switch strings.ToUpper(route.Method) {
    case "GET":
        cms.app.Get(route.Path, cms.pool.TrackRoute(route.Path, handler))
    case "POST":
        cms.app.Post(route.Path, cms.pool.TrackRoute(route.Path, handler))
    case "PUT":
        cms.app.Put(route.Path, cms.pool.TrackRoute(route.Path, handler))
    case "DELETE":
        cms.app.Delete(route.Path, cms.pool.TrackRoute(route.Path, handler))
    default:
        return fmt.Errorf("unsupported method: %s", route.Method)
    }
    
    return nil
}

func (cms *CMS) createDynamicHandler(route *DynamicRoute) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        switch route.ContentType {
        case "json":
            cms.handleJSONContent(ctx, route)
        case "html":
            cms.handleHTMLContent(ctx, route)
        case "text":
            cms.handleTextContent(ctx, route)
        default:
            ctx.JSON(500, map[string]string{"error": "Unknown content type"})
        }
    }
}

func (cms *CMS) handleJSONContent(ctx *zoox.Context, route *DynamicRoute) {
    // Find content by type or ID
    var content *Content
    if id := ctx.Param("id"); id != "" {
        if contentID, err := strconv.Atoi(id); err == nil {
            cms.mutex.RLock()
            content = cms.contents[contentID]
            cms.mutex.RUnlock()
        }
    }
    
    if content == nil {
        ctx.JSON(404, map[string]string{"error": "Content not found"})
        return
    }
    
    ctx.JSON(200, content)
}

func (cms *CMS) handleHTMLContent(ctx *zoox.Context, route *DynamicRoute) {
    template := route.Template
    if template == "" {
        template = `
        <!DOCTYPE html>
        <html>
        <head>
            <title>{{.Title}}</title>
        </head>
        <body>
            <h1>{{.Title}}</h1>
            <div>{{.Body}}</div>
        </body>
        </html>
        `
    }
    
    var content *Content
    if id := ctx.Param("id"); id != "" {
        if contentID, err := strconv.Atoi(id); err == nil {
            cms.mutex.RLock()
            content = cms.contents[contentID]
            cms.mutex.RUnlock()
        }
    }
    
    if content == nil {
        ctx.HTML(404, "<h1>Content not found</h1>", nil)
        return
    }
    
    ctx.HTML(200, template, content)
}

func (cms *CMS) handleTextContent(ctx *zoox.Context, route *DynamicRoute) {
    var content *Content
    if id := ctx.Param("id"); id != "" {
        if contentID, err := strconv.Atoi(id); err == nil {
            cms.mutex.RLock()
            content = cms.contents[contentID]
            cms.mutex.RUnlock()
        }
    }
    
    if content == nil {
        ctx.String(404, "Content not found")
        return
    }
    
    ctx.String(200, content.Body)
}

func main() {
    app := zoox.New()
    
    // Create CMS
    cms := NewCMS(app)
    cms.Setup()
    
    // Create sample content
    sampleContent := []*Content{
        {
            ID:        1,
            Title:     "Welcome to Our Site",
            Body:      "This is the welcome page content.",
            Type:      "page",
            Status:    "published",
            Metadata:  map[string]interface{}{"featured": true},
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
        {
            ID:        2,
            Title:     "About Us",
            Body:      "Learn more about our company and mission.",
            Type:      "page",
            Status:    "published",
            Metadata:  map[string]interface{}{"menu_order": 1},
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }
    
    cms.mutex.Lock()
    for _, content := range sampleContent {
        cms.contents[content.ID] = content
    }
    cms.nextContentID = 3
    cms.mutex.Unlock()
    
    // Create sample routes
    sampleRoutes := []*DynamicRoute{
        {
            ID:          "welcome",
            Method:      "GET",
            Path:        "/welcome",
            ContentType: "html",
            Template:    "",
            Constraints: map[string]string{},
            CacheTTL:    300,
            CreatedAt:   time.Now(),
        },
        {
            ID:          "about",
            Method:      "GET",
            Path:        "/about",
            ContentType: "html",
            Template:    "",
            Constraints: map[string]string{},
            CacheTTL:    300,
            CreatedAt:   time.Now(),
        },
    }
    
    for _, route := range sampleRoutes {
        if err := cms.registerDynamicRoute(route); err != nil {
            log.Printf("Error registering route %s: %v", route.ID, err)
        } else {
            cms.mutex.Lock()
            cms.routes[route.ID] = route
            cms.mutex.Unlock()
        }
    }
    
    log.Println("CMS Server starting on :8080")
    log.Println("Dashboard: http://localhost:8080/admin/dashboard")
    log.Println("API Documentation: http://localhost:8080/api/content")
    
    app.Listen(":8080")
}
```

## 🔍 Testing Your Implementation

Test your CMS with these commands:

```bash
# Create content
curl -X POST http://localhost:8080/api/content \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Page", "body": "This is a test page", "type": "page", "status": "published"}'

# Create dynamic route
curl -X POST http://localhost:8080/api/routes \
  -H "Content-Type: application/json" \
  -d '{"id": "test", "method": "GET", "path": "/test/:id", "content_type": "json", "cache_ttl": 300}'

# Test the dynamic route
curl http://localhost:8080/test/1

# View dashboard
open http://localhost:8080/admin/dashboard
```

## 📚 Key Takeaways

1. **Dynamic Registration**: Routes can be registered at runtime based on configuration
2. **Route Constraints**: Parameter validation improves API reliability
3. **Performance Optimization**: Caching and route pooling enhance performance
4. **Flexible Architecture**: Dynamic systems require careful design and validation
5. **Monitoring**: Track route performance and cache effectiveness

## 🎯 Next Steps

- Explore [Tutorial 06: Template Engine](./06-template-engine.md) for advanced templating
- Learn about [Tutorial 08: WebSocket Development](./08-websocket-development.md) for real-time features
- Study [Tutorial 10: Authentication & Authorization](./10-authentication-authorization.md) for security

## 🤝 Need Help?

If you encounter any issues:
1. Check the [examples directory](../examples/) for working code
2. Review the [API documentation](../DOCUMENTATION.md)
3. Join our community discussions
4. Report bugs in the issue tracker

---

**Congratulations!** You've mastered advanced routing techniques in Zoox. You can now build flexible, high-performance applications with dynamic routing capabilities. 