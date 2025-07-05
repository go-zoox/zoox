# Tutorial 16: Security Best Practices

## Overview
Learn essential security practices for Zoox applications, including authentication, authorization, input validation, and protection against common vulnerabilities.

## Learning Objectives
- Implement secure authentication mechanisms
- Apply proper authorization controls
- Validate and sanitize user inputs
- Protect against OWASP Top 10 vulnerabilities
- Secure API endpoints and data transmission
- Monitor security events and incidents

## Prerequisites
- Complete Tutorial 15: Performance Optimization
- Understanding of web security concepts
- Knowledge of common attack vectors

## Authentication Security

### JWT Implementation with Security

```go
package main

import (
    "crypto/rand"
    "encoding/hex"
    "errors"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/go-zoox/zoox"
    "golang.org/x/crypto/bcrypt"
)

// Security configuration
type SecurityConfig struct {
    JWTSecret           []byte
    TokenExpiry         time.Duration
    RefreshTokenExpiry  time.Duration
    BcryptCost         int
    MaxLoginAttempts   int
    LockoutDuration    time.Duration
}

// User with security fields
type SecureUser struct {
    ID              int       `json:"id"`
    Username        string    `json:"username"`
    Email           string    `json:"email"`
    PasswordHash    string    `json:"-"`
    Role            string    `json:"role"`
    LoginAttempts   int       `json:"-"`
    LockedUntil     time.Time `json:"-"`
    LastLogin       time.Time `json:"last_login"`
    TwoFactorSecret string    `json:"-"`
    TwoFactorEnabled bool     `json:"two_factor_enabled"`
    Created         time.Time `json:"created"`
    Updated         time.Time `json:"updated"`
}

// JWT Claims
type JWTClaims struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// Secure authentication service
type AuthService struct {
    users  map[string]*SecureUser
    config SecurityConfig
}

func NewAuthService(config SecurityConfig) *AuthService {
    return &AuthService{
        users:  make(map[string]*SecureUser),
        config: config,
    }
}

func (s *AuthService) HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.config.BcryptCost)
    return string(bytes), err
}

func (s *AuthService) CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func (s *AuthService) GenerateToken(user *SecureUser) (string, error) {
    claims := JWTClaims{
        UserID:   user.ID,
        Username: user.Username,
        Role:     user.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.TokenExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "zoox-app",
            Subject:   user.Username,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.config.JWTSecret)
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return s.config.JWTSecret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}

func (s *AuthService) Register(username, email, password string) (*SecureUser, error) {
    // Check if user exists
    if _, exists := s.users[username]; exists {
        return nil, errors.New("username already exists")
    }

    // Validate password strength
    if err := s.validatePasswordStrength(password); err != nil {
        return nil, err
    }

    // Hash password
    hashedPassword, err := s.HashPassword(password)
    if err != nil {
        return nil, err
    }

    user := &SecureUser{
        ID:           len(s.users) + 1,
        Username:     username,
        Email:        email,
        PasswordHash: hashedPassword,
        Role:         "user",
        Created:      time.Now(),
        Updated:      time.Now(),
    }

    s.users[username] = user
    return user, nil
}

func (s *AuthService) Login(username, password string) (*SecureUser, string, error) {
    user, exists := s.users[username]
    if !exists {
        return nil, "", errors.New("invalid credentials")
    }

    // Check if account is locked
    if time.Now().Before(user.LockedUntil) {
        return nil, "", errors.New("account locked due to too many failed attempts")
    }

    // Validate password
    if !s.CheckPassword(password, user.PasswordHash) {
        user.LoginAttempts++
        if user.LoginAttempts >= s.config.MaxLoginAttempts {
            user.LockedUntil = time.Now().Add(s.config.LockoutDuration)
        }
        return nil, "", errors.New("invalid credentials")
    }

    // Reset login attempts on successful login
    user.LoginAttempts = 0
    user.LastLogin = time.Now()

    // Generate token
    token, err := s.GenerateToken(user)
    if err != nil {
        return nil, "", err
    }

    return user, token, nil
}

func (s *AuthService) validatePasswordStrength(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters long")
    }

    hasUpper := false
    hasLower := false
    hasDigit := false
    hasSpecial := false

    for _, char := range password {
        switch {
        case 'A' <= char && char <= 'Z':
            hasUpper = true
        case 'a' <= char && char <= 'z':
            hasLower = true
        case '0' <= char && char <= '9':
            hasDigit = true
        default:
            hasSpecial = true
        }
    }

    if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
        return errors.New("password must contain uppercase, lowercase, digit, and special character")
    }

    return nil
}

// Security middleware
func authMiddleware(authService *AuthService) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        authHeader := ctx.Header("Authorization")
        if authHeader == "" {
            ctx.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Authorization header required",
            })
            ctx.Abort()
            return
        }

        tokenParts := strings.Split(authHeader, " ")
        if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
            ctx.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Invalid authorization header format",
            })
            ctx.Abort()
            return
        }

        claims, err := authService.ValidateToken(tokenParts[1])
        if err != nil {
            ctx.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Invalid token",
            })
            ctx.Abort()
            return
        }

        // Set user context
        ctx.Set("user_id", claims.UserID)
        ctx.Set("username", claims.Username)
        ctx.Set("role", claims.Role)

        ctx.Next()
    }
}

// Role-based authorization middleware
func requireRole(role string) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        userRole, exists := ctx.Get("role")
        if !exists {
            ctx.JSON(http.StatusForbidden, map[string]string{
                "error": "No role information found",
            })
            ctx.Abort()
            return
        }

        if userRole != role && userRole != "admin" {
            ctx.JSON(http.StatusForbidden, map[string]string{
                "error": "Insufficient permissions",
            })
            ctx.Abort()
            return
        }

        ctx.Next()
    }
}
```

