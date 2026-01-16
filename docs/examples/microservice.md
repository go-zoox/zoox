# 微服务示例

这是一个微服务架构示例，展示了如何使用 Zoox 构建微服务。

## 架构设计

```
Gateway Service (端口 8080)
├── User Service (端口 8081)
├── Product Service (端口 8082)
└── Order Service (端口 8083)
```

## 项目结构

```
microservice/
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
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	
	// 全局中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	app.Use(middleware.RequestID())
	
	// 代理到各个微服务
	app.Proxy("/api/users", "http://localhost:8081", func(cfg *zoox.ProxyConfig) {
		cfg.Rewrites = []zoox.Rewrite{
			{From: "/api/users/(.*)", To: "/$1"},
		}
		cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
			// 添加请求追踪
			ctx.SetHeader("X-Request-ID", ctx.RequestID())
			return nil
		}
	})
	
	app.Proxy("/api/products", "http://localhost:8082", func(cfg *zoox.ProxyConfig) {
		cfg.Rewrites = []zoox.Rewrite{
			{From: "/api/products/(.*)", To: "/$1"},
		}
		cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
			ctx.SetHeader("X-Request-ID", ctx.RequestID())
			return nil
		}
	})
	
	app.Proxy("/api/orders", "http://localhost:8083", func(cfg *zoox.ProxyConfig) {
		cfg.Rewrites = []zoox.Rewrite{
			{From: "/api/orders/(.*)", To: "/$1"},
		}
		cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
			ctx.SetHeader("X-Request-ID", ctx.RequestID())
			return nil
		}
	})
	
	// 健康检查
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"status": "ok"})
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
	ID   int    `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

var users = []*User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
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
	{ID: 1, Name: "Product 1", Price: 99.99},
	{ID: 2, Name: "Product 2", Price: 199.99},
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
	ID        int   `json:"id"`
	UserID    int   `json:"user_id"`
	ProductID int   `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

var orders = []*Order{
	{ID: 1, UserID: 1, ProductID: 1, Quantity: 2},
	{ID: 2, UserID: 2, ProductID: 2, Quantity: 1},
}

func main() {
	app := zoox.New()
	
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	
	// 获取所有订单
	app.Get("/", func(ctx *zoox.Context) {
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
	
	// 创建订单
	app.Post("/", func(ctx *zoox.Context) {
		var order Order
		ctx.BindJSON(&order)
		order.ID = len(orders) + 1
		orders = append(orders, &order)
		ctx.JSON(201, order)
	})
	
	// 健康检查
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"service": "order-service", "status": "ok"})
	})
	
	app.Run(":8083")
}
```

## 运行和测试

### 启动所有服务

```bash
# 终端1: 启动 Gateway
cd gateway
go mod init gateway
go get github.com/go-zoox/zoox
go run main.go

# 终端2: 启动 User Service
cd user-service
go mod init user-service
go get github.com/go-zoox/zoox
go run main.go

# 终端3: 启动 Product Service
cd product-service
go mod init product-service
go get github.com/go-zoox/zoox
go run main.go

# 终端4: 启动 Order Service
cd order-service
go mod init order-service
go get github.com/go-zoox/zoox
go run main.go
```

### 测试

```bash
# 通过 Gateway 访问 User Service
curl http://localhost:8080/api/users

# 通过 Gateway 访问 Product Service
curl http://localhost:8080/api/products

# 通过 Gateway 访问 Order Service
curl http://localhost:8080/api/orders

# 直接访问各个服务
curl http://localhost:8081/
curl http://localhost:8082/
curl http://localhost:8083/
```

## 特性说明

1. **API Gateway** - 统一的入口点
2. **服务发现** - 通过代理实现服务路由
3. **请求追踪** - 使用 RequestID 追踪请求
4. **服务隔离** - 每个服务独立运行
5. **健康检查** - 每个服务提供健康检查端点

## 扩展建议

1. **服务注册** - 使用服务注册中心（如 Consul）
2. **负载均衡** - 在 Gateway 中实现负载均衡
3. **认证授权** - 在 Gateway 中统一处理认证
4. **限流熔断** - 添加限流和熔断机制
5. **监控日志** - 集成 Prometheus 和日志系统

## 下一步

- 📡 查看 [实时应用示例](real-time-app.md) - WebSocket 应用
- 🏗️ 学习 [RESTful API 示例](rest-api.md) - REST API
- 📚 阅读 [最佳实践](../best-practices.md) - 开发建议

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
