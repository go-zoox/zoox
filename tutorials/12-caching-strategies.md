# Tutorial 12: Caching Strategies

## 📖 Overview

Learn to implement effective caching strategies in Zoox applications for improved performance. This tutorial covers memory caching, Redis integration, cache invalidation, and performance optimization techniques.

## 🎯 Learning Objectives

- Implement memory caching
- Integrate Redis for distributed caching
- Design cache invalidation strategies
- Optimize application performance
- Handle cache-related patterns

## 📋 Prerequisites

- Completed [Tutorial 01: Getting Started](./01-getting-started.md)
- Understanding of caching concepts
- Basic knowledge of Redis (optional)

## 🚀 Getting Started

### Multi-Level Caching System

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"
    
    "github.com/go-redis/redis/v8"
    "github.com/go-zoox/zoox"
    "golang.org/x/net/context"
)

type CacheItem struct {
    Data      interface{}
    ExpiresAt time.Time
    Hits      int64
}

type MemoryCache struct {
    items map[string]CacheItem
    mutex sync.RWMutex
    ttl   time.Duration
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
    cache := &MemoryCache{
        items: make(map[string]CacheItem),
        ttl:   ttl,
    }
    
    // Start cleanup goroutine
    go cache.cleanup()
    
    return cache
}

func (mc *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    
    if ttl == 0 {
        ttl = mc.ttl
    }
    
    mc.items[key] = CacheItem{
        Data:      value,
        ExpiresAt: time.Now().Add(ttl),
        Hits:      0,
    }
}

func (mc *MemoryCache) Get(key string) (interface{}, bool) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    
    item, exists := mc.items[key]
    if !exists || time.Now().After(item.ExpiresAt) {
        delete(mc.items, key)
        return nil, false
    }
    
    // Update hit count
    item.Hits++
    mc.items[key] = item
    
    return item.Data, true
}

func (mc *MemoryCache) Delete(key string) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    delete(mc.items, key)
}

func (mc *MemoryCache) Clear() {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    mc.items = make(map[string]CacheItem)
}

func (mc *MemoryCache) Stats() map[string]interface{} {
    mc.mutex.RLock()
    defer mc.mutex.RUnlock()
    
    totalHits := int64(0)
    for _, item := range mc.items {
        totalHits += item.Hits
    }
    
    return map[string]interface{}{
        "entries":    len(mc.items),
        "total_hits": totalHits,
        "ttl_seconds": int(mc.ttl.Seconds()),
    }
}

func (mc *MemoryCache) cleanup() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        mc.mutex.Lock()
        now := time.Now()
        for key, item := range mc.items {
            if now.After(item.ExpiresAt) {
                delete(mc.items, key)
            }
        }
        mc.mutex.Unlock()
    }
}

// Redis Cache
type RedisCache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewRedisCache(addr, password string, db int, ttl time.Duration) *RedisCache {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    
    return &RedisCache{
        client: client,
        ttl:    ttl,
    }
}

func (rc *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
    if ttl == 0 {
        ttl = rc.ttl
    }
    
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return rc.client.Set(context.Background(), key, data, ttl).Err()
}

func (rc *RedisCache) Get(key string) (interface{}, bool) {
    data, err := rc.client.Get(context.Background(), key).Result()
    if err != nil {
        return nil, false
    }
    
    var value interface{}
    if err := json.Unmarshal([]byte(data), &value); err != nil {
        return nil, false
    }
    
    return value, true
}

func (rc *RedisCache) Delete(key string) error {
    return rc.client.Del(context.Background(), key).Err()
}

func (rc *RedisCache) Clear() error {
    return rc.client.FlushDB(context.Background()).Err()
}

// Multi-level cache manager
type CacheManager struct {
    l1Cache *MemoryCache
    l2Cache *RedisCache
    enabled bool
}

func NewCacheManager(l1TTL, l2TTL time.Duration) *CacheManager {
    return &CacheManager{
        l1Cache: NewMemoryCache(l1TTL),
        l2Cache: NewRedisCache("localhost:6379", "", 0, l2TTL),
        enabled: true,
    }
}

func (cm *CacheManager) Get(key string) (interface{}, bool) {
    if !cm.enabled {
        return nil, false
    }
    
    // Try L1 cache first
    if value, found := cm.l1Cache.Get(key); found {
        return value, true
    }
    
    // Try L2 cache
    if cm.l2Cache != nil {
        if value, found := cm.l2Cache.Get(key); found {
            // Store in L1 cache for faster access
            cm.l1Cache.Set(key, value, 0)
            return value, true
        }
    }
    
    return nil, false
}

func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
    if !cm.enabled {
        return
    }
    
    // Store in L1 cache
    cm.l1Cache.Set(key, value, ttl)
    
    // Store in L2 cache
    if cm.l2Cache != nil {
        cm.l2Cache.Set(key, value, ttl)
    }
}

func (cm *CacheManager) Delete(key string) {
    cm.l1Cache.Delete(key)
    if cm.l2Cache != nil {
        cm.l2Cache.Delete(key)
    }
}