## Input Validation and Sanitization

```go
import (
    "html"
    "regexp"
    "strings"
    "unicode"
)

// Input validator
type InputValidator struct {
    emailRegex    *regexp.Regexp
    phoneRegex    *regexp.Regexp
    usernameRegex *regexp.Regexp
}

func NewInputValidator() *InputValidator {
    return &InputValidator{
        emailRegex:    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
        phoneRegex:    regexp.MustCompile(`^\+?[1-9]\d{1,14}$`),
        usernameRegex: regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`),
    }
}

func (v *InputValidator) ValidateEmail(email string) error {
    if !v.emailRegex.MatchString(email) {
        return errors.New("invalid email format")
    }
    return nil
}

func (v *InputValidator) ValidateUsername(username string) error {
    if !v.usernameRegex.MatchString(username) {
        return errors.New("username must be 3-20 characters, alphanumeric or underscore only")
    }
    return nil
}

func (v *InputValidator) SanitizeHTML(input string) string {
    return html.EscapeString(input)
}

func (v *InputValidator) SanitizeString(input string) string {
    // Remove non-printable characters
    result := strings.Map(func(r rune) rune {
        if unicode.IsPrint(r) {
            return r
        }
        return -1
    }, input)
    
    // Trim whitespace
    return strings.TrimSpace(result)
}

func (v *InputValidator) ValidateStringLength(input string, minLen, maxLen int) error {
    length := len(strings.TrimSpace(input))
    if length < minLen {
        return errors.New(fmt.Sprintf("input too short, minimum %d characters", minLen))
    }
    if length > maxLen {
        return errors.New(fmt.Sprintf("input too long, maximum %d characters", maxLen))
    }
    return nil
}

// Input validation middleware
func validationMiddleware(validator *InputValidator) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        // Get content type
        contentType := ctx.Header("Content-Type")
        
        // Validate JSON content type for POST/PUT requests
        if (ctx.Method() == http.MethodPost || ctx.Method() == http.MethodPut) &&
           !strings.Contains(contentType, "application/json") {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Content-Type must be application/json",
            })
            ctx.Abort()
            return
        }

        ctx.Next()
    }
}
```

## SQL Injection Prevention

```go
import (
    "database/sql"
    "fmt"
)

// Safe database operations
type SafeDB struct {
    db *sql.DB
}

func NewSafeDB(db *sql.DB) *SafeDB {
    return &SafeDB{db: db}
}

