# Tutorial 18: Production Monitoring

## Overview
Master comprehensive production monitoring for Zoox applications, including observability, alerting, logging, metrics collection, and incident response.

## Learning Objectives
- Implement comprehensive monitoring strategies
- Set up metrics collection and visualization
- Configure alerting and notification systems
- Master structured logging and log analysis
- Monitor application performance and health
- Implement distributed tracing
- Set up incident response procedures

## Prerequisites
- Complete Tutorial 17: Deployment Strategies
- Understanding of monitoring concepts
- Experience with production systems

## Metrics and Observability

### Prometheus Integration

```go
// monitoring/prometheus.go
package monitoring

import (
    "net/http"
    "strconv"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/go-zoox/zoox"
)

// Application metrics
var (
    // Request metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status_code"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint"},
    )

    HTTPActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_active_connections",
            Help: "Number of active HTTP connections",
        },
    )

    // Application metrics
    DatabaseConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "database_connections",
            Help: "Number of database connections",
        },
        []string{"database", "state"},
    )

    CacheOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_operations_total",
            Help: "Total number of cache operations",
        },
        []string{"operation", "result"},
    )

    BusinessMetrics = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "business_events_total",
            Help: "Total number of business events",
        },
        []string{"event_type", "status"},
    )

    ErrorRate = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "application_errors_total",
            Help: "Total number of application errors",
        },
        []string{"error_type", "severity"},
    )

    // Resource metrics
    MemoryUsage = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "memory_usage_bytes",
            Help: "Current memory usage in bytes",
        },
    )

    GoroutineCount = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "goroutines_count",
            Help: "Number of goroutines",
        },
    )
)

// Metrics middleware
func MetricsMiddleware() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        start := time.Now()
        
        // Increment active connections
        HTTPActiveConnections.Inc()
        defer HTTPActiveConnections.Dec()

        ctx.Next()

        // Record metrics
        duration := time.Since(start).Seconds()
        method := ctx.Method()
        endpoint := sanitizeEndpoint(ctx.Request().URL.Path)
        statusCode := strconv.Itoa(ctx.Writer.Status())

        HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
        HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)

        // Record error metrics
        if ctx.Writer.Status() >= 400 {
            severity := "warning"
            if ctx.Writer.Status() >= 500 {
                severity = "error"
            }
            ErrorRate.WithLabelValues("http_error", severity).Inc()
        }
    }
}

// Sanitize endpoint to reduce cardinality
func sanitizeEndpoint(path string) string {
    // Replace dynamic segments with placeholders
    // e.g., /users/123 becomes /users/:id
    // This is a simple implementation - use a proper router for complex cases
    if matches := regexp.MustCompile(`/\d+`).FindAllString(path, -1); len(matches) > 0 {
        for _, match := range matches {
            path = strings.Replace(path, match, "/:id", 1)
        }
    }
    return path
}

// Metrics endpoint handler
func MetricsHandler() http.Handler {
    return promhttp.Handler()
}

// Custom business metrics
func RecordUserRegistration(success bool) {
    status := "success"
    if !success {
        status = "failure"
    }
    BusinessMetrics.WithLabelValues("user_registration", status).Inc()
}

func RecordCacheHit(hit bool) {
    result := "hit"
    if !hit {
        result = "miss"
    }
    CacheOperations.WithLabelValues("get", result).Inc()
}

func RecordDatabaseConnectionState(database, state string, count float64) {
    DatabaseConnections.WithLabelValues(database, state).Set(count)
}
```

## Structured Logging

