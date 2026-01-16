# 中间件列表

Zoox 提供了丰富的内置中间件，按功能分类如下。

## 认证中间件

### JWT

JWT 身份认证。

```go
app.Use(middleware.JWT())
```

**文件**: `middleware/jwt.go`

### BasicAuth

HTTP Basic 认证。

```go
app.Use(middleware.BasicAuth("Realm", map[string]string{
	"username": "password",
}))
```

**文件**: `middleware/basic_auth.go`

### BearerToken

Bearer Token 认证。

```go
app.Use(middleware.BearerToken([]string{"token1", "token2"}))
```

**文件**: `middleware/bearer_token.go`

### AuthServer

通过认证服务器验证。

```go
app.Use(middleware.AuthServer(&middleware.AuthServerConfig{
	Server: "https://auth.example.com",
}))
```

**文件**: `middleware/auth_server.go`

## 安全中间件

### Helmet

设置安全响应头。

```go
app.Use(middleware.Helmet(nil))
```

**文件**: `middleware/helmet.go`

### CORS

跨域资源共享。

```go
app.Use(middleware.CORS())
```

**文件**: `middleware/cors.go`

### BodyLimit

限制请求体大小。

```go
app.Use(middleware.BodyLimit(func(cfg *middleware.BodyLimitConfig) {
	cfg.MaxSize = 10 * 1024 * 1024
}))
```

**文件**: `middleware/bodylimit.go`

## 性能中间件

### Gzip

响应压缩。

```go
app.Use(middleware.Gzip())
```

**文件**: `middleware/gzip.go`

### CacheControl

缓存控制。

```go
app.Use(middleware.CacheControl(&middleware.CacheControlConfig{
	Paths:  []string{".*"},
	MaxAge: time.Hour,
}))
```

**文件**: `middleware/cache-control.go`

### StaticCache

静态文件缓存。

```go
app.Use(middleware.StaticCache(&middleware.StaticCacheConfig{
	MaxAge: time.Hour,
}))
```

**文件**: `middleware/static_cache.go`

## 监控中间件

### Prometheus

Prometheus 指标收集。

```go
app.Use(middleware.Prometheus())
```

**文件**: `middleware/prometheus.go`

### Sentry

Sentry 错误追踪。

```go
middleware.InitSentry(middleware.InitSentryOption{
	Dsn: "your-sentry-dsn",
})
app.Use(middleware.Sentry())
```

**文件**: `middleware/sentry.go`

### Logger

请求日志。

```go
app.Use(middleware.Logger())
```

**文件**: `middleware/logger.go`

### Runtime

运行时信息。

```go
app.Use(middleware.Runtime())
```

**文件**: `middleware/runtime.go`

## 工具中间件

### Recovery

Panic 恢复。

```go
app.Use(middleware.Recovery())
```

**文件**: `middleware/recovery.go`

### RequestID

请求 ID 生成。

```go
app.Use(middleware.RequestID())
```

**文件**: `middleware/requestid.go`

### RealIP

真实 IP 获取。

```go
app.Use(middleware.RealIP())
```

**文件**: `middleware/realip.go`

### Timeout

请求超时。

```go
app.Use(middleware.Timeout(5 * time.Second))
```

**文件**: `middleware/timeout.go`

### RateLimit

速率限制。

```go
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  100,
}))
```

**文件**: `middleware/ratelimit.go`

### Rewrite

路径重写。

```go
app.Use(middleware.Rewrite(&middleware.RewriteConfig{
	Rewrites: []middleware.Rewrite{
		{From: "/old", To: "/new"},
	},
}))
```

**文件**: `middleware/rewrite.go`

### Proxy

代理中间件。

```go
app.Use(middleware.Proxy(func(ctx *zoox.Context, cfg *middleware.ProxyConfig) (next, stop bool, err error) {
	// 代理逻辑
	return true, false, nil
}))
```

**文件**: `middleware/proxy.go`

### HealthCheck

健康检查。

```go
app.Use(middleware.HealthCheck())
```

**文件**: `middleware/healthcheck.go`

### PProf

性能分析（开发环境）。

```go
app.Use(middleware.PProf())
```

**文件**: `middleware/pprof.go`

### NotFound

404 处理。

```go
app.Use(middleware.NotFound(func(ctx *zoox.Context) {
	ctx.JSON(404, zoox.H{"error": "Not Found"})
}))
```

**文件**: `middleware/notfound.go`

## 使用建议

### 推荐顺序

```go
app.Use(middleware.Recovery())      // 1. 恢复
app.Use(middleware.Logger())        // 2. 日志
app.Use(middleware.RequestID())     // 3. 请求 ID
app.Use(middleware.RealIP())        // 4. 真实 IP
app.Use(middleware.CORS())          // 5. CORS
app.Use(middleware.BodyLimit(...))  // 6. 请求体限制
app.Use(middleware.RateLimit(...))  // 7. 速率限制
app.Use(middleware.JWT())           // 8. 认证
```

## 详细文档

- [认证中间件](../middleware/authentication.md) - JWT、BasicAuth 等
- [安全中间件](../middleware/security.md) - Helmet、CORS 等
- [性能中间件](../middleware/performance.md) - Gzip、CacheControl 等
- [监控中间件](../middleware/monitoring.md) - Prometheus、Sentry 等

## 下一步

- 📝 查看 [Application API](application.md) - 应用方法参考
- 🔌 了解 [Context API](context.md) - Context 方法参考
- 🛣️ 学习 [Router API](router.md) - 路由相关方法

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
