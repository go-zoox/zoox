package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

// Configuration holds application configuration
type Config struct {
	Port            string
	Environment     string
	JWTSecret       string
	DatabaseURL     string
	RedisURL        string
	LogLevel        string
	RateLimitRPS    int
	CorsOrigins     []string
	TLSCertFile     string
	TLSKeyFile      string
	HealthCheckPath string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		JWTSecret:       getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://localhost/myapp"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		RateLimitRPS:    10,
		CorsOrigins:     []string{"http://localhost:3000", "https://myapp.com"},
		TLSCertFile:     getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:      getEnv("TLS_KEY_FILE", ""),
		HealthCheckPath: getEnv("HEALTH_CHECK_PATH", "/health"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// User represents a user in the system
type User struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at,omitempty"`
}

// AuthService handles authentication
type AuthService struct {
	config *Config
}

// NewAuthService creates a new authentication service
func NewAuthService(config *Config) *AuthService {
	return &AuthService{config: config}
}

// ValidateJWT validates a JWT token and returns user info
func (a *AuthService) ValidateJWT(token string) (*User, error) {
	// In production, use a proper JWT library like github.com/golang-jwt/jwt
	// This is a simplified implementation for demo purposes
	
	// Mock user validation - replace with actual JWT validation
	mockUsers := map[string]*User{
		"admin-token": {
			ID:       1,
			Email:    "admin@example.com",
			Username: "admin",
			Role:     "admin",
			Active:   true,
		},
		"user-token": {
			ID:       2,
			Email:    "user@example.com",
			Username: "user",
			Role:     "user",
			Active:   true,
		},
	}
	
	if user, exists := mockUsers[token]; exists {
		user.LastLoginAt = time.Now()
		return user, nil
	}
	
	return nil, fmt.Errorf("invalid token")
}

// HasPermission checks if user has required permission
func (u *User) HasPermission(permission string) bool {
	switch u.Role {
	case "admin":
		return true // Admin has all permissions
	case "user":
		// User permissions
		userPermissions := []string{"read", "create", "update_own"}
		for _, p := range userPermissions {
			if p == permission {
				return true
			}
		}
	}
	return false
}

// MetricsCollector collects application metrics
type MetricsCollector struct {
	requestCount    int64
	responseTime    time.Duration
	errorCount      int64
	activeUsers     int64
	memoryUsage     int64
	uptimeStart     time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		uptimeStart: time.Now(),
	}
}

// IncrementRequestCount increments the request counter
func (m *MetricsCollector) IncrementRequestCount() {
	m.requestCount++
}

// RecordResponseTime records response time
func (m *MetricsCollector) RecordResponseTime(duration time.Duration) {
	m.responseTime = duration
}

// IncrementErrorCount increments error counter
func (m *MetricsCollector) IncrementErrorCount() {
	m.errorCount++
}

// GetMetrics returns current metrics
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	uptime := time.Since(m.uptimeStart)
	
	return map[string]interface{}{
		"requests_total":      m.requestCount,
		"errors_total":        m.errorCount,
		"response_time_ms":    m.responseTime.Milliseconds(),
		"active_users":        m.activeUsers,
		"memory_usage_mb":     m.memoryUsage / 1024 / 1024,
		"uptime_seconds":      uptime.Seconds(),
		"uptime_human":        uptime.String(),
	}
}

// HealthChecker performs health checks
type HealthChecker struct {
	config *Config
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(config *Config) *HealthChecker {
	return &HealthChecker{config: config}
}

// CheckHealth performs comprehensive health checks
func (h *HealthChecker) CheckHealth() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"environment": h.config.Environment,
		"checks": map[string]interface{}{
			"database":     h.checkDatabase(),
			"redis":        h.checkRedis(),
			"external_api": h.checkExternalAPI(),
			"disk_space":   h.checkDiskSpace(),
			"memory":       h.checkMemory(),
		},
	}
	
	// Check if any service is unhealthy
	allHealthy := true
	for _, check := range health["checks"].(map[string]interface{}) {
		if checkMap, ok := check.(map[string]interface{}); ok {
			if status, exists := checkMap["status"]; exists && status != "healthy" {
				allHealthy = false
				break
			}
		}
	}
	
	if !allHealthy {
		health["status"] = "unhealthy"
	}
	
	return health
}

func (h *HealthChecker) checkDatabase() map[string]interface{} {
	// Mock database check
	return map[string]interface{}{
		"status":       "healthy",
		"response_time": "5ms",
		"connections":  10,
	}
}

