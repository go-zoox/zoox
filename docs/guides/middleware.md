# 中间件使用指南

中间件是 Zoox 框架的核心功能之一，允许你在请求处理流程中插入自定义逻辑。

## 什么是中间件

中间件是一个函数，它在请求到达处理函数之前或之后执行。中间件可以：

- 记录请求日志
- 验证身份认证
- 处理 CORS
- 压缩响应
- 限制请求速率
- 等等...

## 中间件执行流程

中间件的执行遵循以下流程：

```
请求 → 中间件1 → 中间件2 → ... → 处理函数 → 中间件2 → 中间件1 → 响应
```

**说明**: 中间件执行机制参考 `context.go:200-213` 的 `Next()` 方法。

### Next() 方法

中间件必须调用 `ctx.Next()` 才能继续执行下一个中间件或处理函数：

```go
func MyMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		// 请求前执行
		fmt.Println("Before handler")
		
		ctx.Next()  // 继续执行下一个中间件或处理函数
		
		// 响应后执行
		fmt.Println("After handler")
	}
}
```

## 注册中间件

### 全局中间件

全局中间件会应用到所有路由：

```go
app := zoox.New()

// 注册全局中间件
app.Use(middleware.Logger())
app.Use(middleware.Recovery())
app.Use(middleware.CORS())
```

**说明**: 全局中间件注册参考 `application.go:144-147`。

### 路由组中间件

为特定路由组添加中间件：

```go
// 创建路由组
api := app.Group("/api")

// 为路由组添加中间件
api.Use(middleware.JWT())
api.Use(middleware.RateLimit(...))

// 路由组内的所有路由都会应用这些中间件
api.Get("/users", handler)
api.Get("/posts", handler)
```

**说明**: 路由组中间件参考 `group.go:219-222`。

### 单个路由中间件

为单个路由添加中间件：

```go
app.Get("/protected", 
	middleware.Auth(),
	func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"message": "Protected"})
	},
)
```

## 内置中间件

Zoox 提供了丰富的内置中间件：

### 日志中间件

记录请求和响应日志：

```go
app.Use(middleware.Logger())
```

**说明**: 实现参考 `middleware/logger.go`。

### 恢复中间件

自动恢复 panic，防止应用崩溃：

```go
app.Use(middleware.Recovery())
```

**说明**: 实现参考 `middleware/recovery.go`。

### CORS 中间件

处理跨域请求：

```go
// 默认配置（允许所有来源）
app.Use(middleware.CORS())

// 自定义配置
app.Use(middleware.CORS(&middleware.CorsConfig{
	AllowOrigins:     []string{"https://example.com"},
	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	AllowCredentials: true,
	MaxAge:           3600,
}))
```

**说明**: 实现参考 `middleware/cors.go`。

### JWT 中间件

JWT 身份认证：

```go
app.Use(middleware.JWT())
```

**说明**: 实现参考 `middleware/jwt.go`。

### BasicAuth 中间件

HTTP Basic 认证：

```go
app.Use(middleware.BasicAuth("Protected Area", map[string]string{
	"admin": "password",
	"user":  "123456",
}))
```

**说明**: 实现参考 `middleware/basic_auth.go`。

### BearerToken 中间件

Bearer Token 认证：

```go
app.Use(middleware.BearerToken([]string{"token1", "token2"}))
```

**说明**: 实现参考 `middleware/bearer_token.go`。

### Gzip 中间件

压缩响应：

```go
app.Use(middleware.Gzip())
```

**说明**: 实现参考 `middleware/gzip.go`。

### RateLimit 中间件

限制请求速率：

```go
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  100,  // 每分钟最多 100 个请求
}))
```

**说明**: 实现参考 `middleware/ratelimit.go`。

### BodyLimit 中间件

限制请求体大小：

```go
app.Use(middleware.BodyLimit(func(cfg *middleware.BodyLimitConfig) {
	cfg.MaxSize = 10 * 1024 * 1024  // 10MB
}))
```

**说明**: 实现参考 `middleware/bodylimit.go`。

### Helmet 中间件

设置安全响应头：

```go
app.Use(middleware.Helmet(nil))
```

**说明**: 实现参考 `middleware/helmet.go`。

### RequestID 中间件

为每个请求生成唯一 ID：

```go
app.Use(middleware.RequestID())
```

**说明**: 实现参考 `middleware/requestid.go`。

### RealIP 中间件

获取真实客户端 IP：

```go
app.Use(middleware.RealIP())
```

**说明**: 实现参考 `middleware/realip.go`。

### Prometheus 中间件

Prometheus 指标收集：

```go
app.Use(middleware.Prometheus())
```

**说明**: 实现参考 `middleware/prometheus.go`。

### Sentry 中间件

Sentry 错误追踪：