// Always use prepared statements
func (sdb *SafeDB) GetUserByID(userID int) (*SecureUser, error) {
    query := "SELECT id, username, email, role, created FROM users WHERE id = ?"
    
    var user SecureUser
    err := sdb.db.QueryRow(query, userID).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.Role,
        &user.Created,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("user not found")
        }
        return nil, err
    }
    
    return &user, nil
}

func (sdb *SafeDB) SearchUsers(searchTerm string, limit int) ([]*SecureUser, error) {
    // Parameterized query to prevent SQL injection
    query := `
        SELECT id, username, email, role, created 
        FROM users 
        WHERE username LIKE ? OR email LIKE ? 
        ORDER BY username 
        LIMIT ?
    `
    
    searchPattern := "%" + searchTerm + "%"
    
    rows, err := sdb.db.Query(query, searchPattern, searchPattern, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []*SecureUser
    for rows.Next() {
        var user SecureUser
        err := rows.Scan(
            &user.ID,
            &user.Username,
            &user.Email,
            &user.Role,
            &user.Created,
        )
        if err != nil {
            return nil, err
        }
        users = append(users, &user)
    }
    
    return users, rows.Err()
}
```

## CORS and Security Headers

```go
// Security headers middleware
func securityHeadersMiddleware() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        // Prevent XSS attacks
        ctx.Set("X-Content-Type-Options", "nosniff")
        ctx.Set("X-Frame-Options", "DENY")
        ctx.Set("X-XSS-Protection", "1; mode=block")
        
        // HSTS (HTTPS only)
        ctx.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        // Content Security Policy
        ctx.Set("Content-Security-Policy", 
            "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
        
        // Referrer Policy
        ctx.Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Remove server information
        ctx.Set("Server", "")
        
        ctx.Next()
    }
}

// CORS middleware with security
func corsMiddleware(allowedOrigins []string) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        origin := ctx.Header("Origin")
        
        // Check if origin is allowed
        allowed := false
        for _, allowedOrigin := range allowedOrigins {
            if origin == allowedOrigin {
                allowed = true
                break
            }
        }
        
        if allowed {
            ctx.Set("Access-Control-Allow-Origin", origin)
        }
        
        ctx.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        ctx.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        ctx.Set("Access-Control-Allow-Credentials", "true")
        ctx.Set("Access-Control-Max-Age", "86400")
        
        // Handle preflight requests
        if ctx.Method() == http.MethodOptions {
            ctx.Status(http.StatusNoContent)
            ctx.Abort()
            return
        }
        
        ctx.Next()
    }
}
```

## Rate Limiting for Security

```go
// Advanced rate limiter with different limits per endpoint
type EndpointRateLimiter struct {
    limiters map[string]*RateLimiter
    mutex    sync.RWMutex
}

func NewEndpointRateLimiter() *EndpointRateLimiter {
    return &EndpointRateLimiter{
        limiters: make(map[string]*RateLimiter),
    }
}

func (erl *EndpointRateLimiter) AddEndpoint(endpoint string, limit int, window time.Duration) {
    erl.mutex.Lock()
    defer erl.mutex.Unlock()
    
    erl.limiters[endpoint] = NewRateLimiter(limit, window)
}

func (erl *EndpointRateLimiter) Allow(endpoint, clientID string) bool {
    erl.mutex.RLock()
    limiter, exists := erl.limiters[endpoint]
    erl.mutex.RUnlock()
    
    if !exists {
        return true // No rate limit for this endpoint
    }
    
    return limiter.Allow(clientID)
}

// Security-focused rate limiting middleware
func securityRateLimitMiddleware(limiter *EndpointRateLimiter) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        clientID := ctx.ClientIP()
        endpoint := ctx.Request().URL.Path
        
        // Apply stricter limits to sensitive endpoints
        if strings.Contains(endpoint, "/login") || strings.Contains(endpoint, "/register") {
            if !limiter.Allow(endpoint, clientID) {
                // Log potential brute force attack
                log.Printf("Rate limit exceeded for sensitive endpoint %s from IP %s", endpoint, clientID)
                
                ctx.JSON(http.StatusTooManyRequests, map[string]string{
                    "error": "Too many attempts, please try again later",
                })
                ctx.Abort()
                return
            }
        }
        
        ctx.Next()
    }
}
```

## Secure File Upload

```go
import (
    "crypto/sha256"
    "io"
    "mime/multipart"
    "path/filepath"
)