func (h *HealthChecker) checkRedis() map[string]interface{} {
	// Mock Redis check
	return map[string]interface{}{
		"status":       "healthy",
		"response_time": "2ms",
		"memory_usage": "15MB",
	}
}

func (h *HealthChecker) checkExternalAPI() map[string]interface{} {
	// Mock external API check
	return map[string]interface{}{
		"status":       "healthy",
		"response_time": "120ms",
		"last_check":   time.Now().Format(time.RFC3339),
	}
}

func (h *HealthChecker) checkDiskSpace() map[string]interface{} {
	// Mock disk space check
	return map[string]interface{}{
		"status":    "healthy",
		"usage":     "45%",
		"available": "120GB",
	}
}

func (h *HealthChecker) checkMemory() map[string]interface{} {
	// Mock memory check
	return map[string]interface{}{
		"status": "healthy",
		"usage":  "60%",
		"total":  "8GB",
	}
}

func main() {
	// Load configuration
	config := LoadConfig()
	
	// Initialize services
	authService := NewAuthService(config)
	metricsCollector := NewMetricsCollector()
	healthChecker := NewHealthChecker(config)
	
	// Create Zoox application
	app := zoox.New()
	
	// ================================
	// GLOBAL MIDDLEWARE STACK
	// ================================
	
	// Request ID middleware
	app.Use(middleware.RequestID())
	
	// Structured logging middleware
	app.Use(middleware.Logger())
	
	// Recovery middleware with custom error handling
	app.Use(middleware.Recovery())
	
	// CORS middleware with production settings
	app.Use(middleware.CORS(&middleware.CORSConfig{
		AllowOrigins:     config.CorsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "X-Response-Time"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	// Security headers middleware
	app.Use(middleware.Helmet(&middleware.HelmetConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline';",
	}))
	
	// Rate limiting middleware
	app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
		Rate:     float64(config.RateLimitRPS),
		Burst:    config.RateLimitRPS * 2,
		Duration: time.Minute,
	}))
	
	// Metrics collection middleware
	app.Use(func(ctx *zoox.Context) {
		start := time.Now()
		metricsCollector.IncrementRequestCount()
		
		ctx.Next()
		
		duration := time.Since(start)
		metricsCollector.RecordResponseTime(duration)
		ctx.Header("X-Response-Time", duration.String())
		
		if ctx.Status >= 400 {
			metricsCollector.IncrementErrorCount()
		}
	})
	
	// ================================
	// AUTHENTICATION MIDDLEWARE
	// ================================
	
	authMiddleware := func(ctx *zoox.Context) {
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "unauthorized",
				"message": "Authorization header required",
				"code":    "AUTH_001",
			})
			return
		}
		
		// Extract Bearer token
		token := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
				"code":    "AUTH_002",
			})
			return
		}
		
		// Validate token
		user, err := authService.ValidateJWT(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
				"code":    "AUTH_003",
			})
			return
		}
		
		if !user.Active {
			ctx.JSON(http.StatusForbidden, map[string]interface{}{
				"error":   "forbidden",
				"message": "Account is disabled",
				"code":    "AUTH_004",
			})
			return
		}
		
		ctx.Set("user", user)
		ctx.Next()
	}
	
	// Permission middleware
	requirePermission := func(permission string) zoox.HandlerFunc {
		return func(ctx *zoox.Context) {
			user, exists := ctx.Get("user")
			if !exists {
				ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "Authentication required",
					"code":    "PERM_001",
				})
				return
			}
			
			u := user.(User)
			if !u.HasPermission(permission) {
				ctx.JSON(http.StatusForbidden, map[string]interface{}{
					"error":   "forbidden",
					"message": fmt.Sprintf("Permission '%s' required", permission),
					"code":    "PERM_002",
				})
				return
			}
			
			ctx.Next()
		}
	}
	
	// ================================
	// PUBLIC ROUTES
	// ================================
	
	// Health check endpoint
	app.Get(config.HealthCheckPath, func(ctx *zoox.Context) {
		health := healthChecker.CheckHealth()
		
		status := http.StatusOK
		if health["status"] == "unhealthy" {
			status = http.StatusServiceUnavailable
		}
		
		ctx.JSON(status, health)
	})
	
	// Readiness probe (for Kubernetes)
	app.Get("/ready", func(ctx *zoox.Context) {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"status": "ready",
			"timestamp": time.Now().Unix(),
		})
	})
	
	// Liveness probe (for Kubernetes)
	app.Get("/live", func(ctx *zoox.Context) {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"status": "alive",
			"timestamp": time.Now().Unix(),
		})
	})
	
	// Metrics endpoint (for Prometheus)
	app.Get("/metrics", func(ctx *zoox.Context) {
		metrics := metricsCollector.GetMetrics()
		ctx.JSON(http.StatusOK, metrics)
	})
	
	// API documentation
	app.Get("/docs", func(ctx *zoox.Context) {
		docs := map[string]interface{}{
			"title":       "Production API",
			"version":     "1.0.0",
			"environment": config.Environment,
			"endpoints": map[string]interface{}{
				"health":   "GET " + config.HealthCheckPath,
				"metrics":  "GET /metrics",
				"auth":     "POST /api/v1/auth/login",
				"users":    "GET /api/v1/users (requires auth)",
				"admin":    "GET /api/v1/admin/* (requires admin role)",
			},
			"authentication": map[string]interface{}{
				"type":   "Bearer Token",
				"header": "Authorization: Bearer <token>",
				"tokens": map[string]string{
					"admin": "admin-token",
					"user":  "user-token",
				},
			},
		}
		
		ctx.JSON(http.StatusOK, docs)
	})
	
	// ================================
	// API ROUTES
	// ================================
	
	apiV1 := app.Group("/api/v1")
	
	// Authentication endpoints
	authGroup := apiV1.Group("/auth")
	
	authGroup.Post("/login", func(ctx *zoox.Context) {
		var loginReq struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		
		if err := ctx.BindJSON(&loginReq); err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "bad_request",
				"message": "Invalid JSON payload",
				"code":    "LOGIN_001",
			})
			return
		}
		
		// Mock authentication - in production, verify against database
		var token string
		var user *User
		
		switch loginReq.Email {
		case "admin@example.com":
			if loginReq.Password == "admin123" {
				token = "admin-token"
				user = &User{
					ID: 1, Email: "admin@example.com", Username: "admin", Role: "admin", Active: true,
				}
			}
		case "user@example.com":
			if loginReq.Password == "user123" {
				token = "user-token"
				user = &User{
					ID: 2, Email: "user@example.com", Username: "user", Role: "user", Active: true,
				}
			}
		}
		
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "unauthorized",
				"message": "Invalid credentials",
				"code":    "LOGIN_002",
			})
			return
		}
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
			"user":  user,
			"expires_in": 3600,
		})
	})
	
	// Protected routes
	protected := apiV1.Group("/", authMiddleware)
	
	// User profile
	protected.Get("/profile", func(ctx *zoox.Context) {
		user := ctx.Get("user").(User)
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"user": user,
		})
	})
	
	// Users endpoint
	protected.Get("/users", requirePermission("read"), func(ctx *zoox.Context) {
		// Mock users data
		users := []User{
			{ID: 1, Email: "admin@example.com", Username: "admin", Role: "admin", Active: true},
			{ID: 2, Email: "user@example.com", Username: "user", Role: "user", Active: true},
		}
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"users": users,
			"count": len(users),
		})
	})
	
	// Admin routes
	adminGroup := protected.Group("/admin", requirePermission("admin"))
	
	adminGroup.Get("/stats", func(ctx *zoox.Context) {
		stats := map[string]interface{}{
			"total_users":     2,
			"active_sessions": 1,
			"system_load":     "0.8",
			"memory_usage":    "512MB",
			"disk_usage":      "45%",
		}
		
		ctx.JSON(http.StatusOK, stats)
	})
	
	adminGroup.Get("/logs", func(ctx *zoox.Context) {
		// Mock recent logs
		logs := []map[string]interface{}{
			{
				"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				"level":     "INFO",
				"message":   "User login successful",
				"user_id":   2,
			},
			{
				"timestamp": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
				"level":     "WARN",
				"message":   "Rate limit exceeded",
				"ip":        "192.168.1.100",
			},
		}
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"logs": logs,
		})
	})
	
	// ================================
	// SERVER SETUP AND GRACEFUL SHUTDOWN
	// ================================
	
	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      app,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Production API starting...")
		log.Printf("📍 Server running on http://localhost:%s", config.Port)
		log.Printf("🏥 Health Check: http://localhost:%s%s", config.Port, config.HealthCheckPath)
		log.Printf("📊 Metrics: http://localhost:%s/metrics", config.Port)
		log.Printf("📚 Documentation: http://localhost:%s/docs", config.Port)
		log.Printf("🌍 Environment: %s", config.Environment)
		
		if config.TLSCertFile != "" && config.TLSKeyFile != "" {
			log.Printf("🔒 TLS enabled")
			if err := srv.ListenAndServeTLS(config.TLSCertFile, config.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}
	}()
	
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("🛑 Shutting down server...")
	
	// The context is used to inform the server it has 30 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	
	log.Println("✅ Server exited gracefully")
} 