func (cm *CacheManager) Stats() map[string]interface{} {
    return map[string]interface{}{
        "l1_cache": cm.l1Cache.Stats(),
        "enabled":  cm.enabled,
    }
}

// Cache middleware
func (cm *CacheManager) CacheMiddleware(ttl time.Duration) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        // Only cache GET requests
        if ctx.Method() != "GET" {
            ctx.Next()
            return
        }
        
        // Generate cache key
        key := fmt.Sprintf("http:%s:%s", ctx.Method(), ctx.Request.URL.Path)
        if ctx.Request.URL.RawQuery != "" {
            key += "?" + ctx.Request.URL.RawQuery
        }
        
        // Check cache
        if data, found := cm.Get(key); found {
            ctx.JSON(200, data)
            ctx.Header("X-Cache", "HIT")
            return
        }
        
        // Capture response
        recorder := &ResponseRecorder{
            original: ctx.Writer,
        }
        ctx.Writer = recorder
        
        ctx.Next()
        
        // Cache successful responses
        if recorder.statusCode >= 200 && recorder.statusCode < 300 && recorder.data != nil {
            cm.Set(key, recorder.data, ttl)
            ctx.Header("X-Cache", "MISS")
        }
        
        // Restore original writer
        ctx.Writer = recorder.original
    }
}

type ResponseRecorder struct {
    original   zoox.ResponseWriter
    data       interface{}
    statusCode int
}

func (rr *ResponseRecorder) Header() map[string][]string {
    return rr.original.Header()
}

func (rr *ResponseRecorder) Write(data []byte) (int, error) {
    // Try to parse JSON data for caching
    var jsonData interface{}
    if err := json.Unmarshal(data, &jsonData); err == nil {
        rr.data = jsonData
    }
    
    return rr.original.Write(data)
}

func (rr *ResponseRecorder) WriteHeader(statusCode int) {
    rr.statusCode = statusCode
    rr.original.WriteHeader(statusCode)
}