type FileUploadConfig struct {
    MaxFileSize     int64
    AllowedTypes    []string
    UploadDirectory string
    ScanForViruses  bool
}

type SecureFileUpload struct {
    config FileUploadConfig
}

func NewSecureFileUpload(config FileUploadConfig) *SecureFileUpload {
    return &SecureFileUpload{config: config}
}

func (sfu *SecureFileUpload) ValidateFile(header *multipart.FileHeader) error {
    // Check file size
    if header.Size > sfu.config.MaxFileSize {
        return errors.New("file too large")
    }
    
    // Check file extension
    ext := strings.ToLower(filepath.Ext(header.Filename))
    allowed := false
    for _, allowedType := range sfu.config.AllowedTypes {
        if ext == allowedType {
            allowed = true
            break
        }
    }
    
    if !allowed {
        return errors.New("file type not allowed")
    }
    
    return nil
}

func (sfu *SecureFileUpload) SaveFile(file multipart.File, header *multipart.FileHeader) (string, error) {
    // Generate secure filename
    hash := sha256.New()
    io.Copy(hash, file)
    file.Seek(0, 0) // Reset file pointer
    
    hashString := hex.EncodeToString(hash.Sum(nil))
    ext := filepath.Ext(header.Filename)
    filename := hashString + ext
    
    // Create secure file path
    filepath := filepath.Join(sfu.config.UploadDirectory, filename)
    
    // Save file with restricted permissions
    dest, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
    if err != nil {
        return "", err
    }
    defer dest.Close()
    
    _, err = io.Copy(dest, file)
    if err != nil {
        return "", err
    }
    
    return filename, nil
}

// Secure file upload handler
func secureUploadHandler(uploader *SecureFileUpload) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        // Parse multipart form
        err := ctx.Request().ParseMultipartForm(uploader.config.MaxFileSize)
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Invalid multipart form",
            })
            return
        }
        
        file, header, err := ctx.Request().FormFile("file")
        if err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "No file provided",
            })
            return
        }
        defer file.Close()
        
        // Validate file
        if err := uploader.ValidateFile(header); err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": err.Error(),
            })
            return
        }
        
        // Save file
        filename, err := uploader.SaveFile(file, header)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Failed to save file",
            })
            return
        }
        
        ctx.JSON(http.StatusOK, map[string]string{
            "message":  "File uploaded successfully",
            "filename": filename,
        })
    }
}
```

## Security Logging and Monitoring

```go
// Security event types
type SecurityEventType string

const (
    LoginAttempt       SecurityEventType = "login_attempt"
    LoginSuccess       SecurityEventType = "login_success"
    LoginFailure       SecurityEventType = "login_failure"
    UnauthorizedAccess SecurityEventType = "unauthorized_access"
    RateLimitExceeded  SecurityEventType = "rate_limit_exceeded"
    SuspiciousActivity SecurityEventType = "suspicious_activity"
)