```go
// 需要先初始化 Sentry
middleware.InitSentry(middleware.InitSentryOption{
	Dsn:   "your-sentry-dsn",
	Debug: false,
})

app.Use(middleware.Sentry())
```

**说明**: 实现参考 `middleware/sentry.go`。

## 自定义中间件

### 基本中间件

创建一个简单的中间件：

```go
func MyMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		// 请求前执行
		start := time.Now()
		
		ctx.Next()
		
		// 响应后执行
		duration := time.Since(start)
		ctx.Logger.Infof("Request took %v", duration)
	}
}

// 使用
app.Use(MyMiddleware())
```

### 带配置的中间件

```go
type MyMiddlewareConfig struct {
	Enabled bool
	Timeout time.Duration
}

func MyMiddleware(cfg *MyMiddlewareConfig) zoox.Middleware {
	return func(ctx *zoox.Context) {
		if !cfg.Enabled {
			ctx.Next()
			return
		}
		
		// 中间件逻辑
		ctx.Next()
	}
}

// 使用
app.Use(MyMiddleware(&MyMiddlewareConfig{
	Enabled: true,
	Timeout: 5 * time.Second,
}))
```

### 认证中间件示例

```go
func AuthMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		token := ctx.Header().Get("Authorization")
		if token == "" {
			ctx.JSON(401, zoox.H{
				"error": "Unauthorized",
			})
			return  // 不调用 ctx.Next()，停止执行
		}
		
		// 验证 token
		// ...
		
		ctx.Next()
	}
}
```

### 条件中间件

根据条件决定是否执行：

```go
func ConditionalMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		// 只对特定路径执行
		if strings.HasPrefix(ctx.Path, "/api") {
			// 中间件逻辑
		}
		
		ctx.Next()
	}
}
```

## 中间件执行顺序

中间件的执行顺序很重要：

```go
app.Use(middleware1())  // 最先执行（请求前）
app.Use(middleware2())  // 第二个执行（请求前）
app.Use(middleware3())  // 第三个执行（请求前）

// 处理函数执行

// middleware3 的响应后逻辑
// middleware2 的响应后逻辑
// middleware1 的响应后逻辑（最后执行）
```

### 推荐顺序

```go
app.Use(middleware.Recovery())    // 1. 恢复（最外层）
app.Use(middleware.Logger())      // 2. 日志
app.Use(middleware.RequestID())   // 3. 请求 ID
app.Use(middleware.RealIP())      // 4. 真实 IP
app.Use(middleware.CORS())        // 5. CORS
app.Use(middleware.BodyLimit(...)) // 6. 请求体限制
app.Use(middleware.RateLimit(...)) // 7. 速率限制
app.Use(middleware.JWT())         // 8. 认证（如果需要）
```

## 中间件最佳实践

### 1. 总是调用 ctx.Next()

```go
// 正确
func MyMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		ctx.Next()  // 必须调用
	}
}

// 错误：不调用 ctx.Next() 会阻止后续中间件和处理函数执行
func BadMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		// 忘记调用 ctx.Next()
	}
}
```

### 2. 提前返回时不要调用 ctx.Next()

```go
func AuthMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		if !isAuthenticated(ctx) {
			ctx.JSON(401, zoox.H{"error": "Unauthorized"})
			return  // 提前返回，不调用 ctx.Next()
		}
		
		ctx.Next()  // 认证通过，继续执行
	}
}
```

### 3. 使用配置结构体

```go
// 推荐：使用配置结构体
func MyMiddleware(cfg *MyMiddlewareConfig) zoox.Middleware {
	// ...
}

// 不推荐：使用多个参数
func MyMiddleware(enabled bool, timeout time.Duration) zoox.Middleware {
	// ...
}
```

### 4. 错误处理

```go
func MyMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		// 使用 defer 确保错误被处理
		defer func() {
			if err := recover(); err != nil {
				ctx.Logger.Errorf("Middleware error: %v", err)
				ctx.Error(500, "Internal Server Error")
			}
		}()
		
		ctx.Next()
	}
}
```

## 默认中间件

使用 `zoox.Default()` 创建应用时会自动添加默认中间件：

```go
app := zoox.Default()  // 自动添加默认中间件
```

默认中间件包括（参考 `defaults/defaults.go`）：
- Recovery
- RequestID
- RealIP
- Logger
- HealthCheck
- Runtime

## 下一步

- 📝 查看 [Context API](context.md) - 了解如何在中间件中使用 Context
- 🛡️ 学习 [认证中间件](../middleware/authentication.md) - JWT、BasicAuth 等
- 🔒 了解 [安全中间件](../middleware/security.md) - Helmet、CORS 等
- 📊 探索 [监控中间件](../middleware/monitoring.md) - Prometheus、Sentry 等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