func main() {
    app := zoox.New()
    
    // Create cache manager
    cacheManager := NewCacheManager(5*time.Minute, 30*time.Minute)
    
    // Apply cache middleware to specific routes
    app.Use(cacheManager.CacheMiddleware(10 * time.Minute))
    
    // Sample data
    products := []map[string]interface{}{
        {"id": 1, "name": "Laptop", "price": 999.99, "category": "Electronics"},
        {"id": 2, "name": "Mouse", "price": 29.99, "category": "Electronics"},
        {"id": 3, "name": "Keyboard", "price": 79.99, "category": "Electronics"},
        {"id": 4, "name": "Monitor", "price": 299.99, "category": "Electronics"},
    }
    
    // Cached endpoints
    app.Get("/products", func(ctx *zoox.Context) {
        // Simulate database query delay
        time.Sleep(100 * time.Millisecond)
        
        category := ctx.Query("category")
        if category != "" {
            filtered := make([]map[string]interface{}, 0)
            for _, product := range products {
                if product["category"] == category {
                    filtered = append(filtered, product)
                }
            }
            ctx.JSON(200, map[string]interface{}{
                "products": filtered,
                "count":    len(filtered),
                "cached":   false,
            })
            return
        }
        
        ctx.JSON(200, map[string]interface{}{
            "products": products,
            "count":    len(products),
            "cached":   false,
        })
    })
    
    app.Get("/products/:id", func(ctx *zoox.Context) {
        id := ctx.ParamInt("id")
        
        // Simulate database query delay
        time.Sleep(50 * time.Millisecond)
        
        for _, product := range products {
            if product["id"] == id {
                ctx.JSON(200, map[string]interface{}{
                    "product": product,
                    "cached":  false,
                })
                return
            }
        }
        
        ctx.JSON(404, map[string]string{"error": "Product not found"})
    })
    
    // Cache management endpoints
    app.Get("/cache/stats", func(ctx *zoox.Context) {
        ctx.JSON(200, cacheManager.Stats())
    })
    
    app.Delete("/cache", func(ctx *zoox.Context) {
        cacheManager.l1Cache.Clear()
        if cacheManager.l2Cache != nil {
            cacheManager.l2Cache.Clear()
        }
        
        ctx.JSON(200, map[string]string{"message": "Cache cleared"})
    })
    
    app.Delete("/cache/:key", func(ctx *zoox.Context) {
        key := ctx.Param("key")
        cacheManager.Delete(key)
        
        ctx.JSON(200, map[string]string{"message": "Cache key deleted"})
    })
    
    // Demo interface
    app.Get("/", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Caching Demo</title>
            <style>
                body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
                .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
                button { padding: 10px 20px; margin: 5px; cursor: pointer; }
                .result { margin-top: 10px; padding: 10px; background: #f0f0f0; border-radius: 3px; }
                .cache-hit { color: green; }
                .cache-miss { color: orange; }
                .stats { display: flex; gap: 20px; }
                .stat { flex: 1; text-align: center; }
            </style>
        </head>
        <body>
            <h1>Caching Strategies Demo</h1>
            
            <div class="section">
                <h3>Test Cached Endpoints</h3>
                <button onclick="testEndpoint('/products')">Get All Products</button>
                <button onclick="testEndpoint('/products?category=Electronics')">Get Electronics</button>
                <button onclick="testEndpoint('/products/1')">Get Product 1</button>
                <button onclick="testEndpoint('/products/2')">Get Product 2</button>
                <div id="testResult" class="result"></div>
            </div>
            
            <div class="section">
                <h3>Cache Management</h3>
                <button onclick="getStats()">Get Cache Stats</button>
                <button onclick="clearCache()">Clear Cache</button>
                <div id="statsResult" class="result"></div>
            </div>
            
            <div class="section">
                <h3>Performance Test</h3>
                <button onclick="performanceTest()">Run Performance Test</button>
                <div id="perfResult" class="result"></div>
            </div>
            
            <script>
                async function testEndpoint(endpoint) {
                    const start = performance.now();
                    
                    try {
                        const response = await fetch(endpoint);
                        const data = await response.json();
                        const end = performance.now();
                        
                        const cacheStatus = response.headers.get('X-Cache') || 'UNKNOWN';
                        const statusClass = cacheStatus === 'HIT' ? 'cache-hit' : 'cache-miss';
                        
                        document.getElementById('testResult').innerHTML = \`
                            <strong>Endpoint:</strong> \${endpoint}<br>
                            <strong>Cache Status:</strong> <span class="\${statusClass}">\${cacheStatus}</span><br>
                            <strong>Response Time:</strong> \${(end - start).toFixed(2)}ms<br>
                            <strong>Data:</strong> \${JSON.stringify(data, null, 2)}
                        \`;
                    } catch (error) {
                        document.getElementById('testResult').innerHTML = 'Error: ' + error.message;
                    }
                }
                
                async function getStats() {
                    try {
                        const response = await fetch('/cache/stats');
                        const stats = await response.json();
                        
                        document.getElementById('statsResult').innerHTML = \`
                            <strong>Cache Statistics:</strong><br>
                            <pre>\${JSON.stringify(stats, null, 2)}</pre>
                        \`;
                    } catch (error) {
                        document.getElementById('statsResult').innerHTML = 'Error: ' + error.message;
                    }
                }
                
                async function clearCache() {
                    try {
                        const response = await fetch('/cache', { method: 'DELETE' });
                        const result = await response.json();
                        
                        document.getElementById('statsResult').innerHTML = 'Cache cleared successfully';
                    } catch (error) {
                        document.getElementById('statsResult').innerHTML = 'Error: ' + error.message;
                    }
                }
                
                async function performanceTest() {
                    const endpoint = '/products';
                    const iterations = 10;
                    const results = [];
                    
                    document.getElementById('perfResult').innerHTML = 'Running performance test...';
                    
                    // Clear cache first
                    await fetch('/cache', { method: 'DELETE' });
                    
                    for (let i = 0; i < iterations; i++) {
                        const start = performance.now();
                        const response = await fetch(endpoint);
                        await response.json();
                        const end = performance.now();
                        
                        const cacheStatus = response.headers.get('X-Cache') || 'UNKNOWN';
                        results.push({
                            iteration: i + 1,
                            time: end - start,
                            cacheStatus: cacheStatus
                        });
                    }
                    
                    const avgTime = results.reduce((sum, r) => sum + r.time, 0) / results.length;
                    const hitCount = results.filter(r => r.cacheStatus === 'HIT').length;
                    const missCount = results.filter(r => r.cacheStatus === 'MISS').length;
                    
                    document.getElementById('perfResult').innerHTML = \`
                        <strong>Performance Test Results:</strong><br>
                        <strong>Iterations:</strong> \${iterations}<br>
                        <strong>Average Response Time:</strong> \${avgTime.toFixed(2)}ms<br>
                        <strong>Cache Hits:</strong> \${hitCount}<br>
                        <strong>Cache Misses:</strong> \${missCount}<br>
                        <strong>Cache Hit Rate:</strong> \${((hitCount / iterations) * 100).toFixed(1)}%<br>
                        <br>
                        <strong>Detailed Results:</strong><br>
                        <pre>\${JSON.stringify(results, null, 2)}</pre>
                    \`;
                }
            </script>
        </body>
        </html>
        `
        ctx.HTML(200, html, nil)
    })
    
    log.Println("Caching demo server starting on :8080")
    log.Println("Demo: http://localhost:8080")
    log.Println("Note: Redis cache disabled in this demo (L1 cache only)")
    
    app.Listen(":8080")
}
```

## 📚 Key Takeaways

1. **Multi-level Caching**: Combine memory and distributed caching
2. **Cache Invalidation**: Implement proper cache expiration
3. **Performance Monitoring**: Track cache hit rates and performance
4. **Strategic Caching**: Cache expensive operations and frequent requests
5. **Cache Management**: Provide tools for cache administration

## 🎯 Next Steps

- Learn [Tutorial 13: Monitoring & Logging](./13-monitoring-logging.md)
- Explore [Tutorial 14: Testing Strategies](./14-testing-strategies.md)
- Study [Tutorial 16: Performance Optimization](./16-performance-optimization.md)

---

**Congratulations!** You've mastered caching strategies in Zoox! 