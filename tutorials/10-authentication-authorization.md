# Tutorial 10: Authentication & Authorization

## 📖 Overview

Learn to implement secure authentication and authorization in Zoox applications. This tutorial covers JWT tokens, session management, role-based access control, and security best practices.

## 🎯 Learning Objectives

- Implement JWT authentication
- Build session management
- Create role-based access control
- Secure API endpoints
- Handle authentication middleware

## 📋 Prerequisites

- Completed [Tutorial 01: Getting Started](./01-getting-started.md)
- Understanding of authentication concepts
- Basic knowledge of security principles

## 🚀 Getting Started

### JWT Authentication System

```go
package main

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
    
    "github.com/dgrijalva/jwt-go"
    "github.com/go-zoox/zoox"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID       int      `json:"id"`
    Username string   `json:"username"`
    Email    string   `json:"email"`
    Password string   `json:"-"`
    Roles    []string `json:"roles"`
    Active   bool     `json:"active"`
}

type AuthService struct {
    users     map[string]*User
    jwtSecret []byte
    nextID    int
}

func NewAuthService() *AuthService {
    secret := make([]byte, 32)
    rand.Read(secret)
    
    return &AuthService{
        users:     make(map[string]*User),
        jwtSecret: secret,
        nextID:    1,
    }
}

func (as *AuthService) Register(username, email, password string) (*User, error) {
    // Check if user exists
    if _, exists := as.users[username]; exists {
        return nil, fmt.Errorf("user already exists")
    }
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    
    user := &User{
        ID:       as.nextID,
        Username: username,
        Email:    email,
        Password: string(hashedPassword),
        Roles:    []string{"user"},
        Active:   true,
    }
    
    as.users[username] = user
    as.nextID++
    
    return user, nil
}

func (as *AuthService) Login(username, password string) (string, *User, error) {
    user, exists := as.users[username]
    if !exists || !user.Active {
        return "", nil, fmt.Errorf("invalid credentials")
    }
    
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return "", nil, fmt.Errorf("invalid credentials")
    }
    
    // Generate JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":  user.ID,
        "username": user.Username,
        "roles":    user.Roles,
        "exp":      time.Now().Add(24 * time.Hour).Unix(),
    })
    
    tokenString, err := token.SignedString(as.jwtSecret)
    if err != nil {
        return "", nil, err
    }
    
    return tokenString, user, nil
}

func (as *AuthService) ValidateToken(tokenString string) (*User, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return as.jwtSecret, nil
    })
    
    if err != nil || !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, fmt.Errorf("invalid token claims")
    }
    
    username, ok := claims["username"].(string)
    if !ok {
        return nil, fmt.Errorf("invalid username in token")
    }
    
    user, exists := as.users[username]
    if !exists || !user.Active {
        return nil, fmt.Errorf("user not found or inactive")
    }
    
    return user, nil
}

func (as *AuthService) AuthMiddleware() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        authHeader := ctx.Header("Authorization")
        if authHeader == "" {
            ctx.JSON(401, map[string]string{"error": "Authorization header required"})
            return
        }
        
        // Extract token from "Bearer <token>"
        if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
            ctx.JSON(401, map[string]string{"error": "Invalid authorization format"})
            return
        }
        
        tokenString := authHeader[7:]
        user, err := as.ValidateToken(tokenString)
        if err != nil {
            ctx.JSON(401, map[string]string{"error": "Invalid token"})
            return
        }
        
        ctx.Set("user", user)
        ctx.Next()
    }
}

func (as *AuthService) RequireRole(role string) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        user, exists := ctx.Get("user")
        if !exists {
            ctx.JSON(401, map[string]string{"error": "Authentication required"})
            return
        }
        
        u := user.(*User)
        for _, userRole := range u.Roles {
            if userRole == role || userRole == "admin" {
                ctx.Next()
                return
            }
        }
        
        ctx.JSON(403, map[string]string{"error": "Insufficient permissions"})
    }
}

func main() {
    app := zoox.New()
    
    authService := NewAuthService()
    
    // Create default admin user
    authService.Register("admin", "admin@example.com", "admin123")
    if user, exists := authService.users["admin"]; exists {
        user.Roles = []string{"admin", "user"}
    }
    
    // Public endpoints
    app.Post("/register", func(ctx *zoox.Context) {
        var req struct {
            Username string `json:"username"`
            Email    string `json:"email"`
            Password string `json:"password"`
        }
        
        if err := ctx.BindJSON(&req); err != nil {
            ctx.JSON(400, map[string]string{"error": "Invalid request"})
            return
        }
        
        user, err := authService.Register(req.Username, req.Email, req.Password)
        if err != nil {
            ctx.JSON(400, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.JSON(201, map[string]interface{}{
            "message": "User registered successfully",
            "user":    user,
        })
    })
    
    app.Post("/login", func(ctx *zoox.Context) {
        var req struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        
        if err := ctx.BindJSON(&req); err != nil {
            ctx.JSON(400, map[string]string{"error": "Invalid request"})
            return
        }
        
        token, user, err := authService.Login(req.Username, req.Password)
        if err != nil {
            ctx.JSON(401, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.JSON(200, map[string]interface{}{
            "token": token,
            "user":  user,
        })
    })
    
    // Protected endpoints
    protected := app.Group("/api")
    protected.Use(authService.AuthMiddleware())
    
    protected.Get("/profile", func(ctx *zoox.Context) {
        user := ctx.Get("user").(*User)
        ctx.JSON(200, user)
    })
    
    protected.Get("/users", authService.RequireRole("admin"), func(ctx *zoox.Context) {
        users := make([]*User, 0, len(authService.users))
        for _, user := range authService.users {
            users = append(users, user)
        }
        ctx.JSON(200, users)
    })
    
    // Login form
    app.Get("/", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Authentication Demo</title>
            <style>
                body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
                .form-group { margin: 15px 0; }
                label { display: block; margin-bottom: 5px; }
                input { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
                button { width: 100%; padding: 10px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
                .result { margin-top: 20px; padding: 10px; border-radius: 4px; }
                .success { background: #d4edda; color: #155724; }
                .error { background: #f8d7da; color: #721c24; }
                .protected { margin-top: 20px; }
            </style>
        </head>
        <body>
            <h1>Authentication Demo</h1>
            
            <div class="form-group">
                <label>Username:</label>
                <input type="text" id="username" value="admin">
            </div>
            
            <div class="form-group">
                <label>Password:</label>
                <input type="password" id="password" value="admin123">
            </div>
            
            <button onclick="login()">Login</button>
            
            <div id="result" class="result" style="display: none;"></div>
            
            <div class="protected">
                <h3>Protected Actions</h3>
                <button onclick="getProfile()" disabled id="profileBtn">Get Profile</button>
                <button onclick="getUsers()" disabled id="usersBtn">Get Users (Admin)</button>
            </div>
            
            <script>
                let token = '';
                
                async function login() {
                    const username = document.getElementById('username').value;
                    const password = document.getElementById('password').value;
                    
                    try {
                        const response = await fetch('/login', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({ username, password })
                        });
                        
                        const data = await response.json();
                        const result = document.getElementById('result');
                        
                        if (response.ok) {
                            token = data.token;
                            result.className = 'result success';
                            result.innerHTML = 'Login successful! Welcome ' + data.user.username;
                            result.style.display = 'block';
                            
                            // Enable protected buttons
                            document.getElementById('profileBtn').disabled = false;
                            document.getElementById('usersBtn').disabled = false;
                        } else {
                            result.className = 'result error';
                            result.innerHTML = 'Login failed: ' + data.error;
                            result.style.display = 'block';
                        }
                    } catch (error) {
                        console.error('Login error:', error);
                    }
                }
                
                async function getProfile() {
                    try {
                        const response = await fetch('/api/profile', {
                            headers: {
                                'Authorization': 'Bearer ' + token
                            }
                        });
                        
                        const data = await response.json();
                        alert('Profile: ' + JSON.stringify(data, null, 2));
                    } catch (error) {
                        console.error('Profile error:', error);
                    }
                }
                
                async function getUsers() {
                    try {
                        const response = await fetch('/api/users', {
                            headers: {
                                'Authorization': 'Bearer ' + token
                            }
                        });
                        
                        const data = await response.json();
                        alert('Users: ' + JSON.stringify(data, null, 2));
                    } catch (error) {
                        console.error('Users error:', error);
                    }
                }
            </script>
        </body>
        </html>
        `
        ctx.HTML(200, html, nil)
    })
    
    app.Listen(":8080")
}
```

## 📚 Key Takeaways

1. **JWT Tokens**: Secure stateless authentication
2. **Password Security**: Hash passwords with bcrypt
3. **Role-Based Access**: Implement granular permissions
4. **Middleware**: Use authentication middleware for protection
5. **Security**: Follow security best practices

## 🎯 Next Steps

- Learn [Tutorial 11: Database Integration](./11-database-integration.md)
- Explore [Tutorial 12: Caching Strategies](./12-caching-strategies.md)
- Study [Tutorial 15: Security Best Practices](./15-security-best-practices.md)

---

**Congratulations!** You've mastered authentication and authorization in Zoox! 