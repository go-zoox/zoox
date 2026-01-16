# 第一个应用完整教程

本教程将带你从零开始构建一个完整的 Zoox 应用，包括项目结构、路由、中间件、错误处理等核心功能。

## 项目结构

首先，让我们创建一个合理的项目结构：

```
my-zoox-app/
├── main.go           # 应用入口
├── go.mod            # Go 模块文件
├── go.sum            # 依赖校验文件
└── README.md         # 项目说明
```

## 步骤 1: 创建 Hello World 应用

创建 `main.go` 文件：

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	// 创建应用实例
	app := zoox.New()
	
	// 定义路由和处理函数
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "Hello, Zoox!",
		})
	})
	
	// 启动服务器
	app.Run(":8080")
}
```

运行应用：

```bash
go mod init my-zoox-app
go get github.com/go-zoox/zoox
go run main.go
```

测试：

```bash
curl http://localhost:8080
```

## 步骤 2: 添加中间件

中间件是 Zoox 的核心功能之一。让我们添加一些常用的中间件：

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	
	// 全局中间件（按顺序执行）
	app.Use(middleware.Logger())    // 日志中间件
	app.Use(middleware.Recovery())  // 恢复中间件（捕获 panic）
	app.Use(middleware.CORS())      // CORS 中间件
	
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "Hello, Zoox!",
		})
	})
	
	app.Run(":8080")
}
```

**说明**:
- `middleware.Logger()` - 记录请求日志（参考: `middleware/logger.go`）
- `middleware.Recovery()` - 自动恢复 panic，防止应用崩溃（参考: `middleware/recovery.go`）
- `middleware.CORS()` - 处理跨域请求（参考: `middleware/cors.go`）
- `app.Use()` - 注册全局中间件（参考: `group.go:219-222`）

## 步骤 3: 处理不同类型的请求

### GET 请求（查询参数）

```go
app.Get("/search", func(ctx *zoox.Context) {
	// 获取查询参数
	query := ctx.Query().Get("q")
	page := ctx.Query().Get("page")
	
	ctx.JSON(200, zoox.H{
		"query": query,
		"page":  page,
		"results": []string{"result1", "result2"},
	})
})
```

测试：

```bash
curl "http://localhost:8080/search?q=test&page=1"
```

### POST 请求（JSON Body）

```go
app.Post("/users", func(ctx *zoox.Context) {
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	
	// 绑定 JSON 请求体
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Error(400, "Invalid JSON")
		return
	}
	
	ctx.JSON(201, zoox.H{
		"message": "User created",
		"user":   user,
	})
})
```

测试：

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

### 文件上传

```go
app.Post("/upload", func(ctx *zoox.Context) {
	// 获取上传的文件
	file, fileHeader, err := ctx.File("file")
	if err != nil {
		ctx.Error(400, "No file uploaded")
		return
	}
	defer file.Close()
	
	ctx.JSON(200, zoox.H{
		"message":    "File uploaded",
		"filename":   fileHeader.Filename,
		"size":       fileHeader.Size,
		"mediatype":  fileHeader.Header.Get("Content-Type"),
	})
})
```

## 步骤 4: 响应不同类型的数据

### JSON 响应

```go
app.Get("/api/user", func(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"id":    1,
		"name":  "John Doe",
		"email": "john@example.com",
	})
})
```

### HTML 响应