```go
// logging/logger.go
package logging

import (
    "context"
    "os"
    "time"

    "github.com/sirupsen/logrus"
    "github.com/go-zoox/zoox"
)

// Logger configuration
type Config struct {
    Level      string
    Format     string
    Output     string
    ServiceName string
    Version    string
}

// Custom logger
type Logger struct {
    *logrus.Logger
    config Config
}

// Log entry with context
type Entry struct {
    *logrus.Entry
}

func NewLogger(config Config) *Logger {
    log := logrus.New()

    // Set log level
    level, err := logrus.ParseLevel(config.Level)
    if err != nil {
        level = logrus.InfoLevel
    }
    log.SetLevel(level)

    // Set formatter
    if config.Format == "json" {
        log.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: time.RFC3339,
            FieldMap: logrus.FieldMap{
                logrus.FieldKeyTime:  "timestamp",
                logrus.FieldKeyLevel: "level",
                logrus.FieldKeyMsg:   "message",
            },
        })
    } else {
        log.SetFormatter(&logrus.TextFormatter{
            FullTimestamp: true,
            TimestampFormat: time.RFC3339,
        })
    }

    // Set output
    if config.Output == "stdout" {
        log.SetOutput(os.Stdout)
    } else {
        file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err == nil {
            log.SetOutput(file)
        }
    }

    return &Logger{
        Logger: log,
        config: config,
    }
}

// Create entry with default fields
func (l *Logger) WithContext(ctx context.Context) *Entry {
    entry := l.Logger.WithFields(logrus.Fields{
        "service": l.config.ServiceName,
        "version": l.config.Version,
    })

    // Add request context if available
    if requestID := ctx.Value("request_id"); requestID != nil {
        entry = entry.WithField("request_id", requestID)
    }
    if userID := ctx.Value("user_id"); userID != nil {
        entry = entry.WithField("user_id", userID)
    }

    return &Entry{Entry: entry}
}

// Logging middleware
func LoggingMiddleware(logger *Logger) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        start := time.Now()

        // Generate request ID
        requestID := generateRequestID()
        ctx.Set("request_id", requestID)

        // Create context with request info
        logCtx := context.WithValue(context.Background(), "request_id", requestID)
        if userID, exists := ctx.Get("user_id"); exists {
            logCtx = context.WithValue(logCtx, "user_id", userID)
        }

        // Log request start
        logger.WithContext(logCtx).WithFields(logrus.Fields{
            "method":     ctx.Method(),
            "path":       ctx.Request().URL.Path,
            "query":      ctx.Request().URL.RawQuery,
            "ip":         ctx.ClientIP(),
            "user_agent": ctx.Header("User-Agent"),
        }).Info("Request started")

        ctx.Next()

        // Log request completion
        duration := time.Since(start)
        entry := logger.WithContext(logCtx).WithFields(logrus.Fields{
            "method":      ctx.Method(),
            "path":        ctx.Request().URL.Path,
            "status_code": ctx.Writer.Status(),
            "duration_ms": duration.Milliseconds(),
            "bytes_sent":  ctx.Writer.Size(),
        })

        if ctx.Writer.Status() >= 400 {
            if ctx.Writer.Status() >= 500 {
                entry.Error("Request completed with server error")
            } else {
                entry.Warn("Request completed with client error")
            }
        } else {
            entry.Info("Request completed successfully")
        }
    }
}

// Generate unique request ID
func generateRequestID() string {
    return fmt.Sprintf("%d-%s", time.Now().UnixNano(), 
        strings.Replace(uuid.New().String(), "-", "", -1)[:8])
}

// Structured error logging
func (e *Entry) LogError(err error, msg string, fields map[string]interface{}) {
    entry := e.WithError(err).WithField("error_type", fmt.Sprintf("%T", err))
    for k, v := range fields {
        entry = entry.WithField(k, v)
    }
    entry.Error(msg)
}

// Business event logging
func (e *Entry) LogBusinessEvent(eventType string, fields map[string]interface{}) {
    entry := e.WithField("event_type", eventType)
    for k, v := range fields {
        entry = entry.WithField(k, v)
    }
    entry.Info("Business event recorded")
}

// Security event logging
func (e *Entry) LogSecurityEvent(eventType string, severity string, fields map[string]interface{}) {
    entry := e.WithFields(logrus.Fields{
        "security_event": eventType,
        "severity":       severity,
    })
    for k, v := range fields {
        entry = entry.WithField(k, v)
    }
    entry.Warn("Security event detected")
}
```

## Health Monitoring

