# JWT 认证

Zoox 提供了完整的 JWT（JSON Web Token）支持，可以轻松实现身份认证和授权。

## 基本用法

### 配置 JWT

```go
app := zoox.New()

// 设置密钥（用于签名和验证 JWT）
app.Config.SecretKey = "your-secret-key"
```

**说明**: JWT 实现参考 `context.go:1035-1047`。

### 生成 JWT Token

```go
app.Post("/login", func(ctx *zoox.Context) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := ctx.BindJSON(&creds); err != nil {
		ctx.Error(400, "Invalid JSON")
		return
	}
	
	// 验证用户名密码
	if creds.Username == "admin" && creds.Password == "password" {
		// 生成 JWT Token
		jwt := ctx.Jwt()
		token, err := jwt.Sign(map[string]interface{}{
			"user_id":  1,
			"username": "admin",
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})
		
		if err != nil {
			ctx.Error(500, "Failed to generate token")
			return
		}
		
		ctx.JSON(200, zoox.H{
			"token": token,
		})
	} else {
		ctx.Error(401, "Invalid credentials")
	}
})
```

### 验证 JWT Token

```go
app.Get("/profile", func(ctx *zoox.Context) {
	// 获取 Bearer Token
	token, ok := ctx.BearerToken()
	if !ok {
		ctx.Error(401, "No token provided")
		return
	}
	
	// 验证 Token
	jwt := ctx.Jwt()
	claims, err := jwt.Verify(token)
	if err != nil {
		ctx.Error(401, "Invalid token")
		return
	}
	
	// 使用 Claims
	ctx.JSON(200, zoox.H{
		"user_id":  claims["user_id"],
		"username": claims["username"],
	})
})
```

## JWT 中间件

使用内置的 JWT 中间件自动验证：

```go
import "github.com/go-zoox/zoox/middleware"

app := zoox.New()

// 为路由组添加 JWT 中间件
api := app.Group("/api")
api.Use(middleware.JWT())

// 受保护的路由
api.Get("/profile", func(ctx *zoox.Context) {
	// 如果到达这里，说明 Token 已验证
	token, _ := ctx.BearerToken()
	jwt := ctx.Jwt()
	claims, _ := jwt.Verify(token)
	
	ctx.JSON(200, claims)
})
```

**说明**: JWT 中间件参考 `middleware/jwt.go`。

## 从查询参数获取 Token

JWT 中间件也支持从查询参数获取 Token：

```go
// 客户端可以这样访问
// GET /api/profile?access_token=your-token-here

// 中间件会自动从查询参数或 Bearer Token 中获取
```

## 完整示例

### 登录和认证流程

```go
package main

import (
	"time"
	
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	
	// 配置密钥
	app.Config.SecretKey = "your-secret-key"
	
	// 登录接口（不需要认证）
	app.Post("/login", func(ctx *zoox.Context) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		
		if err := ctx.BindJSON(&creds); err != nil {
			ctx.Error(400, "Invalid JSON")
			return
		}
		
		// 验证用户名密码
		if creds.Username == "admin" && creds.Password == "password" {
			// 生成 JWT Token
			jwt := ctx.Jwt()
			token, err := jwt.Sign(map[string]interface{}{
				"user_id":  1,
				"username": "admin",
				"role":     "admin",
				"exp":      time.Now().Add(24 * time.Hour).Unix(),
				"iat":      time.Now().Unix(),
			})
			
			if err != nil {
				ctx.Error(500, "Failed to generate token")
				return
			}
			
			ctx.JSON(200, zoox.H{
				"token": token,
				"expires_in": 86400, // 24小时
			})
		} else {
			ctx.Error(401, "Invalid credentials")
		}
	})
	
	// 受保护的 API 路由组
	api := app.Group("/api")
	api.Use(middleware.JWT())
	
	// 获取当前用户信息
	api.Get("/me", func(ctx *zoox.Context) {
		token, _ := ctx.BearerToken()
		jwt := ctx.Jwt()
		claims, _ := jwt.Verify(token)
		
		ctx.JSON(200, zoox.H{
			"user_id":  claims["user_id"],
			"username": claims["username"],
			"role":     claims["role"],
		})
	})
	
	// 刷新 Token
	api.Post("/refresh", func(ctx *zoox.Context) {
		token, _ := ctx.BearerToken()
		jwt := ctx.Jwt()
		claims, err := jwt.Verify(token)
		if err != nil {
			ctx.Error(401, "Invalid token")
			return
		}
		
		// 生成新 Token
		newToken, err := jwt.Sign(map[string]interface{}{
			"user_id":  claims["user_id"],
			"username": claims["username"],
			"role":     claims["role"],
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
			"iat":      time.Now().Unix(),
		})
		
		if err != nil {
			ctx.Error(500, "Failed to generate token")
			return
		}
		
		ctx.JSON(200, zoox.H{
			"token": newToken,
		})
	})
	
	app.Run(":8080")
}
```

