# Tutorial 15: Performance Optimization

## Overview
Learn essential performance optimization techniques for Zoox applications, including caching, connection pooling, middleware optimization, and monitoring performance metrics.

## Learning Objectives
- Implement effective caching strategies
- Optimize database connections and queries
- Use connection pooling and resource management
- Apply middleware optimization techniques
- Monitor and measure performance
- Implement rate limiting and throttling

## Prerequisites
- Complete Tutorial 14: Testing Strategies
- Understanding of Go performance concepts
- Basic knowledge of caching systems

## Caching Strategies

### Memory Caching

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "sync"
    "time"

    "github.com/go-zoox/zoox"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
    Data      interface{}
    ExpiresAt time.Time
}

// MemoryCache provides in-memory caching
type MemoryCache struct {
    items map[string]CacheItem
    mutex sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
    cache := &MemoryCache{
        items: make(map[string]CacheItem),
    }
    
    // Start cleanup goroutine
    go cache.cleanup()
    
    return cache
}

func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.items[key] = CacheItem{
        Data:      value,
        ExpiresAt: time.Now().Add(ttl),
    }
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    item, exists := c.items[key]
    if !exists || time.Now().After(item.ExpiresAt) {
        return nil, false
    }
    
    return item.Data, true
}

func (c *MemoryCache) Delete(key string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    delete(c.items, key)
}

func (c *MemoryCache) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mutex.Lock()
        now := time.Now()
        for key, item := range c.items {
            if now.After(item.ExpiresAt) {
                delete(c.items, key)
            }
        }
        c.mutex.Unlock()
    }
}

// Response caching middleware
func cacheMiddleware(cache *MemoryCache, ttl time.Duration) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        // Only cache GET requests
        if ctx.Method() != http.MethodGet {
            ctx.Next()
            return
        }
        
        cacheKey := ctx.Request().URL.Path + "?" + ctx.Request().URL.RawQuery
        
        // Try to get from cache
        if cached, found := cache.Get(cacheKey); found {
            ctx.JSON(http.StatusOK, cached)
            return
        }
        
        // Capture response
        originalWriter := ctx.Writer
        responseCapture := &responseWriter{
            ResponseWriter: originalWriter,
            body:          make([]byte, 0),
        }
        ctx.Writer = responseCapture
        
        ctx.Next()
        
        // Cache successful responses
        if responseCapture.statusCode == http.StatusOK && len(responseCapture.body) > 0 {
            var data interface{}
            if err := json.Unmarshal(responseCapture.body, &data); err == nil {
                cache.Set(cacheKey, data, ttl)
            }
        }
        
        // Write the actual response
        originalWriter.WriteHeader(responseCapture.statusCode)
        originalWriter.Write(responseCapture.body)
    }
}

type responseWriter struct {
    http.ResponseWriter
    body       []byte
    statusCode int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    rw.body = append(rw.body, b...)
    return len(b), nil
}

func (rw *responseWriter) WriteHeader(statusCode int) {
    rw.statusCode = statusCode
}

