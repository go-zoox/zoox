# API Gateway 示例

这是一个 API Gateway 示例，展示了如何使用 Zoox 构建一个完整的 API 网关服务。

## 架构设计

```
Client
  │
  ▼
API Gateway (端口 8080)
  ├── Authentication & Authorization
  ├── Rate Limiting
  ├── Request Routing
  └── Response Aggregation
  │
  ├── User Service (端口 8081)
  ├── Product Service (端口 8082)
  └── Order Service (端口 8083)
```

## 项目结构

```
api-gateway/
├── gateway/
│   └── main.go
├── user-service/
│   └── main.go
├── product-service/
│   └── main.go
└── order-service/
    └── main.go
```

## Gateway Service

### gateway/main.go

```go
package main

import (
	"time"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	app.Config.TrustProxy = true // 网关在上游代理后面时建议开启

	// 全局中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	app.Use(middleware.RequestID())

	// 请求限流
	app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
		Period: 1 * time.Minute,
		Limit:  100, // 每分钟 100 个请求
	}))

	// API 版本路由
	api := app.Group("/api/v1")

	// 认证中间件（某些路由需要）
	authRequired := api.Group("")
	authRequired.Use(func(ctx *zoox.Context) {
		token := ctx.Header().Get("Authorization")
		if token == "" {
			ctx.Error(401, "Unauthorized")
			return
		}
		// 验证 token（简化示例）
		if token != "Bearer valid-token" {
			ctx.Error(403, "Forbidden")
			return
		}
		ctx.Next()
	})

	// 代理到 User Service
	api.Proxy("/users", "http://localhost:8081", func(cfg *zoox.ProxyConfig) {
		cfg.Rewrites = []zoox.Rewrite{
			{From: "/api/v1/users/(.*)", To: "/$1"},
		}
		cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
			// 添加请求追踪
			ctx.SetHeader("X-Request-ID", ctx.RequestID())
			ctx.SetHeader("X-Forwarded-For", ctx.IP())
			ctx.SetHeader("X-User-Service", "true")
			return nil
		}
		cfg.OnResponseWithContext = func(ctx *zoox.Context) error {
			// 记录响应时间
			ctx.Logger.Infof("User service response: %d", ctx.StatusCode())
			return nil
		}
	})

	// 代理到 Product Service（需要认证）
	authRequired.Proxy("/products", "http://localhost:8082", func(cfg *zoox.ProxyConfig) {
		cfg.Rewrites = []zoox.Rewrite{
			{From: "/api/v1/products/(.*)", To: "/$1"},
		}
		cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
			ctx.SetHeader("X-Request-ID", ctx.RequestID())
			ctx.SetHeader("X-Forwarded-For", ctx.IP())
			ctx.SetHeader("X-Product-Service", "true")
			return nil
		}
	})

	// 代理到 Order Service（需要认证）
	authRequired.Proxy("/orders", "http://localhost:8083", func(cfg *zoox.ProxyConfig) {
		cfg.Rewrites = []zoox.Rewrite{
			{From: "/api/v1/orders/(.*)", To: "/$1"},
		}
		cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
			ctx.SetHeader("X-Request-ID", ctx.RequestID())
			ctx.SetHeader("X-Forwarded-For", ctx.IP())
			ctx.SetHeader("X-Order-Service", "true")
			return nil
		}
	})

	// 聚合 API - 获取用户订单（聚合多个服务）
	api.Get("/users/:userId/orders", func(ctx *zoox.Context) {
		userID := ctx.Param().Get("userId")

		// 这里应该使用 HTTP 客户端调用各个服务并聚合结果
		// 为简化示例，直接返回模拟数据
		ctx.JSON(200, zoox.H{
			"userId": userID,
			"orders": []zoox.H{
				{"id": 1, "product": "Product A", "quantity": 2},
				{"id": 2, "product": "Product B", "quantity": 1},
			},
		})
	})

	// 网关健康检查
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"status": "ok",
			"service": "api-gateway",
			"version": "1.0.0",
		})
	})

	// 服务状态聚合
	app.Get("/status", func(ctx *zoox.Context) {
		// 检查所有后端服务的健康状态
		ctx.JSON(200, zoox.H{
			"gateway": "ok",
			"services": map[string]string{
				"user-service": "ok",
				"product-service": "ok",
				"order-service": "ok",
			},
		})
	})

	app.Run(":8080")
}
```

## User Service

### user-service/main.go

```go
package main

import (
	"strconv"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []*User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
	{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
}

func main() {
	app := zoox.New()

	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// 获取所有用户
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"users": users})
	})

	// 获取单个用户
	app.Get("/:id", func(ctx *zoox.Context) {
		id, _ := strconv.Atoi(ctx.Param().Get("id"))
		for _, user := range users {
			if user.ID == id {
				ctx.JSON(200, user)
				return
			}
		}
		ctx.Error(404, "User not found")
	})

	// 健康检查
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"service": "user-service", "status": "ok"})
	})

	app.Run(":8081")
}
```

## Product Service

### product-service/main.go