## Token 刷新

实现 Token 刷新机制：

```go
api.Post("/refresh", func(ctx *zoox.Context) {
	// 获取旧的 Token
	token, _ := ctx.BearerToken()
	jwt := ctx.Jwt()
	
	// 验证旧 Token（即使过期也要能验证）
	claims, err := jwt.Verify(token)
	if err != nil {
		ctx.Error(401, "Invalid token")
		return
	}
	
	// 检查是否在刷新窗口内（例如：过期后30分钟内）
	exp := int64(claims["exp"].(float64))
	if time.Now().Unix() > exp+1800 { // 30分钟
		ctx.Error(401, "Token refresh window expired")
		return
	}
	
	// 生成新 Token
	newToken, _ := jwt.Sign(map[string]interface{}{
		"user_id":  claims["user_id"],
		"username": claims["username"],
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	
	ctx.JSON(200, zoox.H{"token": newToken})
})
```

## 自定义 Claims

在 Token 中包含自定义信息：

```go
token, err := jwt.Sign(map[string]interface{}{
	// 标准 Claims
	"sub":   "user123",           // Subject
	"iss":   "zoox-app",          // Issuer
	"aud":   "zoox-client",       // Audience
	"exp":   time.Now().Add(24 * time.Hour).Unix(), // Expiration
	"iat":   time.Now().Unix(),   // Issued At
	"nbf":   time.Now().Unix(),   // Not Before
	
	// 自定义 Claims
	"user_id":   1,
	"username":  "admin",
	"role":      "admin",
	"permissions": []string{"read", "write"},
})
```

## 最佳实践

### 1. 使用强密钥

```go
// 推荐：使用随机生成的强密钥
import "github.com/go-zoox/random"

app.Config.SecretKey = random.String(32)

// 不推荐：使用弱密钥
app.Config.SecretKey = "123456"
```

### 2. 设置合理的过期时间

```go
// 根据应用需求设置
exp := time.Now().Add(24 * time.Hour).Unix()  // 普通应用：24小时
exp := time.Now().Add(1 * time.Hour).Unix()   // 安全敏感：1小时
exp := time.Now().Add(7 * 24 * time.Hour).Unix() // 长期：7天
```

### 3. 包含必要的 Claims

```go
token, _ := jwt.Sign(map[string]interface{}{
	"user_id": userID,
	"exp":     expiration,
	"iat":     issuedAt,
	// 避免包含敏感信息
})
```

### 4. 验证 Token 时检查 Claims

```go
claims, err := jwt.Verify(token)
if err != nil {
	ctx.Error(401, "Invalid token")
	return
}

// 检查过期时间
if exp, ok := claims["exp"].(float64); ok {
	if time.Now().Unix() > int64(exp) {
		ctx.Error(401, "Token expired")
		return
	}
}

// 检查用户角色
if role, ok := claims["role"].(string); ok && role != "admin" {
	ctx.Error(403, "Insufficient permissions")
	return
}
```

### 5. 使用 HTTPS

在生产环境中，始终使用 HTTPS 传输 Token，防止中间人攻击。

## 与 Session 结合

可以同时使用 JWT 和 Session：

```go
app.Post("/login", func(ctx *zoox.Context) {
	// 验证用户名密码
	// ...
	
	// 生成 JWT Token
	token, _ := jwt.Sign(claims)
	
	// 同时设置 Session（用于服务端验证）
	ctx.Session().Set("user_id", userID)
	ctx.Session().Set("token_hash", hashToken(token))
	
	ctx.JSON(200, zoox.H{"token": token})
})
```

## 下一步

- 🛡️ 学习 [认证中间件](../middleware/authentication.md) - JWT、BasicAuth 等中间件
- 🍪 查看 [Session 管理](session.md) - Session 和 Cookie
- 🔒 了解 [安全最佳实践](../best-practices.md) - 安全建议

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