```go
app.Get("/home", func(ctx *zoox.Context) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Zoox App</title>
	</head>
	<body>
		<h1>Welcome to Zoox!</h1>
	</body>
	</html>
	`
	ctx.HTML(200, html)
})
```

### 字符串响应

```go
app.Get("/text", func(ctx *zoox.Context) {
	ctx.String(200, "Plain text response")
})
```

### 重定向

```go
app.Get("/redirect", func(ctx *zoox.Context) {
	ctx.Redirect("http://example.com", 302)
})
```

## 步骤 5: 错误处理

Zoox 提供了多种错误处理方式：

### 使用 ctx.Error() 处理系统错误

```go
app.Get("/error", func(ctx *zoox.Context) {
	// 系统错误（如 404, 500）
	ctx.Error(404, "Resource not found")
})
```

### 使用 ctx.Fail() 处理业务错误

```go
app.Post("/users", func(ctx *zoox.Context) {
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Fail(err, 4000001, "Invalid request body", 400)
		return
	}
	
	// 业务逻辑验证
	if user.Name == "" {
		ctx.Fail(nil, 4000002, "Name is required", 400)
		return
	}
	
	ctx.Success(zoox.H{
		"message": "User created",
		"user":    user,
	})
})
```

**说明**:
- `ctx.Error()` - 处理系统错误（参考: `context.go:477-497`）
- `ctx.Fail()` - 处理业务错误，返回标准错误格式（参考: `context.go:512-541`）
- `ctx.Success()` - 返回成功响应（参考: `context.go:499-506`）

### 使用 ctx.Success() 返回成功响应

```go
app.Get("/api/data", func(ctx *zoox.Context) {
	data := []zoox.H{
		{"id": 1, "name": "Item 1"},
		{"id": 2, "name": "Item 2"},
	}
	
	ctx.Success(data)
})
```

响应格式：

```json
{
  "code": 200,
  "message": "success",
  "result": [...]
}
```

## 步骤 6: 使用路由组

路由组允许你为多个路由添加共同的前缀和中间件：

```go
// API v1 路由组
apiV1 := app.Group("/api/v1")
apiV1.Use(middleware.RequestID()) // 为 API 路由添加请求 ID

apiV1.Get("/users", func(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"users": []string{"user1", "user2"}})
})

apiV1.Get("/posts", func(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"posts": []string{"post1", "post2"}})
})

// API v2 路由组
apiV2 := app.Group("/api/v2")
apiV2.Get("/users", func(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{"users": []string{"user1", "user2"}})
})
```

## 完整示例代码

以下是包含所有功能的完整 `main.go`：

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
	
	// 首页
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "Welcome to Zoox!",
			"version": zoox.Version,
		})
	})
	
	// 搜索（GET 查询参数）
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
			ctx.Fail(err, 4000001, "Invalid JSON", 400)
			return
		}
		
		if user.Name == "" {
			ctx.Fail(nil, 4000002, "Name is required", 400)
			return
		}
		
		ctx.Success(zoox.H{
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
	
	// HTML 响应
	app.Get("/home", func(ctx *zoox.Context) {
		html := `<!DOCTYPE html>
<html>
<head><title>Zoox App</title></head>
<body><h1>Welcome to Zoox!</h1></body>
</html>`
		ctx.HTML(200, html)
	})
	
	// API 路由组
	api := app.Group("/api/v1")
	api.Use(middleware.RequestID())
	
	api.Get("/data", func(ctx *zoox.Context) {
		ctx.Success([]zoox.H{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
		})
	})
	
	// 启动服务器
	app.Run(":8080")
}
```

## 测试完整应用

```bash
# 启动应用
go run main.go

# 测试各个端点
curl http://localhost:8080/
curl "http://localhost:8080/search?q=test"
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
curl http://localhost:8080/users/123
curl http://localhost:8080/home
curl http://localhost:8080/api/v1/data
```

## 下一步

现在你已经掌握了 Zoox 的基础功能，可以：

1. 🔍 深入了解 [路由系统](../guides/routing.md) - 学习路由的高级特性
2. 🔌 学习 [中间件系统](../guides/middleware.md) - 创建自定义中间件
3. 📝 查看 [Context API](../guides/context.md) - 了解所有可用的方法
4. 💡 浏览 [常见场景示例](examples.md) - 学习实际应用场景

---

**继续学习？** 👉 [路由系统详解](../guides/routing.md) | [中间件使用指南](../guides/middleware.md)
