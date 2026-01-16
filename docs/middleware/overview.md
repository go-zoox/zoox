# 中间件概览

中间件是 Zoox 框架的核心功能之一，允许你在请求处理流程中插入自定义逻辑。

## 什么是中间件

中间件是一个函数，它在请求到达处理函数之前或之后执行。中间件可以：

- 记录请求日志
- 验证身份认证
- 处理 CORS
- 压缩响应
- 限制请求速率
- 错误恢复
- 等等...

## 中间件执行流程

```
请求 → 中间件1 → 中间件2 → ... → 处理函数 → 中间件2 → 中间件1 → 响应
```

### Next() 方法

中间件必须调用 `ctx.Next()` 才能继续执行：

```go
func MyMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		// 请求前执行
		fmt.Println("Before handler")
		
		ctx.Next()  // 继续执行
		
		// 响应后执行
		fmt.Println("After handler")
	}
}
```

## 中间件分类

Zoox 提供了丰富的内置中间件，按功能分类：

### 认证中间件

- **JWT** - JWT 身份认证
- **BasicAuth** - HTTP Basic 认证
- **BearerToken** - Bearer Token 认证
- **AuthServer** - 通过认证服务器验证

### 安全中间件

- **Helmet** - 设置安全响应头
- **CORS** - 跨域资源共享
- **BodyLimit** - 限制请求体大小

### 性能中间件

- **Gzip** - 响应压缩
- **CacheControl** - 缓存控制
- **StaticCache** - 静态文件缓存

### 监控中间件

- **Prometheus** - Prometheus 指标收集
- **Sentry** - Sentry 错误追踪
- **Logger** - 请求日志
- **Runtime** - 运行时信息

### 工具中间件

- **Recovery** - Panic 恢复
- **RequestID** - 请求 ID 生成
- **RealIP** - 真实 IP 获取
- **Timeout** - 请求超时
- **RateLimit** - 速率限制

## 注册中间件

### 全局中间件

```go
app := zoox.New()

app.Use(middleware.Logger())
app.Use(middleware.Recovery())
app.Use(middleware.CORS())
```

### 路由组中间件

```go
api := app.Group("/api")
api.Use(middleware.JWT())
api.Use(middleware.RateLimit(...))
```

### 单个路由中间件

```go
app.Get("/protected", 
	middleware.Auth(),
	handler,
)
```

## 中间件执行顺序

中间件按照注册顺序执行：

```go
app.Use(middleware1())  // 1. 最先执行（请求前）
app.Use(middleware2())  // 2. 第二个执行（请求前）
app.Use(middleware3())  // 3. 第三个执行（请求前）

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

## 自定义中间件

### 基本中间件

```go
func MyMiddleware() zoox.Middleware {
	return func(ctx *zoox.Context) {
		start := time.Now()
		
		ctx.Next()
		
		duration := time.Since(start)
		ctx.Logger.Infof("Request took %v", duration)
	}
}
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
```

## 默认中间件

使用 `zoox.Default()` 创建应用时会自动添加默认中间件：

- Recovery
- RequestID
- RealIP
- Logger
- HealthCheck
- Runtime

## 详细文档

- [认证中间件](authentication.md) - JWT、BasicAuth 等
- [安全中间件](security.md) - Helmet、CORS 等
- [性能中间件](performance.md) - Gzip、CacheControl 等
- [监控中间件](monitoring.md) - Prometheus、Sentry 等

## 下一步

- 🔌 学习 [中间件使用指南](../guides/middleware.md) - 详细使用说明
- 🛡️ 查看 [认证中间件](authentication.md) - 认证相关中间件
- 🔒 了解 [安全中间件](security.md) - 安全相关中间件

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