```go
// health/comprehensive.go
package health

import (
    "context"
    "database/sql"
    "fmt"
    "net/http"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/go-zoox/zoox"
)

type HealthStatus string

const (
    StatusHealthy   HealthStatus = "healthy"
    StatusUnhealthy HealthStatus = "unhealthy"
    StatusDegraded  HealthStatus = "degraded"
)

type HealthCheck struct {
    Name        string                 `json:"name"`
    Status      HealthStatus           `json:"status"`
    Message     string                 `json:"message,omitempty"`
    LastChecked time.Time              `json:"last_checked"`
    Duration    time.Duration          `json:"duration"`
    Details     map[string]interface{} `json:"details,omitempty"`
}

type HealthResponse struct {
    Status    HealthStatus           `json:"status"`
    Timestamp time.Time              `json:"timestamp"`
    Uptime    time.Duration          `json:"uptime"`
    Version   string                 `json:"version"`
    Checks    map[string]HealthCheck `json:"checks"`
}

type HealthMonitor struct {
    startTime time.Time
    version   string
    checks    map[string]func() HealthCheck
}

func NewHealthMonitor(version string) *HealthMonitor {
    return &HealthMonitor{
        startTime: time.Now(),
        version:   version,
        checks:    make(map[string]func() HealthCheck),
    }
}

// Add health checks
func (hm *HealthMonitor) AddDatabaseCheck(name string, db *sql.DB) {
    hm.checks[name] = func() HealthCheck {
        start := time.Now()
        
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        if err := db.PingContext(ctx); err != nil {
            return HealthCheck{
                Name:        name,
                Status:      StatusUnhealthy,
                Message:     err.Error(),
                LastChecked: time.Now(),
                Duration:    time.Since(start),
            }
        }
        
        // Check connection pool stats
        stats := db.Stats()
        details := map[string]interface{}{
            "open_connections": stats.OpenConnections,
            "in_use":          stats.InUse,
            "idle":            stats.Idle,
            "max_open":        stats.MaxOpenConnections,
        }
        
        status := StatusHealthy
        if stats.OpenConnections > int(float64(stats.MaxOpenConnections)*0.8) {
            status = StatusDegraded
        }
        
        return HealthCheck{
            Name:        name,
            Status:      status,
            LastChecked: time.Now(),
            Duration:    time.Since(start),
            Details:     details,
        }
    }
}

func (hm *HealthMonitor) AddRedisCheck(name string, client *redis.Client) {
    hm.checks[name] = func() HealthCheck {
        start := time.Now()
        
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        // Test basic connectivity
        if err := client.Ping(ctx).Err(); err != nil {
            return HealthCheck{
                Name:        name,
                Status:      StatusUnhealthy,
                Message:     err.Error(),
                LastChecked: time.Now(),
                Duration:    time.Since(start),
            }
        }
        
        // Get Redis info
        info, err := client.Info(ctx, "memory").Result()
        if err != nil {
            return HealthCheck{
                Name:        name,
                Status:      StatusDegraded,
                Message:     fmt.Sprintf("Could not get Redis info: %v", err),
                LastChecked: time.Now(),
                Duration:    time.Since(start),
            }
        }
        
        return HealthCheck{
            Name:        name,
            Status:      StatusHealthy,
            LastChecked: time.Now(),
            Duration:    time.Since(start),
            Details: map[string]interface{}{
                "info": info,
            },
        }
    }
}

func (hm *HealthMonitor) AddCustomCheck(name string, checkFn func() HealthCheck) {
    hm.checks[name] = checkFn
}

// HTTP handlers
func (hm *HealthMonitor) HealthHandler() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        response := hm.runHealthChecks()
        
        status := http.StatusOK
        if response.Status == StatusUnhealthy {
            status = http.StatusServiceUnavailable
        } else if response.Status == StatusDegraded {
            status = http.StatusOK // Still return 200 for degraded
        }
        
        ctx.JSON(status, response)
    }
}

func (hm *HealthMonitor) ReadinessHandler() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        response := hm.runHealthChecks()
        
        // For readiness, only healthy is acceptable
        if response.Status == StatusHealthy {
            ctx.JSON(http.StatusOK, map[string]interface{}{
                "ready": true,
                "timestamp": time.Now(),
            })
        } else {
            ctx.JSON(http.StatusServiceUnavailable, map[string]interface{}{
                "ready": false,
                "timestamp": time.Now(),
                "checks": response.Checks,
            })
        }
    }
}

func (hm *HealthMonitor) LivenessHandler() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        // Simple liveness check - just return OK if server is running
        ctx.JSON(http.StatusOK, map[string]interface{}{
            "alive": true,
            "timestamp": time.Now(),
            "uptime": time.Since(hm.startTime).String(),
        })
    }
}

func (hm *HealthMonitor) runHealthChecks() HealthResponse {
    checks := make(map[string]HealthCheck)
    overallStatus := StatusHealthy
    
    for name, checkFn := range hm.checks {
        check := checkFn()
        checks[name] = check
        
        if check.Status == StatusUnhealthy {
            overallStatus = StatusUnhealthy
        } else if check.Status == StatusDegraded && overallStatus == StatusHealthy {
            overallStatus = StatusDegraded
        }
    }
    
    return HealthResponse{
        Status:    overallStatus,
        Timestamp: time.Now(),
        Uptime:    time.Since(hm.startTime),
        Version:   hm.version,
        Checks:    checks,
    }
}
```