```go
package main

import (
	"strconv"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var products = []*Product{
	{ID: 1, Name: "Product A", Price: 99.99},
	{ID: 2, Name: "Product B", Price: 199.99},
	{ID: 3, Name: "Product C", Price: 299.99},
}

func main() {
	app := zoox.New()

	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// 获取所有产品
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"products": products})
	})

	// 获取单个产品
	app.Get("/:id", func(ctx *zoox.Context) {
		id, _ := strconv.Atoi(ctx.Param().Get("id"))
		for _, product := range products {
			if product.ID == id {
				ctx.JSON(200, product)
				return
			}
		}
		ctx.Error(404, "Product not found")
	})

	// 健康检查
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"service": "product-service", "status": "ok"})
	})

	app.Run(":8082")
}
```

## Order Service

### order-service/main.go

```go
package main

import (
	"strconv"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

type Order struct {
	ID        int    `json:"id"`
	UserID    int    `json:"userId"`
	ProductID int    `json:"productId"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
}

var orders = []*Order{
	{ID: 1, UserID: 1, ProductID: 1, Quantity: 2, Status: "pending"},
	{ID: 2, UserID: 1, ProductID: 2, Quantity: 1, Status: "completed"},
	{ID: 3, UserID: 2, ProductID: 3, Quantity: 1, Status: "pending"},
}

func main() {
	app := zoox.New()

	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// 获取所有订单
	app.Get("/", func(ctx *zoox.Context) {
		userID := ctx.Query().Get("userId")
		if userID != "" {
			uid, _ := strconv.Atoi(userID)
			var userOrders []*Order
			for _, order := range orders {
				if order.UserID == uid {
					userOrders = append(userOrders, order)
				}
			}
			ctx.JSON(200, zoox.H{"orders": userOrders})
			return
		}
		ctx.JSON(200, zoox.H{"orders": orders})
	})

	// 获取单个订单
	app.Get("/:id", func(ctx *zoox.Context) {
		id, _ := strconv.Atoi(ctx.Param().Get("id"))
		for _, order := range orders {
			if order.ID == id {
				ctx.JSON(200, order)
				return
			}
		}
		ctx.Error(404, "Order not found")
	})

	// 健康检查
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"service": "order-service", "status": "ok"})
	})

	app.Run(":8083")
}
```

## 功能特性

### 1. 请求路由

Gateway 根据路径将请求路由到不同的后端服务：

- `/api/v1/users/*` → User Service
- `/api/v1/products/*` → Product Service
- `/api/v1/orders/*` → Order Service

### 2. 认证和授权

使用中间件进行认证验证：

```go
authRequired.Use(func(ctx *zoox.Context) {
	token := ctx.Header().Get("Authorization")
	// 验证 token
})
```

### 3. 请求限流

使用限流中间件保护后端服务：

```go
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: 1 * time.Minute,
	Limit:  100,
}))
```

### 4. 请求追踪

通过请求头传递追踪信息：

```go
ctx.SetHeader("X-Request-ID", ctx.RequestID())
ctx.SetHeader("X-Forwarded-For", ctx.IP())
```

### 5. 响应聚合

聚合多个服务的响应数据：

```go
api.Get("/users/:userId/orders", func(ctx *zoox.Context) {
	// 调用多个服务并聚合结果
})
```

## 测试

### 1. 启动所有服务

```bash
# 终端 1 - Gateway
cd gateway && go run main.go

# 终端 2 - User Service
cd user-service && go run main.go

# 终端 3 - Product Service
cd product-service && go run main.go

# 终端 4 - Order Service
cd order-service && go run main.go
```

### 2. 测试 Gateway

```bash
# 健康检查
curl http://localhost:8080/health

# 获取用户列表（不需要认证）
curl http://localhost:8080/api/v1/users/

# 获取产品列表（需要认证）
curl -H "Authorization: Bearer valid-token" \
  http://localhost:8080/api/v1/products/

# 获取订单列表（需要认证）
curl -H "Authorization: Bearer valid-token" \
  http://localhost:8080/api/v1/orders/

# 获取用户订单（聚合 API）
curl http://localhost:8080/api/v1/users/1/orders
```

### 3. 测试限流

```bash
# 快速发送 101 个请求，第 101 个应该被限流
for i in {1..101}; do
  curl http://localhost:8080/api/v1/users/
done
```

## 高级功能

### 负载均衡

可以扩展为支持多个后端服务实例的负载均衡：

```go
// 使用服务发现或负载均衡器
backendURLs := []string{
	"http://user-service-1:8081",
	"http://user-service-2:8081",
	"http://user-service-3:8081",
}
```

### 熔断器

添加熔断器防止级联故障：

```go
// 实现熔断逻辑
if errorRate > threshold {
	// 暂时停止转发请求
}
```

### API 版本管理

支持多个 API 版本：

```go
v1 := app.Group("/api/v1")
v2 := app.Group("/api/v2")
```

## 下一步

- 📡 查看 [RESTful API 示例](rest-api.md) - REST API 开发
- 🔌 查看 [WebSocket 应用示例](real-time-app.md) - WebSocket 应用
- 🏗️ 学习 [微服务示例](microservice.md) - 微服务架构
- 📚 阅读 [最佳实践](../best-practices.md) - 开发建议
