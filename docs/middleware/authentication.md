# 认证中间件

Zoox 提供了多种认证中间件，支持不同的认证方式。

## JWT 中间件

JWT（JSON Web Token）是最常用的认证方式。

### 基本用法

```go
import "github.com/go-zoox/zoox/middleware"

app := zoox.New()

// 配置密钥
app.Config.SecretKey = "your-secret-key"

// 使用 JWT 中间件
app.Use(middleware.JWT())
```

### 为路由组添加 JWT

```go
api := app.Group("/api")
api.Use(middleware.JWT())

// 所有 /api/* 路由都需要 JWT 认证
api.Get("/users", handler)
```

### Token 获取方式

JWT 中间件支持两种方式获取 Token：

1. **Bearer Token**（推荐）:
   ```
   Authorization: Bearer your-token-here
   ```

2. **查询参数**:
   ```
   GET /api/users?access_token=your-token-here
   ```

**说明**: JWT 中间件实现参考 `middleware/jwt.go`。

## BasicAuth 中间件

HTTP Basic 认证，适用于简单的认证场景。

### 基本用法

```go
import "github.com/go-zoox/zoox/middleware"

app := zoox.New()

app.Use(middleware.BasicAuth("Protected Area", map[string]string{
	"admin": "password",
	"user":  "123456",
}))
```

### 自定义 Realm

```go
app.Use(middleware.BasicAuth("My App", map[string]string{
	"admin": "admin123",
	"user":  "user123",
}))
```

**说明**: BasicAuth 中间件实现参考 `middleware/basic_auth.go`。

## BearerToken 中间件

简单的 Bearer Token 认证，适用于 API 密钥场景。

### 基本用法

```go
import "github.com/go-zoox/zoox/middleware"

app := zoox.New()

app.Use(middleware.BearerToken([]string{
	"token1",
	"token2",
	"token3",
}))
```

### 使用场景

```go
// API 密钥认证
api := app.Group("/api")
api.Use(middleware.BearerToken([]string{
	os.Getenv("API_KEY_1"),
	os.Getenv("API_KEY_2"),
}))
```

**说明**: BearerToken 中间件实现参考 `middleware/bearer_token.go`。

## AuthServer 中间件

通过外部认证服务器进行认证。

### 基本用法

```go
import "github.com/go-zoox/zoox/middleware"

app := zoox.New()

app.Use(middleware.AuthServer(&middleware.AuthServerConfig{
	Server: "https://auth.example.com",
}))
```

### 支持的认证方式

AuthServer 中间件支持：

1. **Bearer Token** - 通过认证服务器验证 Token
2. **Basic Auth** - 通过认证服务器验证用户名密码

**说明**: AuthServer 中间件实现参考 `middleware/auth_server.go`。

## 组合使用

可以组合多个认证方式：

```go
// 公共路由
app.Get("/public", publicHandler)

// API 路由（需要 Bearer Token）
api := app.Group("/api")
api.Use(middleware.BearerToken([]string{apiKey}))

// 管理路由（需要 JWT）
admin := app.Group("/admin")
admin.Use(middleware.JWT())
```

## 自定义认证中间件

### 示例：API Key 认证

```go
func APIKeyAuth(apiKeys []string) zoox.Middleware {
	return func(ctx *zoox.Context) {
		apiKey := ctx.Header().Get("X-API-Key")
		if apiKey == "" {
			ctx.Error(401, "API key required")
			return
		}
		
		valid := false
		for _, key := range apiKeys {
			if apiKey == key {
				valid = true
				break
			}
		}
		
		if !valid {
			ctx.Error(401, "Invalid API key")
			return
		}
		
		ctx.Next()
	}
}

// 使用
app.Use(APIKeyAuth([]string{"key1", "key2"}))
```

### 示例：Session 认证

```go
func SessionAuth() zoox.Middleware {
	return func(ctx *zoox.Context) {
		userID := ctx.Session().Get("user_id")
		if userID == nil {
			ctx.Error(401, "Not authenticated")
			return
		}
		
		ctx.Next()
	}
}

// 使用
app.Use(SessionAuth())
```

## 最佳实践

### 1. 选择合适的认证方式

- **JWT** - 适用于无状态的 API 服务
- **BasicAuth** - 适用于简单的内部服务
- **BearerToken** - 适用于 API 密钥场景
- **Session** - 适用于传统的 Web 应用

### 2. 保护敏感路由

```go
// 公共路由
app.Get("/", publicHandler)
app.Get("/login", loginHandler)

// 需要认证的路由
api := app.Group("/api")
api.Use(middleware.JWT())
api.Get("/users", handler)
```

### 3. 错误处理

```go
func AuthMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		token := ctx.Header().Get("Authorization")
		if token == "" {
			if ctx.AcceptJSON() {
				ctx.JSON(401, zoox.H{
					"error": "Unauthorized",
					"message": "Token required",
				})
			} else {
				ctx.Error(401, "Unauthorized")
			}
			return
		}
		
		ctx.Next()
	}
}
```

### 4. 记录认证失败

```go
func AuthMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		token := ctx.Header().Get("Authorization")
		if token == "" {
			ctx.Logger.Warnf("Authentication failed: no token from %s", ctx.IP())
			ctx.Error(401, "Unauthorized")
			return
		}
		
		ctx.Next()
	}
}
```

## 下一步

- 🔐 学习 [JWT 组件](../components/jwt.md) - JWT 生成和验证
- 🍪 查看 [Session 管理](../components/session.md) - Session 使用
- 🔒 了解 [安全中间件](security.md) - 安全相关中间件

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