## Alerting System

```go
// alerting/manager.go
package alerting

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type AlertLevel string

const (
    AlertInfo     AlertLevel = "info"
    AlertWarning  AlertLevel = "warning"
    AlertCritical AlertLevel = "critical"
)

type Alert struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    Level       AlertLevel             `json:"level"`
    Service     string                 `json:"service"`
    Timestamp   time.Time              `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Resolved    bool                   `json:"resolved"`
}

type AlertManager struct {
    alerts      map[string]*Alert
    mutex       sync.RWMutex
    webhookURL  string
    slackToken  string
    emailConfig EmailConfig
}

type EmailConfig struct {
    SMTPHost     string
    SMTPPort     int
    Username     string
    Password     string
    FromAddress  string
    ToAddresses  []string
}

func NewAlertManager(webhookURL, slackToken string, emailConfig EmailConfig) *AlertManager {
    return &AlertManager{
        alerts:      make(map[string]*Alert),
        webhookURL:  webhookURL,
        slackToken:  slackToken,
        emailConfig: emailConfig,
    }
}

func (am *AlertManager) TriggerAlert(alert Alert) {
    alert.ID = generateAlertID()
    alert.Timestamp = time.Now()
    
    am.mutex.Lock()
    am.alerts[alert.ID] = &alert
    am.mutex.Unlock()
    
    // Send notifications
    go am.sendNotifications(alert)
}

func (am *AlertManager) ResolveAlert(alertID string) {
    am.mutex.Lock()
    if alert, exists := am.alerts[alertID]; exists {
        alert.Resolved = true
        alert.Timestamp = time.Now()
    }
    am.mutex.Unlock()
}

func (am *AlertManager) GetActiveAlerts() []*Alert {
    am.mutex.RLock()
    defer am.mutex.RUnlock()
    
    var active []*Alert
    for _, alert := range am.alerts {
        if !alert.Resolved {
            active = append(active, alert)
        }
    }
    return active
}

func (am *AlertManager) sendNotifications(alert Alert) {
    // Send to webhook
    if am.webhookURL != "" {
        am.sendWebhook(alert)
    }
    
    // Send to Slack
    if am.slackToken != "" {
        am.sendSlack(alert)
    }
    
    // Send email for critical alerts
    if alert.Level == AlertCritical && am.emailConfig.SMTPHost != "" {
        am.sendEmail(alert)
    }
}

func (am *AlertManager) sendWebhook(alert Alert) {
    payload, _ := json.Marshal(alert)
    
    client := &http.Client{Timeout: 10 * time.Second}
    _, err := client.Post(am.webhookURL, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        fmt.Printf("Failed to send webhook: %v\n", err)
    }
}

