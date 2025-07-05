# Tutorial 13: Monitoring & Logging

## 📖 Overview

Learn to implement comprehensive monitoring and logging for Zoox applications. This tutorial covers structured logging, metrics collection, health checks, and performance monitoring for production-ready applications.

## 🎯 Learning Objectives

- Implement structured logging
- Collect application metrics
- Build health check systems
- Monitor application performance
- Set up alerting and dashboards

## 📋 Prerequisites

- Completed [Tutorial 01: Getting Started](./01-getting-started.md)
- Understanding of logging and monitoring concepts
- Basic knowledge of metrics and observability

## 🚀 Getting Started

### Comprehensive Monitoring System

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "runtime"
    "sync"
    "time"
    
    "github.com/go-zoox/zoox"
    "github.com/sirupsen/logrus"
)

// Structured Logger
type Logger struct {
    *logrus.Logger
}

func NewLogger() *Logger {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })
    logger.SetLevel(logrus.InfoLevel)
    
    return &Logger{Logger: logger}
}

func (l *Logger) LogRequest(method, path string, statusCode int, duration time.Duration, userID string) {
    l.WithFields(logrus.Fields{
        "type":        "request",
        "method":      method,
        "path":        path,
        "status_code": statusCode,
        "duration_ms": duration.Milliseconds(),
        "user_id":     userID,
    }).Info("HTTP request processed")
}

func (l *Logger) LogError(err error, context map[string]interface{}) {
    fields := logrus.Fields{
        "type":  "error",
        "error": err.Error(),
    }
    
    for k, v := range context {
        fields[k] = v
    }
    
    l.WithFields(fields).Error("Application error occurred")
}

func (l *Logger) LogMetric(name string, value float64, tags map[string]string) {
    fields := logrus.Fields{
        "type":        "metric",
        "metric_name": name,
        "value":       value,
    }
    
    for k, v := range tags {
        fields["tag_"+k] = v
    }
    
    l.WithFields(fields).Info("Metric recorded")
}

// Metrics Collector
type Metrics struct {
    counters map[string]int64
    gauges   map[string]float64
    timers   map[string][]time.Duration
    mutex    sync.RWMutex
}

func NewMetrics() *Metrics {
    return &Metrics{
        counters: make(map[string]int64),
        gauges:   make(map[string]float64),
        timers:   make(map[string][]time.Duration),
    }
}

func (m *Metrics) IncrementCounter(name string) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.counters[name]++
}

func (m *Metrics) SetGauge(name string, value float64) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.gauges[name] = value
}

func (m *Metrics) RecordTimer(name string, duration time.Duration) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.timers[name] = append(m.timers[name], duration)
    
    // Keep only last 100 measurements
    if len(m.timers[name]) > 100 {
        m.timers[name] = m.timers[name][1:]
    }
}

func (m *Metrics) GetStats() map[string]interface{} {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    stats := map[string]interface{}{
        "counters": make(map[string]int64),
        "gauges":   make(map[string]float64),
        "timers":   make(map[string]map[string]float64),
    }
    
    // Copy counters
    for k, v := range m.counters {
        stats["counters"].(map[string]int64)[k] = v
    }
    
    // Copy gauges
    for k, v := range m.gauges {
        stats["gauges"].(map[string]float64)[k] = v
    }
    
    // Process timers
    timers := make(map[string]map[string]float64)
    for name, durations := range m.timers {
        if len(durations) == 0 {
            continue
        }
        
        var total time.Duration
        min := durations[0]
        max := durations[0]
        
        for _, d := range durations {
            total += d
            if d < min {
                min = d
            }
            if d > max {
                max = d
            }
        }
        
        avg := total / time.Duration(len(durations))
        
        timers[name] = map[string]float64{
            "count":   float64(len(durations)),
            "avg_ms":  float64(avg.Milliseconds()),
            "min_ms":  float64(min.Milliseconds()),
            "max_ms":  float64(max.Milliseconds()),
            "total_ms": float64(total.Milliseconds()),
        }
    }
    
    stats["timers"] = timers
    return stats
}

// Health Check System
type HealthChecker struct {
    checks map[string]HealthCheck
    mutex  sync.RWMutex
}

type HealthCheck func() error

type HealthStatus struct {
    Status string                 `json:"status"`
    Checks map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
}

