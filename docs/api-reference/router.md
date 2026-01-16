# Router API 参考

Router 提供了路由注册和管理的功能。

## 路由注册

### Get/Post/Put/Patch/Delete/Head/Options/Connect

注册 HTTP 方法路由。

```go
app.Get("/", handler)
app.Post("/users", handler)
app.Put("/users/:id", handler)
app.Patch("/users/:id", handler)
app.Delete("/users/:id", handler)
app.Head("/", handler)
app.Options("/", handler)
app.Connect("/", handler)
```

### Any(path string, handler ...HandlerFunc)

注册所有 HTTP 方法的路由。

```go
app.Any("/all", handler)
```

## 路由组

### Group(prefix string, cb ...GroupFunc)

创建路由组。

```go
api := app.Group("/api/v1")
api.Get("/users", handler)
```

### 嵌套路由组

```go
api := app.Group("/api")
v1 := api.Group("/v1")
v1.Get("/users", handler)  // 路径: /api/v1/users
```

## 路由参数

### 命名参数 :id

```go
app.Get("/users/:id", handler)
// 匹配: /users/123
// ctx.Param().Get("id") => "123"
```

### 花括号参数 {id}

```go
app.Get("/users/{id}", handler)
// 匹配: /users/123
// ctx.Param().Get("id") => "123"
```

### 通配符 *filepath

```go
app.Get("/static/*filepath", handler)
// 匹配: /static/css/style.css
// ctx.Param().Get("filepath") => "css/style.css"
```

## 静态文件

### Static(basePath, rootDir string, options ...*StaticOptions)

提供静态文件服务。

```go
app.Static("/static", "./public")
```

### StaticFS(relativePath string, fs http.FileSystem)

使用自定义文件系统。

```go
app.StaticFS("/static", http.Dir("./public"))
```

## 代理

### Proxy(path, target string, options ...func(cfg *ProxyConfig))

设置代理路由。

```go
app.Proxy("/api", "http://backend:8080")
```

## JSON-RPC

### JSONRPC(path string, handler JSONRPCHandlerFunc)

注册 JSON-RPC 路由。

```go
app.JSONRPC("/rpc", func(registry zoox.JSONRPCRegistry) {
	registry.Register("method", handler)
})
```

## WebSocket

### WebSocket(path string, opts ...func(opt *WebSocketOption))

注册 WebSocket 路由。

```go
server, _ := app.WebSocket("/ws")
server.OnMessage(func(message []byte) {
	server.WriteText("Echo: " + string(message))
})
```

## 中间件

### Use(middlewares ...HandlerFunc)

为路由组添加中间件。

```go
api := app.Group("/api")
api.Use(middleware.JWT())
api.Use(middleware.RateLimit(...))
```

## 路由匹配

路由匹配遵循以下优先级：

1. 精确匹配（如 `/users`）
2. 命名参数（如 `/users/:id`）
3. 通配符（如 `/static/*filepath`）

## 完整示例

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	
	// 基本路由
	app.Get("/", handler)
	app.Post("/users", handler)
	
	// 路由参数
	app.Get("/users/:id", handler)
	app.Get("/posts/:postId/comments/:commentId", handler)
	
	// 路由组
	api := app.Group("/api/v1")
	api.Use(middleware.JWT())
	api.Get("/users", handler)
	api.Post("/users", handler)
	
	// 静态文件
	app.Static("/static", "./public")
	
	// 代理
	app.Proxy("/backend", "http://backend:8080")
	
	// WebSocket
	app.WebSocket("/ws", func(opt *zoox.WebSocketOption) {
		// WebSocket 配置
	})
	
	app.Run(":8080")
}
```

## 下一步

- 📝 查看 [Application API](application.md) - 应用方法参考
- 🔌 了解 [Context API](context.md) - Context 方法参考
- 🛡️ 学习 [中间件列表](middleware-list.md) - 所有内置中间件

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