func (am *AlertManager) sendSlack(alert Alert) {
    color := "good"
    if alert.Level == AlertWarning {
        color = "warning"
    } else if alert.Level == AlertCritical {
        color = "danger"
    }
    
    payload := map[string]interface{}{
        "channel": "#alerts",
        "attachments": []map[string]interface{}{
            {
                "color":     color,
                "title":     alert.Title,
                "text":      alert.Description,
                "fields": []map[string]interface{}{
                    {"title": "Service", "value": alert.Service, "short": true},
                    {"title": "Level", "value": string(alert.Level), "short": true},
                    {"title": "Time", "value": alert.Timestamp.Format(time.RFC3339), "short": true},
                },
            },
        },
    }
    
    jsonPayload, _ := json.Marshal(payload)
    
    req, _ := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonPayload))
    req.Header.Set("Authorization", "Bearer "+am.slackToken)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    _, err := client.Do(req)
    if err != nil {
        fmt.Printf("Failed to send Slack message: %v\n", err)
    }
}

// Monitoring-based alerting
func (am *AlertManager) StartMonitoring() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        am.checkMetrics()
    }
}

func (am *AlertManager) checkMetrics() {
    // Check error rate
    if errorRate := am.getErrorRate(); errorRate > 5.0 {
        am.TriggerAlert(Alert{
            Title:       "High Error Rate",
            Description: fmt.Sprintf("Error rate is %.2f%%, above threshold of 5%%", errorRate),
            Level:       AlertCritical,
            Service:     "api",
            Metadata: map[string]interface{}{
                "error_rate": errorRate,
                "threshold":  5.0,
            },
        })
    }
    
    // Check response time
    if avgResponseTime := am.getAverageResponseTime(); avgResponseTime > 2000 {
        am.TriggerAlert(Alert{
            Title:       "High Response Time",
            Description: fmt.Sprintf("Average response time is %dms, above threshold of 2000ms", avgResponseTime),
            Level:       AlertWarning,
            Service:     "api",
            Metadata: map[string]interface{}{
                "response_time": avgResponseTime,
                "threshold":     2000,
            },
        })
    }
    
    // Check memory usage
    if memoryUsage := am.getMemoryUsage(); memoryUsage > 85.0 {
        am.TriggerAlert(Alert{
            Title:       "High Memory Usage",
            Description: fmt.Sprintf("Memory usage is %.2f%%, above threshold of 85%%", memoryUsage),
            Level:       AlertWarning,
            Service:     "system",
            Metadata: map[string]interface{}{
                "memory_usage": memoryUsage,
                "threshold":    85.0,
            },
        })
    }
}

// Metric collection methods (implement based on your metrics system)
func (am *AlertManager) getErrorRate() float64 {
    // Implementation would query Prometheus or your metrics system
    return 0.0
}

func (am *AlertManager) getAverageResponseTime() int {
    // Implementation would query Prometheus or your metrics system
    return 0
}

func (am *AlertManager) getMemoryUsage() float64 {
    // Implementation would query system metrics
    return 0.0
}

