# Session 管理

Zoox 提供了完整的 Session 管理功能，支持安全的会话存储。

## 基本用法

### 配置 Session

```go
app := zoox.New()

// 设置密钥（用于加密 Session）
app.Config.SecretKey = "your-secret-key"

// 配置 Session
app.Config.Session.MaxAge = 24 * time.Hour  // Session 过期时间
app.Config.Session.HttpOnly = true           // 防止 XSS
app.Config.Session.Secure = true             // 仅 HTTPS
app.Config.Session.SameSite = "Lax"          // CSRF 保护
```

**说明**: Session 配置参考 `config/config.go:38`。

### 使用 Session

```go
app.Get("/login", func(ctx *zoox.Context) {
	// 设置 Session
	ctx.Session().Set("user_id", 123)
	ctx.Session().Set("username", "alice")
	
	ctx.JSON(200, zoox.H{"message": "Logged in"})
})

app.Get("/profile", func(ctx *zoox.Context) {
	// 获取 Session
	userID := ctx.Session().Get("user_id")
	username := ctx.Session().Get("username")
	
	if userID == nil {
		ctx.Error(401, "Not authenticated")
		return
	}
	
	ctx.JSON(200, zoox.H{
		"user_id":  userID,
		"username": username,
	})
})

app.Get("/logout", func(ctx *zoox.Context) {
	// 清除 Session
	ctx.Session().Clear()
	ctx.JSON(200, zoox.H{"message": "Logged out"})
})
```

**说明**: Session 方法参考 `context.go:1021-1033`。

## Session 操作

### 设置值

```go
// 设置单个值
ctx.Session().Set("key", "value")

// 设置多个值
ctx.Session().Set("user_id", 123)
ctx.Session().Set("username", "alice")
ctx.Session().Set("role", "admin")
```

### 获取值

```go
// 获取值
userID := ctx.Session().Get("user_id")

// 获取值（带类型断言）
if userID, ok := ctx.Session().Get("user_id").(int); ok {
	// 使用 userID
}
```

### 删除值

```go
// 删除单个值
ctx.Session().Delete("user_id")

// 清除所有 Session
ctx.Session().Clear()
```

### 检查值是否存在

```go
if ctx.Session().Get("user_id") != nil {
	// Session 存在
}
```

## Session 存储

### Cookie 存储（默认）

Session 默认存储在 Cookie 中，经过加密处理。

### Redis 存储

如果配置了 Redis，Session 可以存储在 Redis 中：

```go
app := zoox.New()

// 配置 Redis
app.Config.Redis.Host = "localhost"
app.Config.Redis.Port = 6379

// Session 会自动使用 Redis（如果可用）
```

## 完整示例

### 登录系统

```go
package main

import (
	"github.com/go-zoox/zoox"
	"time"
)

func main() {
	app := zoox.New()
	
	// 配置 Session
	app.Config.SecretKey = "your-secret-key"
	app.Config.Session.MaxAge = 24 * time.Hour
	
	// 登录
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
			// 设置 Session
			ctx.Session().Set("user_id", 1)
			ctx.Session().Set("username", "admin")
			ctx.Session().Set("role", "admin")
			
			ctx.JSON(200, zoox.H{"message": "Logged in"})
		} else {
			ctx.Error(401, "Invalid credentials")
		}
	})
	
	// 获取当前用户
	app.Get("/me", func(ctx *zoox.Context) {
		userID := ctx.Session().Get("user_id")
		if userID == nil {
			ctx.Error(401, "Not authenticated")
			return
		}
		
		ctx.JSON(200, zoox.H{
			"user_id":  userID,
			"username": ctx.Session().Get("username"),
			"role":     ctx.Session().Get("role"),
		})
	})
	
	// 登出
	app.Get("/logout", func(ctx *zoox.Context) {
		ctx.Session().Clear()
		ctx.JSON(200, zoox.H{"message": "Logged out"})
	})
	
	app.Run(":8080")
}
```

## Session 中间件

创建 Session 验证中间件：

```go
func RequireAuth() zoox.Middleware {
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
app.Get("/protected", RequireAuth(), func(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"message": "Protected resource"})
})
```

## 最佳实践

### 1. 使用强密钥

```go
// 推荐：使用随机生成的强密钥
app.Config.SecretKey = random.String(32)

// 不推荐：使用弱密钥
app.Config.SecretKey = "123456"
```

### 2. 设置合理的过期时间

```go
// 根据应用需求设置
app.Config.Session.MaxAge = 24 * time.Hour  // 普通应用
app.Config.Session.MaxAge = time.Hour       // 安全敏感应用
```

### 3. 启用安全选项

```go
app.Config.Session.HttpOnly = true   // 防止 XSS
app.Config.Session.Secure = true     // 仅 HTTPS（生产环境）
app.Config.Session.SameSite = "Lax"  // CSRF 保护
```

### 4. 定期清理过期 Session

如果使用 Redis 存储，可以设置 TTL 自动清理。

## 下一步

- 🍪 学习 [Cookie 操作](../guides/context.md#cookie-和-session) - Cookie 使用
- 🔐 查看 [JWT 认证](jwt.md) - JWT 生成和验证
- 🛡️ 了解 [认证中间件](../middleware/authentication.md) - 认证相关中间件

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