// Security event
type SecurityEvent struct {
    Type        SecurityEventType `json:"type"`
    UserID      int              `json:"user_id,omitempty"`
    Username    string           `json:"username,omitempty"`
    IP          string           `json:"ip"`
    UserAgent   string           `json:"user_agent"`
    Endpoint    string           `json:"endpoint"`
    Timestamp   time.Time        `json:"timestamp"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Severity    string           `json:"severity"`
}

// Security logger
type SecurityLogger struct {
    events []SecurityEvent
    mutex  sync.Mutex
}

func NewSecurityLogger() *SecurityLogger {
    return &SecurityLogger{
        events: make([]SecurityEvent, 0),
    }
}

func (sl *SecurityLogger) LogEvent(event SecurityEvent) {
    sl.mutex.Lock()
    defer sl.mutex.Unlock()
    
    event.Timestamp = time.Now()
    sl.events = append(sl.events, event)
    
    // Log to console/file
    log.Printf("SECURITY EVENT: %s - %s from %s", event.Type, event.Username, event.IP)
    
    // Keep only last 1000 events
    if len(sl.events) > 1000 {
        sl.events = sl.events[len(sl.events)-1000:]
    }
}

func (sl *SecurityLogger) GetEvents(limit int) []SecurityEvent {
    sl.mutex.Lock()
    defer sl.mutex.Unlock()
    
    if limit > len(sl.events) {
        limit = len(sl.events)
    }
    
    start := len(sl.events) - limit
    return sl.events[start:]
}

// Security monitoring middleware
func securityMonitoringMiddleware(logger *SecurityLogger) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        start := time.Now()
        
        ctx.Next()
        
        // Log security events based on response
        status := ctx.Writer.Status()
        if status == http.StatusUnauthorized || status == http.StatusForbidden {
            logger.LogEvent(SecurityEvent{
                Type:      UnauthorizedAccess,
                IP:        ctx.ClientIP(),
                UserAgent: ctx.Header("User-Agent"),
                Endpoint:  ctx.Request().URL.Path,
                Severity:  "medium",
                Details: map[string]interface{}{
                    "status_code": status,
                    "duration":    time.Since(start).String(),
                },
            })
        }
    }
}
```

## Complete Secure Application Example

```go
func main() {
    // Security configuration
    config := SecurityConfig{
        JWTSecret:          []byte("your-super-secret-jwt-key-change-this"),
        TokenExpiry:        24 * time.Hour,
        RefreshTokenExpiry: 7 * 24 * time.Hour,
        BcryptCost:         12,
        MaxLoginAttempts:   5,
        LockoutDuration:    15 * time.Minute,
    }
    
    // Initialize services
    authService := NewAuthService(config)
    validator := NewInputValidator()
    secLogger := NewSecurityLogger()
    rateLimiter := NewEndpointRateLimiter()
    
    // Configure rate limits
    rateLimiter.AddEndpoint("/login", 5, time.Minute)
    rateLimiter.AddEndpoint("/register", 3, time.Minute)
    
    app := zoox.New()
    
    // Security middleware stack
    app.Use(securityHeadersMiddleware())
    app.Use(corsMiddleware([]string{"https://yourdomain.com"}))
    app.Use(securityRateLimitMiddleware(rateLimiter))
    app.Use(validationMiddleware(validator))
    app.Use(securityMonitoringMiddleware(secLogger))
    
    // Public routes
    app.Post("/register", registerHandler(authService, validator))
    app.Post("/login", loginHandler(authService, secLogger))
    
    // Protected routes
    protected := app.Group("/api")
    protected.Use(authMiddleware(authService))
    
    protected.Get("/profile", getProfileHandler())
    protected.Put("/profile", updateProfileHandler(validator))
    
    // Admin routes
    admin := protected.Group("/admin")
    admin.Use(requireRole("admin"))
    
    admin.Get("/users", listUsersHandler())
    admin.Get("/security-events", getSecurityEventsHandler(secLogger))
    
    fmt.Println("Secure server starting on :8443 (HTTPS)")
    log.Fatal(app.ListenTLS(":8443", "server.crt", "server.key"))
}

// Secure handlers implementation would go here...
```

## Key Security Takeaways

1. **Authentication**: Use strong password policies and secure token management
2. **Authorization**: Implement role-based access control
3. **Input Validation**: Validate and sanitize all user inputs
4. **SQL Injection**: Always use parameterized queries
5. **XSS Protection**: Escape output and use security headers
6. **CSRF Protection**: Implement CSRF tokens for state-changing operations
7. **HTTPS**: Always use HTTPS in production
8. **Rate Limiting**: Protect against brute force and DoS attacks
9. **Security Monitoring**: Log and monitor security events
10. **Regular Updates**: Keep dependencies updated and scan for vulnerabilities

## Next Steps

- Tutorial 17: Deployment Strategies - Deploy secure applications
- Tutorial 18: Production Monitoring - Monitor production systems
- Implement security scanning in CI/CD
- Learn about penetration testing
- Study OWASP guidelines and best practices

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Guidelines](https://golang.org/doc/security)
- [JWT Best Practices](https://tools.ietf.org/html/draft-ietf-oauth-jwt-bcp-07)
- [Secure Coding Practices](https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/) 