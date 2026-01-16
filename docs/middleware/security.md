# 安全中间件

Zoox 提供了多种安全中间件，帮助保护你的应用。

## Helmet 中间件

设置安全响应头，防止常见的 Web 漏洞。

### 基本用法

```go
import "github.com/go-zoox/zoox/middleware"

app := zoox.New()

app.Use(middleware.Helmet(nil))
```

### 自定义配置

```go
app.Use(middleware.Helmet(&middleware.HelmetConfig{
	// 自定义配置
}))
```

**说明**: Helmet 中间件实现参考 `middleware/helmet.go`。

## CORS 中间件

处理跨域资源共享（CORS）。

### 基本用法

```go
import "github.com/go-zoox/zoox/middleware"

app := zoox.New()

// 默认配置（允许所有来源）
app.Use(middleware.CORS())
```

### 自定义配置

```go
app.Use(middleware.CORS(&middleware.CorsConfig{
	AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	AllowCredentials: true,
	MaxAge:           3600,
	ExposeHeaders:    []string{"X-Total-Count"},
}))
```

### 动态 Origin 验证

```go
app.Use(middleware.CORS(&middleware.CorsConfig{
	AllowOriginFunc: func(origin string) bool {
		// 自定义验证逻辑
		return strings.HasSuffix(origin, ".example.com")
	},
}))
```

**说明**: CORS 中间件实现参考 `middleware/cors.go`。

## BodyLimit 中间件

限制请求体大小，防止 DoS 攻击。

### 基本用法

```go
import "github.com/go-zoox/zoox/middleware"

app := zoox.New()

app.Use(middleware.BodyLimit(func(cfg *middleware.BodyLimitConfig) {
	cfg.MaxSize = 10 * 1024 * 1024  // 10MB
}))
```

### 不同路由不同限制

```go
// 全局限制
app.Use(middleware.BodyLimit(func(cfg *middleware.BodyLimitConfig) {
	cfg.MaxSize = 1 * 1024 * 1024  // 1MB
}))

// API 路由允许更大的请求体
api := app.Group("/api")
api.Use(middleware.BodyLimit(func(cfg *middleware.BodyLimitConfig) {
	cfg.MaxSize = 10 * 1024 * 1024  // 10MB
}))
```

**说明**: BodyLimit 中间件实现参考 `middleware/bodylimit.go`。

## RateLimit 中间件

限制请求速率，防止暴力攻击和 DoS。

### 基本用法

```go
import (
	"time"
	"github.com/go-zoox/zoox/middleware"
)

app := zoox.New()

app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  100,  // 每分钟最多 100 个请求
}))
```

### 基于 IP 的限流

```go
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  60,  // 每分钟 60 个请求
	KeyFunc: func(ctx *zoox.Context) string {
		return ctx.IP()  // 基于 IP 限流
	},
}))
```

### 不同路由不同限制

```go
// 公共路由：宽松限制
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  1000,
}))

// API 路由：严格限制
api := app.Group("/api")
api.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  100,
}))
```

**说明**: RateLimit 中间件实现参考 `middleware/ratelimit.go`。

## 组合使用

将多个安全中间件组合使用：

```go
app := zoox.New()

// 安全响应头
app.Use(middleware.Helmet(nil))

// CORS
app.Use(middleware.CORS(&middleware.CorsConfig{
	AllowOrigins: []string{"https://example.com"},
}))

// 请求体限制
app.Use(middleware.BodyLimit(func(cfg *middleware.BodyLimitConfig) {
	cfg.MaxSize = 10 * 1024 * 1024
}))

// 速率限制
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  100,
}))
```

## 最佳实践

### 1. 生产环境配置

```go
if app.IsProd() {
	// 严格的安全配置
	app.Use(middleware.Helmet(nil))
	app.Use(middleware.CORS(&middleware.CorsConfig{
		AllowOrigins:     []string{"https://yourdomain.com"},
		AllowCredentials: true,
	}))
	app.Use(middleware.BodyLimit(func(cfg *middleware.BodyLimitConfig) {
		cfg.MaxSize = 5 * 1024 * 1024  // 5MB
	}))
	app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
		Period: time.Minute,
		Limit:  60,
	}))
}
```

### 2. 开发环境配置

```go
if !app.IsProd() {
	// 宽松的开发配置
	app.Use(middleware.CORS())  // 允许所有来源
}
```

### 3. 错误处理

```go
app.Use(middleware.BodyLimit(func(cfg *middleware.BodyLimitConfig) {
	cfg.MaxSize = 10 * 1024 * 1024
	cfg.OnError = func(ctx *zoox.Context, err error) {
		ctx.Logger.Errorf("Body limit exceeded: %v", err)
		ctx.JSON(413, zoox.H{
			"error": "Request entity too large",
		})
	}
}))
```

### 4. 监控和日志

```go
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  100,
	OnLimit: func(ctx *zoox.Context) {
		ctx.Logger.Warnf("Rate limit exceeded for %s", ctx.IP())
	},
}))
```

## 下一步

- 🛡️ 学习 [认证中间件](authentication.md) - JWT、BasicAuth 等
- 📊 查看 [监控中间件](monitoring.md) - Prometheus、Sentry 等
- 🚀 了解 [性能中间件](performance.md) - Gzip、CacheControl 等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