// User service with caching
type User struct {
    ID       int       `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Created  time.Time `json:"created"`
}

type UserService struct {
    users map[int]*User
    cache *MemoryCache
    mutex sync.RWMutex
}

func NewUserService(cache *MemoryCache) *UserService {
    return &UserService{
        users: make(map[int]*User),
        cache: cache,
    }
}

func (s *UserService) GetUser(id int) *User {
    // Try cache first
    cacheKey := fmt.Sprintf("user:%d", id)
    if cached, found := s.cache.Get(cacheKey); found {
        return cached.(*User)
    }
    
    // Get from "database"
    s.mutex.RLock()
    user := s.users[id]
    s.mutex.RUnlock()
    
    // Cache the result
    if user != nil {
        s.cache.Set(cacheKey, user, 10*time.Minute)
    }
    
    return user
}

func (s *UserService) CreateUser(name, email string) *User {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    id := len(s.users) + 1
    user := &User{
        ID:      id,
        Name:    name,
        Email:   email,
        Created: time.Now(),
    }
    
    s.users[id] = user
    
    // Cache the new user
    cacheKey := fmt.Sprintf("user:%d", id)
    s.cache.Set(cacheKey, user, 10*time.Minute)
    
    return user
}

func (s *UserService) UpdateUser(id int, name, email string) *User {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    user := s.users[id]
    if user == nil {
        return nil
    }
    
    user.Name = name
    user.Email = email
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%d", id)
    s.cache.Delete(cacheKey)
    
    // Cache the updated user
    s.cache.Set(cacheKey, user, 10*time.Minute)
    
    return user
}
```

## Connection Pooling

```go
// Database connection pool simulation
type DatabasePool struct {
    connections chan *Connection
    maxConns    int
}

type Connection struct {
    ID       int
    InUse    bool
    LastUsed time.Time
}

func NewDatabasePool(maxConns int) *DatabasePool {
    pool := &DatabasePool{
        connections: make(chan *Connection, maxConns),
        maxConns:    maxConns,
    }
    
    // Initialize connections
    for i := 0; i < maxConns; i++ {
        pool.connections <- &Connection{
            ID:       i + 1,
            InUse:    false,
            LastUsed: time.Now(),
        }
    }
    
    return pool
}

func (p *DatabasePool) GetConnection() (*Connection, error) {
    select {
    case conn := <-p.connections:
        conn.InUse = true
        conn.LastUsed = time.Now()
        return conn, nil
    case <-time.After(5 * time.Second):
        return nil, fmt.Errorf("connection pool timeout")
    }
}

func (p *DatabasePool) ReleaseConnection(conn *Connection) {
    conn.InUse = false
    conn.LastUsed = time.Now()
    
    select {
    case p.connections <- conn:
        // Connection returned to pool
    default:
        // Pool is full, connection will be discarded
    }
}

// Database service with connection pooling
type DatabaseService struct {
    pool *DatabasePool
}

func NewDatabaseService(maxConns int) *DatabaseService {
    return &DatabaseService{
        pool: NewDatabasePool(maxConns),
    }
}

func (db *DatabaseService) QueryUser(id int) (*User, error) {
    conn, err := db.pool.GetConnection()
    if err != nil {
        return nil, err
    }
    defer db.pool.ReleaseConnection(conn)
    
    // Simulate database query
    time.Sleep(10 * time.Millisecond)
    
    return &User{
        ID:      id,
        Name:    fmt.Sprintf("User %d", id),
        Email:   fmt.Sprintf("user%d@example.com", id),
        Created: time.Now(),
    }, nil
}
```

## Request Rate Limiting

```go
// Rate limiter implementation
type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.Mutex
    limit    int
    window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    rl := &RateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
    
    // Cleanup old entries
    go rl.cleanup()
    
    return rl
}

func (rl *RateLimiter) Allow(clientID string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    now := time.Now()
    cutoff := now.Add(-rl.window)
    
    // Clean old requests
    requests := rl.requests[clientID]
    var validRequests []time.Time
    for _, req := range requests {
        if req.After(cutoff) {
            validRequests = append(validRequests, req)
        }
    }
    
    // Check if limit exceeded
    if len(validRequests) >= rl.limit {
        rl.requests[clientID] = validRequests
        return false
    }
    
    // Add current request
    validRequests = append(validRequests, now)
    rl.requests[clientID] = validRequests
    
    return true
}

func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        rl.mutex.Lock()
        now := time.Now()
        cutoff := now.Add(-rl.window)
        
        for clientID, requests := range rl.requests {
            var validRequests []time.Time
            for _, req := range requests {
                if req.After(cutoff) {
                    validRequests = append(validRequests, req)
                }
            }
            
            if len(validRequests) == 0 {
                delete(rl.requests, clientID)
            } else {
                rl.requests[clientID] = validRequests
            }
        }
        rl.mutex.Unlock()
    }
}

// Rate limiting middleware
func rateLimitMiddleware(limiter *RateLimiter) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        clientID := ctx.ClientIP()
        
        if !limiter.Allow(clientID) {
            ctx.JSON(http.StatusTooManyRequests, map[string]string{
                "error": "Rate limit exceeded",
            })
            ctx.Abort()
            return
        }
        
        ctx.Next()
    }
}
```

## Response Compression

```go
// Compression middleware
func compressionMiddleware() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        // Check if client accepts gzip
        if !strings.Contains(ctx.Header("Accept-Encoding"), "gzip") {
            ctx.Next()
            return
        }
        
        // Create gzip writer
        ctx.Set("Content-Encoding", "gzip")
        
        originalWriter := ctx.Writer
        gzipWriter := gzip.NewWriter(originalWriter)
        defer gzipWriter.Close()
        
        // Replace the writer
        ctx.Writer = &gzipResponseWriter{
            ResponseWriter: originalWriter,
            gzipWriter:     gzipWriter,
        }
        
        ctx.Next()
    }
}

type gzipResponseWriter struct {
    http.ResponseWriter
    gzipWriter *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
    return w.gzipWriter.Write(b)
}
```

## Performance Monitoring

```go
// Performance monitoring middleware
type PerformanceMetrics struct {
    RequestCount    int64
    TotalDuration   time.Duration
    AverageDuration time.Duration
    MaxDuration     time.Duration
    MinDuration     time.Duration
    mutex           sync.Mutex
}

func NewPerformanceMetrics() *PerformanceMetrics {
    return &PerformanceMetrics{
        MinDuration: time.Hour, // Initialize with high value
    }
}

func (pm *PerformanceMetrics) Record(duration time.Duration) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    pm.RequestCount++
    pm.TotalDuration += duration
    pm.AverageDuration = pm.TotalDuration / time.Duration(pm.RequestCount)
    
    if duration > pm.MaxDuration {
        pm.MaxDuration = duration
    }
    
    if duration < pm.MinDuration {
        pm.MinDuration = duration
    }
}

func (pm *PerformanceMetrics) GetStats() map[string]interface{} {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    return map[string]interface{}{
        "request_count":    pm.RequestCount,
        "total_duration":   pm.TotalDuration.String(),
        "average_duration": pm.AverageDuration.String(),
        "max_duration":     pm.MaxDuration.String(),
        "min_duration":     pm.MinDuration.String(),
    }
}

func performanceMiddleware(metrics *PerformanceMetrics) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        start := time.Now()
        
        ctx.Next()
        
        duration := time.Since(start)
        metrics.Record(duration)
        
        // Add performance headers
        ctx.Set("X-Response-Time", duration.String())
    }
}
```

## Complete Example Application

```go
func main() {
    // Initialize components
    cache := NewMemoryCache()
    userService := NewUserService(cache)
    dbService := NewDatabaseService(10)
    rateLimiter := NewRateLimiter(100, time.Minute)
    metrics := NewPerformanceMetrics()
    
    app := zoox.New()
    
    // Apply middleware in order
    app.Use(performanceMiddleware(metrics))
    app.Use(rateLimitMiddleware(rateLimiter))
    app.Use(compressionMiddleware())
    app.Use(cacheMiddleware(cache, 5*time.Minute))
    
    // Routes
    app.Post("/users", func(ctx *zoox.Context) {
        var req struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }
        
        if err := ctx.BindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Invalid JSON",
            })
            return
        }
        
        user := userService.CreateUser(req.Name, req.Email)
        ctx.JSON(http.StatusCreated, user)
    })
    
    app.Get("/users/:id", func(ctx *zoox.Context) {
        id, err := strconv.Atoi(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Invalid user ID",
            })
            return
        }
        
        user := userService.GetUser(id)
        if user == nil {
            ctx.JSON(http.StatusNotFound, map[string]string{
                "error": "User not found",
            })
            return
        }
        
        ctx.JSON(http.StatusOK, user)
    })
    
    app.Get("/users/:id/db", func(ctx *zoox.Context) {
        id, err := strconv.Atoi(ctx.Param("id"))
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Invalid user ID",
            })
            return
        }
        
        user, err := dbService.QueryUser(id)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "error": err.Error(),
            })
            return
        }
        
        ctx.JSON(http.StatusOK, user)
    })
    
    app.Get("/metrics", func(ctx *zoox.Context) {
        ctx.JSON(http.StatusOK, metrics.GetStats())
    })
    
    // Health check endpoint
    app.Get("/health", func(ctx *zoox.Context) {
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "status":    "healthy",
            "timestamp": time.Now(),
            "uptime":    time.Since(startTime).String(),
        })
    })
    
    fmt.Println("High-performance server starting on :8080")
    log.Fatal(app.Listen(":8080"))
}

var startTime = time.Now()
```

## Optimization Best Practices

### 1. Memory Management
```go
// Use object pools for frequently allocated objects
var userPool = sync.Pool{
    New: func() interface{} {
        return &User{}
    },
}

func getUser() *User {
    return userPool.Get().(*User)
}

func putUser(user *User) {
    // Reset user fields
    user.ID = 0
    user.Name = ""
    user.Email = ""
    userPool.Put(user)
}
```

### 2. JSON Optimization
```go
// Pre-allocate JSON encoders
var jsonEncoderPool = sync.Pool{
    New: func() interface{} {
        return json.NewEncoder(nil)
    },
}

func writeJSON(w http.ResponseWriter, data interface{}) error {
    encoder := jsonEncoderPool.Get().(*json.Encoder)
    defer jsonEncoderPool.Put(encoder)
    
    encoder.Reset(w)
    return encoder.Encode(data)
}
```

### 3. Database Query Optimization
```go
// Use prepared statements
type PreparedQueries struct {
    getUserByID    *sql.Stmt
    createUser     *sql.Stmt
    updateUser     *sql.Stmt
}

func NewPreparedQueries(db *sql.DB) (*PreparedQueries, error) {
    getUserByID, err := db.Prepare("SELECT id, name, email FROM users WHERE id = ?")
    if err != nil {
        return nil, err
    }
    
    createUser, err := db.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
    if err != nil {
        return nil, err
    }
    
    updateUser, err := db.Prepare("UPDATE users SET name = ?, email = ? WHERE id = ?")
    if err != nil {
        return nil, err
    }
    
    return &PreparedQueries{
        getUserByID: getUserByID,
        createUser:  createUser,
        updateUser:  updateUser,
    }, nil
}
```

## Performance Testing

```go
// Load testing helper
func loadTest(url string, concurrent int, requests int) {
    var wg sync.WaitGroup
    start := time.Now()
    
    for i := 0; i < concurrent; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            client := &http.Client{
                Timeout: 10 * time.Second,
            }
            
            for j := 0; j < requests/concurrent; j++ {
                resp, err := client.Get(url)
                if err != nil {
                    fmt.Printf("Error: %v\n", err)
                    continue
                }
                resp.Body.Close()
            }
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    fmt.Printf("Completed %d requests in %v\n", requests, duration)
    fmt.Printf("Requests per second: %.2f\n", float64(requests)/duration.Seconds())
}
```

## Key Takeaways

1. **Caching**: Implement multi-layer caching (memory, Redis, CDN)
2. **Connection Pooling**: Reuse database connections effectively
3. **Rate Limiting**: Protect against abuse and ensure fair usage
4. **Compression**: Reduce bandwidth usage with gzip compression
5. **Monitoring**: Track performance metrics continuously
6. **Memory Management**: Use object pools and avoid memory leaks
7. **Database Optimization**: Use prepared statements and query optimization

## Next Steps

- Tutorial 16: Security Best Practices - Implement security measures
- Tutorial 17: Deployment Strategies - Deploy optimized applications
- Explore profiling tools (pprof)
- Learn about microservices optimization
- Study CDN and edge computing strategies

## Additional Resources

- [Go Performance Tips](https://golang.org/doc/effective_go.html#performance)
- [pprof Profiling](https://golang.org/pkg/runtime/pprof/)
- [Benchmarking in Go](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Memory Management](https://golang.org/doc/gc-guide) 