func generateAlertID() string {
    return fmt.Sprintf("%d-%s", time.Now().UnixNano(), 
        strings.Replace(uuid.New().String(), "-", "", -1)[:8])
}
```

## Complete Monitoring Setup

```go
// main.go with comprehensive monitoring
func main() {
    // Configuration
    config := loadConfig()
    
    // Initialize monitoring components
    logger := logging.NewLogger(logging.Config{
        Level:       config.LogLevel,
        Format:      "json",
        Output:      "stdout",
        ServiceName: "zoox-app",
        Version:     config.Version,
    })
    
    healthMonitor := health.NewHealthMonitor(config.Version)
    alertManager := alerting.NewAlertManager(
        config.WebhookURL,
        config.SlackToken,
        config.EmailConfig,
    )
    
    // Initialize database and add health check
    db := initDatabase(config.DatabaseURL)
    healthMonitor.AddDatabaseCheck("primary_db", db)
    
    // Initialize Redis and add health check
    redisClient := initRedis(config.RedisURL)
    healthMonitor.AddRedisCheck("redis_cache", redisClient)
    
    // Add custom health checks
    healthMonitor.AddCustomCheck("external_api", func() health.HealthCheck {
        start := time.Now()
        resp, err := http.Get("https://api.external-service.com/health")
        if err != nil {
            return health.HealthCheck{
                Name:        "external_api",
                Status:      health.StatusUnhealthy,
                Message:     err.Error(),
                LastChecked: time.Now(),
                Duration:    time.Since(start),
            }
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != 200 {
            return health.HealthCheck{
                Name:        "external_api",
                Status:      health.StatusDegraded,
                Message:     fmt.Sprintf("HTTP %d", resp.StatusCode),
                LastChecked: time.Now(),
                Duration:    time.Since(start),
            }
        }
        
        return health.HealthCheck{
            Name:        "external_api",
            Status:      health.StatusHealthy,
            LastChecked: time.Now(),
            Duration:    time.Since(start),
        }
    })
    
    // Start alert monitoring
    go alertManager.StartMonitoring()
    
    // Initialize Zoox app
    app := zoox.New()
    
    // Add monitoring middleware
    app.Use(monitoring.MetricsMiddleware())
    app.Use(logging.LoggingMiddleware(logger))
    
    // Health endpoints
    app.Get("/health", healthMonitor.HealthHandler())
    app.Get("/health/ready", healthMonitor.ReadinessHandler())
    app.Get("/health/live", healthMonitor.LivenessHandler())
    
    // Metrics endpoint for Prometheus
    app.Get("/metrics", func(ctx *zoox.Context) {
        monitoring.MetricsHandler().ServeHTTP(ctx.Writer, ctx.Request())
    })
    
    // Alert management endpoints
    app.Get("/alerts", func(ctx *zoox.Context) {
        alerts := alertManager.GetActiveAlerts()
        ctx.JSON(http.StatusOK, alerts)
    })
    
    // Your application routes here...
    setupApplicationRoutes(app, logger, alertManager)
    
    // Graceful shutdown with monitoring
    setupGracefulShutdown(app, logger, db, redisClient)
    
    logger.WithContext(context.Background()).Info("Starting server with comprehensive monitoring")
    log.Fatal(app.Listen(":" + config.Port))
}

func setupGracefulShutdown(app *zoox.Engine, logger *logging.Logger, db *sql.DB, redis *redis.Client) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        ctx := context.Background()
        logger.WithContext(ctx).Info("Shutting down gracefully...")
        
        // Close database connections
        if err := db.Close(); err != nil {
            logger.WithContext(ctx).WithError(err).Error("Error closing database")
        }
        
        // Close Redis connections
        if err := redis.Close(); err != nil {
            logger.WithContext(ctx).WithError(err).Error("Error closing Redis")
        }
        
        logger.WithContext(ctx).Info("Shutdown complete")
        os.Exit(0)
    }()
}
```

## Monitoring Dashboard Configuration

### Grafana Dashboard JSON

```json
{
  "dashboard": {
    "title": "Zoox Application Monitoring",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(http_requests_total{status_code=~\"5..\"}[5m]) / rate(http_requests_total[5m]) * 100"
          }
        ]
      },
      {
        "title": "Active Connections",
        "type": "singlestat",
        "targets": [
          {
            "expr": "http_active_connections"
          }
        ]
      }
    ]
  }
}
```

## Key Monitoring Takeaways

1. **Metrics**: Collect comprehensive application and business metrics
2. **Logging**: Implement structured logging with correlation IDs
3. **Health Checks**: Monitor application and dependency health
4. **Alerting**: Set up proactive alerting based on SLOs
5. **Observability**: Ensure you can understand system behavior
6. **Dashboards**: Create actionable monitoring dashboards
7. **Incident Response**: Have clear procedures for incident handling

## Next Steps

- Implement distributed tracing with Jaeger/Zipkin
- Set up log aggregation with ELK stack
- Learn about SRE practices and SLOs
- Explore chaos engineering
- Study advanced monitoring patterns

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [The Twelve-Factor App](https://12factor.net/)
- [SRE Book](https://sre.google/books/) 