func NewHealthChecker() *HealthChecker {
    return &HealthChecker{
        checks: make(map[string]HealthCheck),
    }
}

func (hc *HealthChecker) AddCheck(name string, check HealthCheck) {
    hc.mutex.Lock()
    defer hc.mutex.Unlock()
    hc.checks[name] = check
}

func (hc *HealthChecker) CheckHealth() HealthStatus {
    hc.mutex.RLock()
    defer hc.mutex.RUnlock()
    
    status := HealthStatus{
        Status: "healthy",
        Checks: make(map[string]CheckResult),
    }
    
    for name, check := range hc.checks {
        if err := check(); err != nil {
            status.Checks[name] = CheckResult{
                Status:  "unhealthy",
                Message: err.Error(),
            }
            status.Status = "unhealthy"
        } else {
            status.Checks[name] = CheckResult{
                Status: "healthy",
            }
        }
    }
    
    return status
}

// Monitoring Middleware
func MonitoringMiddleware(logger *Logger, metrics *Metrics) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        start := time.Now()
        
        // Process request
        ctx.Next()
        
        // Record metrics
        duration := time.Since(start)
        path := ctx.Request.URL.Path
        method := ctx.Method()
        statusCode := ctx.Writer.Status()
        
        // Log request
        userID := ""
        if user := ctx.Get("user"); user != nil {
            userID = fmt.Sprintf("%v", user)
        }
        
        logger.LogRequest(method, path, statusCode, duration, userID)
        
        // Update metrics
        metrics.IncrementCounter("http_requests_total")
        metrics.IncrementCounter(fmt.Sprintf("http_requests_%s", method))
        metrics.IncrementCounter(fmt.Sprintf("http_status_%d", statusCode))
        metrics.RecordTimer("http_request_duration", duration)
        
        // Update response time gauge
        metrics.SetGauge("http_response_time_ms", float64(duration.Milliseconds()))
    }
}

// System Metrics Collector
func CollectSystemMetrics(metrics *Metrics) {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            // Memory metrics
            metrics.SetGauge("memory_alloc_bytes", float64(m.Alloc))
            metrics.SetGauge("memory_total_alloc_bytes", float64(m.TotalAlloc))
            metrics.SetGauge("memory_sys_bytes", float64(m.Sys))
            metrics.SetGauge("memory_heap_alloc_bytes", float64(m.HeapAlloc))
            metrics.SetGauge("memory_heap_sys_bytes", float64(m.HeapSys))
            
            // GC metrics
            metrics.SetGauge("gc_num", float64(m.NumGC))
            metrics.SetGauge("gc_pause_total_ns", float64(m.PauseTotalNs))
            
            // Goroutine count
            metrics.SetGauge("goroutines", float64(runtime.NumGoroutine()))
        }
    }()
}

