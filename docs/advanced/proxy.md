# 代理功能

Zoox 提供了强大的代理功能，支持反向代理、路径重写和请求/响应钩子。

## 基本用法

### 简单代理

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	app := zoox.New()
	
	// 将 /api 路径的请求代理到后端服务
	app.Proxy("/api", "http://backend:8080")
	
	app.Run(":8080")
}
```

**说明**: 代理实现参考 `group.go:145-192`。

## Trust Proxy（网关/中间代理）

当 Zoox 运行在上游网关（Ingress/Nginx/LB）后面，并且你希望保留上游
传入的 `X-Forwarded-*`（例如 `X-Forwarded-Proto: https`）时，可开启：

```go
app := zoox.New()
app.Config.TrustProxy = true
```

开启后，底层代理会优先使用上游 `X-Forwarded-*`，再回退到 TLS/默认值。

### 路由级覆盖（可选）

```go
app.Proxy("/api", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.TrustProxy = true
})
```

## 路径重写

### 基本重写

```go
app.Proxy("/api/v1", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.Rewrites = []zoox.Rewrite{
		{From: "/api/v1/(.*)", To: "/$1"},
	}
})
```

访问 `/api/v1/users` 会被重写为 `/users` 并代理到后端。

### 多个重写规则

```go
app.Proxy("/api", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.Rewrites = []zoox.Rewrite{
		{From: "/api/v1/(.*)", To: "/v1/$1"},
		{From: "/api/v2/(.*)", To: "/v2/$1"},
	}
})
```

## 请求/响应钩子

### OnRequestWithContext

在请求发送到后端前执行：

```go
app.Proxy("/api", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
		// 添加请求头
		ctx.SetHeader("X-Forwarded-For", ctx.IP())
		ctx.SetHeader("X-Request-ID", ctx.RequestID())
		
		// 修改请求路径
		// ctx.Request.URL.Path = "/new-path"
		
		return nil
	}
})
```

### OnResponseWithContext

在响应返回给客户端前执行：

```go
app.Proxy("/api", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.OnResponseWithContext = func(ctx *zoox.Context) error {
		// 修改响应头
		ctx.SetHeader("X-Proxy-By", "zoox")
		
		// 记录响应
		ctx.Logger.Infof("Proxied response: %d", ctx.StatusCode())
		
		return nil
	}
})
```

## 完整示例

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	
	// 全局中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	
	// 代理配置
	app.Proxy("/api", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
		// 路径重写
		cfg.Rewrites = []zoox.Rewrite{
			{From: "/api/(.*)", To: "/$1"},
		}
		
		// 请求钩子
		cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
			// 添加认证头
			token := ctx.Header().Get("Authorization")
			if token != "" {
				ctx.SetHeader("X-Backend-Auth", token)
			}
			
			// 记录请求
			ctx.Logger.Infof("Proxying request: %s %s", ctx.Method, ctx.Path)
			
			return nil
		}
		
		// 响应钩子
		cfg.OnResponseWithContext = func(ctx *zoox.Context) error {
			// 添加响应头
			ctx.SetHeader("X-Proxy-By", "zoox")
			
			// 记录响应
			ctx.Logger.Infof("Proxied response: %d", ctx.StatusCode())
			
			return nil
		}
	})
	
	app.Run(":8080")
}
```

## 错误处理

代理过程中的错误处理：

```go
app.Proxy("/api", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
		// 如果返回错误，代理请求会被取消
		if someCondition {
			return errors.New("request rejected")
		}
		return nil
	}
	
	cfg.OnResponseWithContext = func(ctx *zoox.Context) error {
		// 如果返回错误，响应会被标记为错误
		if ctx.StatusCode() >= 500 {
			ctx.Logger.Errorf("Backend error: %d", ctx.StatusCode())
		}
		return nil
	}
})
```

## 负载均衡

虽然 Zoox 本身不提供负载均衡，但可以通过中间件实现：

```go
type LoadBalancer struct {
	backends []string
	current  int
	mutex    sync.Mutex
}

func (lb *LoadBalancer) Next() string {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	
	backend := lb.backends[lb.current]
	lb.current = (lb.current + 1) % len(lb.backends)
	return backend
}

func main() {
	app := zoox.New()
	
	lb := &LoadBalancer{
		backends: []string{
			"http://backend1:8080",
			"http://backend2:8080",
			"http://backend3:8080",
		},
	}
	
	app.Proxy("/api", "", func(cfg *zoox.ProxyConfig) {
		cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
			// 选择后端服务器
			backend := lb.Next()
			ctx.Request.URL.Host = backend
			return nil
		}
	})
	
	app.Run(":8080")
}
```

## 使用中间件重写

也可以使用 Rewrite 中间件：

```go
import "github.com/go-zoox/zoox/middleware"

app.Use(middleware.Rewrite(&middleware.RewriteConfig{
	Rewrites: []middleware.Rewrite{
		{From: "/api/v1/(.*)", To: "/$1"},
	},
}))
```

**说明**: Rewrite 中间件参考 `middleware/rewrite.go`。

## 最佳实践

### 1. 添加请求追踪

```go
cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
	ctx.SetHeader("X-Request-ID", ctx.RequestID())
	ctx.SetHeader("X-Forwarded-For", ctx.IP())
	return nil
}
```

### 2. 错误处理

```go
cfg.OnResponseWithContext = func(ctx *zoox.Context) error {
	if ctx.StatusCode() >= 500 {
		// 记录错误
		ctx.Logger.Errorf("Backend error: %d", ctx.StatusCode())
	}
	return nil
}
```

### 3. 认证传递

```go
cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
	// 传递认证信息
	token := ctx.Header().Get("Authorization")
	if token != "" {
		ctx.SetHeader("X-Backend-Auth", token)
	}
	return nil
}
```

## 下一步

- ⏰ 学习 [定时任务](cron-jobs.md) - Cron 任务调度
- 📦 查看 [任务队列](job-queue.md) - 后台任务处理
- 🚀 探索 [其他高级功能](jsonrpc.md) - JSON-RPC 等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
