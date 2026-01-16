# 5分钟快速开始

本指南将帮助你在 5 分钟内快速上手 Zoox 框架。我们将通过四个简单步骤创建一个完整的 Web 应用。

## 第一步：创建应用

首先，创建一个新的 Go 应用并初始化：

```bash
mkdir zoox-quickstart
cd zoox-quickstart
go mod init zoox-quickstart
go get github.com/go-zoox/zoox
```

创建 `main.go` 文件，导入 Zoox 并创建应用实例：

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	// 创建应用实例
	app := zoox.New()
	
	// 你的代码将在这里...
	
	// 启动服务器
	app.Run(":8080")
}
```

**说明**:
- `zoox.New()` 创建一个新的应用实例（参考: `application.go:132-157`）
- `zoox.Default()` 也可以使用，它会自动添加一些默认中间件（如日志、恢复等）
- `app.Run(":8080")` 启动服务器监听 8080 端口（参考: `application.go:297-330`）

## 第二步：定义路由

现在让我们添加一些路由。Zoox 支持所有 HTTP 方法：

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	app := zoox.New()
	
	// GET 路由
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "Hello, Zoox!",
		})
	})
	
	// POST 路由
	app.Post("/users", func(ctx *zoox.Context) {
		ctx.JSON(201, zoox.H{
			"message": "User created",
		})
	})
	
	// 带参数的路由
	app.Get("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		ctx.JSON(200, zoox.H{
			"id":   id,
			"name": "John Doe",
		})
	})
	
	app.Run(":8080")
}
```

**说明**:
- `app.Get()`, `app.Post()` 等方法用于注册路由（参考: `group.go:79-107`）
- `ctx.Param().Get("id")` 获取路由参数（参考: `context.go:224-227`）
- `ctx.JSON()` 返回 JSON 响应（参考: `context.go:402-413`）
- `zoox.H` 是 `map[string]interface{}` 的快捷方式

## 第三步：处理请求

让我们添加更完整的请求处理逻辑：

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	app := zoox.New()
	
	// 获取查询参数
	app.Get("/search", func(ctx *zoox.Context) {
		query := ctx.Query().Get("q")
		ctx.JSON(200, zoox.H{
			"query": query,
			"results": []string{"result1", "result2"},
		})
	})
	
	// 处理 POST JSON 请求
	app.Post("/users", func(ctx *zoox.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		
		// 绑定 JSON 请求体到结构体
		if err := ctx.BindJSON(&user); err != nil {
			ctx.Error(400, "Invalid JSON")
			return
		}
		
		ctx.JSON(201, zoox.H{
			"message": "User created",
			"user":   user,
		})
	})
	
	// 获取路由参数
	app.Get("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		ctx.JSON(200, zoox.H{
			"id":   id,
			"name": "User " + id,
		})
	})
	
	app.Run(":8080")
}
```

**说明**:
- `ctx.Query().Get("q")` 获取查询参数（参考: `context.go:215-222`）
- `ctx.BindJSON(&user)` 将 JSON 请求体绑定到结构体（参考: `context.go:766-801`）
- `ctx.Error()` 返回错误响应（参考: `context.go:477-497`）

## 第四步：启动服务

现在运行你的应用：

```bash
go run main.go
```

你应该看到类似以下的输出：

```
  ____               
 /_  / ___  ___ __ __
  / /_/ _ \/ _ \\ \ /
 /___/\___/\___/_\_\  v1.16.6

Lightweight, high performance Go web framework

https://github.com/go-zoox/zoox
____________________________________O/_______
                                    O\

[router] register:      GET /
[router] register:     POST /users
[router] register:      GET /users/:id
[router] register:      GET /search
Server started at http://127.0.0.1:8080
```

## 测试你的应用

在另一个终端中测试各个端点：

```bash
# 测试首页
curl http://localhost:8080/

# 测试查询参数
curl "http://localhost:8080/search?q=test"

# 测试 POST 请求
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# 测试路由参数
curl http://localhost:8080/users/123
```

## 完整示例代码

以下是完整的 `main.go` 文件：

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	app := zoox.New()
	
	// 首页
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "Hello, Zoox!",
			"version": zoox.Version,
		})
	})
	
	// 搜索（带查询参数）
	app.Get("/search", func(ctx *zoox.Context) {
		query := ctx.Query().Get("q")
		ctx.JSON(200, zoox.H{
			"query":   query,
			"results": []string{"result1", "result2"},
		})
	})
	
	// 创建用户（POST JSON）
	app.Post("/users", func(ctx *zoox.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		
		if err := ctx.BindJSON(&user); err != nil {
			ctx.Error(400, "Invalid JSON")
			return
		}
		
		ctx.JSON(201, zoox.H{
			"message": "User created",
			"user":    user,
		})
	})
	
	// 获取用户（路由参数）
	app.Get("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		ctx.JSON(200, zoox.H{
			"id":   id,
			"name": "User " + id,
		})
	})
	
	// 启动服务器
	app.Run(":8080")
}
```

## 下一步

恭喜！你已经掌握了 Zoox 的基础用法。接下来可以：

1. 📚 学习 [第一个应用完整教程](first-app.md) - 了解如何构建完整的应用
2. 🛣️ 深入了解 [路由系统](guides/routing.md) - 学习路由的高级特性
3. 🔌 学习 [中间件使用](guides/middleware.md) - 添加日志、认证等功能
4. 💡 查看 [常见场景示例](examples.md) - 学习实际应用场景

---

**继续学习？** 👉 [第一个应用教程](first-app.md) | [路由系统详解](../guides/routing.md)