func main() {
    app := zoox.New()
    
    // Initialize monitoring components
    logger := NewLogger()
    metrics := NewMetrics()
    healthChecker := NewHealthChecker()
    
    // Start system metrics collection
    CollectSystemMetrics(metrics)
    
    // Add health checks
    healthChecker.AddCheck("database", func() error {
        // Simulate database check
        time.Sleep(10 * time.Millisecond)
        return nil // or return error if unhealthy
    })
    
    healthChecker.AddCheck("external_service", func() error {
        // Simulate external service check
        return nil
    })
    
    // Apply monitoring middleware
    app.Use(MonitoringMiddleware(logger, metrics))
    
    // Health check endpoint
    app.Get("/health", func(ctx *zoox.Context) {
        health := healthChecker.CheckHealth()
        
        if health.Status == "healthy" {
            ctx.JSON(200, health)
        } else {
            ctx.JSON(503, health)
        }
    })
    
    // Metrics endpoint
    app.Get("/metrics", func(ctx *zoox.Context) {
        ctx.JSON(200, metrics.GetStats())
    })
    
    // Sample application endpoints
    app.Get("/api/users", func(ctx *zoox.Context) {
        // Simulate processing time
        time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
        
        users := []map[string]interface{}{
            {"id": 1, "name": "John Doe", "email": "john@example.com"},
            {"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
        }
        
        ctx.JSON(200, users)
    })
    
    app.Get("/api/users/:id", func(ctx *zoox.Context) {
        id := ctx.Param("id")
        
        // Simulate processing time
        time.Sleep(time.Duration(20+rand.Intn(80)) * time.Millisecond)
        
        // Simulate occasional errors
        if rand.Float32() < 0.1 {
            logger.LogError(fmt.Errorf("user not found"), map[string]interface{}{
                "user_id": id,
                "endpoint": "/api/users/:id",
            })
            ctx.JSON(404, map[string]string{"error": "User not found"})
            return
        }
        
        ctx.JSON(200, map[string]interface{}{
            "id":    id,
            "name":  "User " + id,
            "email": "user" + id + "@example.com",
        })
    })
    
    // Error endpoint for testing
    app.Get("/api/error", func(ctx *zoox.Context) {
        err := fmt.Errorf("simulated error")
        logger.LogError(err, map[string]interface{}{
            "endpoint": "/api/error",
            "severity": "high",
        })
        ctx.JSON(500, map[string]string{"error": "Internal server error"})
    })
    
    // Monitoring dashboard
    app.Get("/", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Monitoring Dashboard</title>
            <style>
                body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
                .dashboard { max-width: 1200px; margin: 0 auto; }
                .card { background: white; padding: 20px; margin: 20px 0; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
                .metric { display: inline-block; margin: 10px; padding: 15px; background: #f8f9fa; border-radius: 5px; min-width: 120px; text-align: center; }
                .metric-value { font-size: 24px; font-weight: bold; color: #007bff; }
                .metric-label { font-size: 12px; color: #666; }
                .status-healthy { color: #28a745; }
                .status-unhealthy { color: #dc3545; }
                .logs { max-height: 300px; overflow-y: auto; background: #f8f9fa; padding: 10px; border-radius: 5px; font-family: monospace; font-size: 12px; }
                button { padding: 10px 20px; margin: 5px; border: none; border-radius: 5px; cursor: pointer; }
                .btn-primary { background: #007bff; color: white; }
                .btn-success { background: #28a745; color: white; }
                .btn-danger { background: #dc3545; color: white; }
            </style>
        </head>
        <body>
            <div class="dashboard">
                <h1>Monitoring Dashboard</h1>
                
                <div class="card">
                    <h3>Health Status</h3>
                    <div id="healthStatus">Loading...</div>
                    <button class="btn-primary" onclick="checkHealth()">Check Health</button>
                </div>
                
                <div class="card">
                    <h3>Metrics</h3>
                    <div id="metrics">Loading...</div>
                    <button class="btn-primary" onclick="loadMetrics()">Refresh Metrics</button>
                </div>
                
                <div class="card">
                    <h3>Test Endpoints</h3>
                    <button class="btn-success" onclick="testEndpoint('/api/users')">Test Users API</button>
                    <button class="btn-success" onclick="testEndpoint('/api/users/1')">Test User Detail</button>
                    <button class="btn-danger" onclick="testEndpoint('/api/error')">Test Error</button>
                </div>
                
                <div class="card">
                    <h3>Load Test</h3>
                    <button class="btn-primary" onclick="runLoadTest()">Run Load Test</button>
                    <div id="loadTestResult"></div>
                </div>
            </div>
            
            <script>
                let refreshInterval;
                
                async function checkHealth() {
                    try {
                        const response = await fetch('/health');
                        const health = await response.json();
                        
                        let html = '<div class="status-' + (health.status === 'healthy' ? 'healthy' : 'unhealthy') + '">';
                        html += '<strong>Overall Status: ' + health.status.toUpperCase() + '</strong></div><br>';
                        
                        for (const [name, check] of Object.entries(health.checks)) {
                            html += '<div class="status-' + (check.status === 'healthy' ? 'healthy' : 'unhealthy') + '">';
                            html += name + ': ' + check.status.toUpperCase();
                            if (check.message) {
                                html += ' - ' + check.message;
                            }
                            html += '</div>';
                        }
                        
                        document.getElementById('healthStatus').innerHTML = html;
                    } catch (error) {
                        document.getElementById('healthStatus').innerHTML = 'Error checking health: ' + error.message;
                    }
                }
                
                async function loadMetrics() {
                    try {
                        const response = await fetch('/metrics');
                        const metrics = await response.json();
                        
                        let html = '';
                        
                        // Counters
                        if (metrics.counters) {
                            html += '<h4>Counters</h4>';
                            for (const [name, value] of Object.entries(metrics.counters)) {
                                html += '<div class="metric"><div class="metric-value">' + value + '</div><div class="metric-label">' + name + '</div></div>';
                            }
                        }
                        
                        // Gauges
                        if (metrics.gauges) {
                            html += '<h4>Gauges</h4>';
                            for (const [name, value] of Object.entries(metrics.gauges)) {
                                const displayValue = name.includes('bytes') ? formatBytes(value) : value.toFixed(2);
                                html += '<div class="metric"><div class="metric-value">' + displayValue + '</div><div class="metric-label">' + name + '</div></div>';
                            }
                        }
                        
                        // Timers
                        if (metrics.timers) {
                            html += '<h4>Timers</h4>';
                            for (const [name, timer] of Object.entries(metrics.timers)) {
                                html += '<div class="metric"><div class="metric-value">' + timer.avg_ms.toFixed(2) + 'ms</div><div class="metric-label">' + name + ' (avg)</div></div>';
                            }
                        }
                        
                        document.getElementById('metrics').innerHTML = html;
                    } catch (error) {
                        document.getElementById('metrics').innerHTML = 'Error loading metrics: ' + error.message;
                    }
                }
                
                async function testEndpoint(endpoint) {
                    try {
                        const start = performance.now();
                        const response = await fetch(endpoint);
                        const end = performance.now();
                        
                        console.log('Tested ' + endpoint + ' - Status: ' + response.status + ' - Time: ' + (end - start).toFixed(2) + 'ms');
                        
                        // Refresh metrics after test
                        setTimeout(loadMetrics, 100);
                    } catch (error) {
                        console.error('Error testing endpoint:', error);
                    }
                }
                
                async function runLoadTest() {
                    const button = event.target;
                    button.disabled = true;
                    button.textContent = 'Running...';
                    
                    const results = [];
                    const requests = 50;
                    
                    for (let i = 0; i < requests; i++) {
                        const start = performance.now();
                        try {
                            const response = await fetch('/api/users');
                            const end = performance.now();
                            results.push({
                                success: response.ok,
                                time: end - start,
                                status: response.status
                            });
                        } catch (error) {
                            results.push({
                                success: false,
                                time: 0,
                                error: error.message
                            });
                        }
                    }
                    
                    const successful = results.filter(r => r.success).length;
                    const avgTime = results.reduce((sum, r) => sum + r.time, 0) / results.length;
                    
                    document.getElementById('loadTestResult').innerHTML = 
                        '<h4>Load Test Results</h4>' +
                        '<p>Requests: ' + requests + '</p>' +
                        '<p>Successful: ' + successful + ' (' + ((successful/requests)*100).toFixed(1) + '%)</p>' +
                        '<p>Average Response Time: ' + avgTime.toFixed(2) + 'ms</p>';
                    
                    button.disabled = false;
                    button.textContent = 'Run Load Test';
                    
                    // Refresh metrics
                    setTimeout(loadMetrics, 500);
                }
                
                function formatBytes(bytes) {
                    if (bytes === 0) return '0 Bytes';
                    const k = 1024;
                    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
                    const i = Math.floor(Math.log(bytes) / Math.log(k));
                    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
                }
                
                // Auto-refresh
                checkHealth();
                loadMetrics();
                refreshInterval = setInterval(() => {
                    checkHealth();
                    loadMetrics();
                }, 30000);
            </script>
        </body>
        </html>
        `
        ctx.HTML(200, html, nil)
    })
    
    log.Println("Monitoring server starting on :8080")
    log.Println("Dashboard: http://localhost:8080")
    log.Println("Health: http://localhost:8080/health")
    log.Println("Metrics: http://localhost:8080/metrics")
    
    app.Listen(":8080")
}
```

## 📚 Key Takeaways

1. **Structured Logging**: Use structured logs for better analysis
2. **Metrics Collection**: Track key performance indicators
3. **Health Checks**: Monitor system health and dependencies
4. **Real-time Monitoring**: Build dashboards for operational visibility
5. **Alerting**: Set up alerts for critical issues

## 🎯 Next Steps

- Learn [Tutorial 14: Testing Strategies](./14-testing-strategies.md)
- Explore [Tutorial 15: Security Best Practices](./15-security-best-practices.md)
- Study [Tutorial 16: Performance Optimization](./16-performance-optimization.md)

---

**Congratulations!** You've mastered monitoring and logging in Zoox! 