# 路由系统详解

Zoox 使用基于 Trie 树的高性能路由系统，支持路由参数、路由组、静态文件服务等高级功能。

## 路由基础

### 基本路由

Zoox 支持所有 HTTP 方法：

```go
app := zoox.New()

app.Get("/", handler)
app.Post("/users", handler)
app.Put("/users/:id", handler)
app.Patch("/users/:id", handler)
app.Delete("/users/:id", handler)
app.Head("/", handler)
app.Options("/", handler)
app.Connect("/", handler)
```

### 路由参数

Zoox 支持三种路由参数格式：

#### 1. 命名参数 `:id`

```go
app.Get("/users/:id", func(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	ctx.JSON(200, zoox.H{"id": id})
})
```

访问 `/users/123` 时，`id` 的值为 `"123"`。

#### 2. 花括号参数 `{id}`

```go
app.Get("/users/{id}", func(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	ctx.JSON(200, zoox.H{"id": id})
})
```

功能与 `:id` 相同，只是语法不同。

#### 3. 通配符 `*filepath`

```go
app.Get("/static/*filepath", func(ctx *zoox.Context) {
	filepath := ctx.Param().Get("filepath")
	ctx.JSON(200, zoox.H{"filepath": filepath})
})
```

访问 `/static/css/style.css` 时，`filepath` 的值为 `"css/style.css"`。

**说明**: 路由参数解析逻辑参考 `router.go:60-88`。

### 路由匹配优先级

路由匹配遵循以下优先级：

1. 精确匹配（如 `/users`）
2. 命名参数（如 `/users/:id`）
3. 通配符（如 `/static/*filepath`）

## 路由组

路由组允许你为多个路由添加共同的前缀和中间件：

```go
// 创建路由组
api := app.Group("/api/v1")

// 为路由组添加中间件
api.Use(middleware.Logger())
api.Use(middleware.RequestID())

// 路由组内的路由会自动添加前缀
api.Get("/users", handler)      // 实际路径: /api/v1/users
api.Get("/posts", handler)      // 实际路径: /api/v1/posts
```

**说明**: 路由组实现参考 `group.go:39-50`。

### 嵌套路由组

路由组可以嵌套：

```go
api := app.Group("/api")
v1 := api.Group("/v1")
v1.Get("/users", handler)  // 实际路径: /api/v1/users
```

## 静态文件服务

### 基本用法

```go
// 提供静态文件服务
// 访问 http://localhost:8080/static/ 会映射到 ./public/ 目录
app.Static("/static", "./public")
```

### 高级配置

```go
app.Static("/static", "./public", &zoox.StaticOptions{
	Gzip:         true,              // 启用 Gzip 压缩
	CacheControl: "public, max-age=3600", // 缓存控制
	MaxAge:       1 * time.Hour,    // 最大缓存时间
	Index:         true,             // 支持 index.html
})
```

**说明**: 静态文件服务实现参考 `group.go:345-401`。

### 自定义文件系统

```go
import "net/http"

app.StaticFS("/static", http.Dir("./public"))
```

## 代理路由

Zoox 支持将请求代理到后端服务：

```go
// 基本代理
app.Proxy("/api", "http://backend:8080")

// 带路径重写的代理
app.Proxy("/api/v1", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.Rewrites = []zoox.Rewrite{
		{From: "/api/v1/(.*)", To: "/$1"},
	}
})

// 带请求/响应钩子的代理
app.Proxy("/api", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.OnRequestWithContext = func(ctx *zoox.Context) error {
		// 在请求前执行
		ctx.SetHeader("X-Forwarded-For", ctx.IP())
		return nil
	}
	
	cfg.OnResponseWithContext = func(ctx *zoox.Context) error {
		// 在响应后执行
		return nil
	}
})
```

**说明**: 代理功能实现参考 `group.go:145-192`。

## Trie 树路由原理

Zoox 使用 Trie 树（前缀树）实现路由匹配，具有以下优势：

1. **O(1) 查找时间** - 路由匹配时间复杂度为 O(1)
2. **内存高效** - 共享公共前缀，节省内存
3. **支持动态路由** - 轻松支持路由参数和通配符

### Trie 树结构示例

对于以下路由：
- `GET /users`
- `GET /users/:id`
- `GET /posts/:id/comments`

Trie 树结构：

```
GET
├── users
│   ├── (exact match: /users)
│   └── :id (param match: /users/:id)
└── posts
    └── :id
        └── comments (exact match: /posts/:id/comments)
```

**说明**: Trie 树实现参考 `components/router/trie.go` 和 `router.go:25-40`。

## 路由注册

### 路由注册顺序

路由按照注册顺序进行匹配，先注册的路由优先匹配。

```go
app.Get("/users/:id", handler1)  // 先注册
app.Get("/users/new", handler2)  // 后注册，但精确匹配优先
```

访问 `/users/new` 时，会匹配到 `handler2`（精确匹配优先）。

### 重复路由检测

如果注册了重复的路由，应用启动时会 panic：

```go
app.Get("/users", handler1)
app.Get("/users", handler2)  // Panic: route already registered
```

**说明**: 重复路由检测参考 `router.go:50-52`。

## 路由参数获取

### 基本获取

```go
app.Get("/users/:id", func(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	ctx.JSON(200, zoox.H{"id": id})
})
```

### 获取所有参数

```go
app.Get("/users/:id/posts/:postId", func(ctx *zoox.Context) {
	params := ctx.Params()
	
	id := params.Get("id")
	postId := params.Get("postId")
	
	ctx.JSON(200, zoox.H{
		"user_id": id,
		"post_id": postId,
	})
})
```

### 绑定到结构体

```go
type UserParams struct {
	ID string `param:"id"`
}

app.Get("/users/:id", func(ctx *zoox.Context) {
	var params UserParams
	if err := ctx.BindParams(&params); err != nil {
		ctx.Error(400, "Invalid parameters")
		return
	}
	
	ctx.JSON(200, zoox.H{"id": params.ID})
})
```

**说明**: 参数绑定参考 `context.go:839-850`。

## 404 处理

### 默认 404 处理

如果路由未匹配，Zoox 会返回默认的 404 响应。

### 自定义 404 处理

```go
app.NotFound(func(ctx *zoox.Context) {
	ctx.JSON(404, zoox.H{
		"error": "Not Found",
		"path":  ctx.Path,
	})
})
```

或者使用 `Fallback` 方法（别名）：

```go
app.Fallback(func(ctx *zoox.Context) {
	ctx.JSON(404, zoox.H{"error": "Not Found"})
})
```

**说明**: 404 处理参考 `application.go:159-167`。

## 路由最佳实践

### 1. 使用路由组组织代码

```go
// 推荐：使用路由组
api := app.Group("/api/v1")
api.Get("/users", handler)
api.Get("/posts", handler)

// 不推荐：重复前缀
app.Get("/api/v1/users", handler)
app.Get("/api/v1/posts", handler)
```

### 2. 路由参数命名

```go
// 推荐：使用有意义的参数名
app.Get("/users/:userId", handler)
app.Get("/posts/:postId", handler)

// 不推荐：使用通用名称
app.Get("/users/:id", handler)
app.Get("/posts/:id", handler)  // 可能冲突
```

### 3. 静态文件路由

```go
// 推荐：使用明确的静态文件路径
app.Static("/static", "./public")
app.Static("/assets", "./assets")

// 注意：避免与 API 路由冲突
app.Get("/api/users", handler)  // API 路由
app.Static("/api", "./public")  // 可能冲突！
```

### 4. 路由顺序

```go
// 推荐：先注册精确路由，后注册参数路由
app.Get("/users/new", handler1)      // 精确匹配
app.Get("/users/:id", handler2)     // 参数匹配

// 不推荐：参数路由在前
app.Get("/users/:id", handler2)
app.Get("/users/new", handler1)    // 可能被 :id 匹配
```

## 性能优化

### 1. 路由数量

Trie 树路由系统可以高效处理大量路由，但建议：

- 保持路由结构清晰
- 避免过度嵌套
- 使用路由组组织相关路由

### 2. 路由参数

路由参数匹配性能：

- 命名参数 `:id` - O(1)
- 花括号参数 `{id}` - O(1)
- 通配符 `*filepath` - O(n)，n 为路径段数

### 3. 静态文件

对于静态文件服务：

- 使用 CDN 处理静态资源
- 启用缓存控制
- 使用 Gzip 压缩

## 下一步

- 🔌 学习 [中间件系统](middleware.md) - 为路由添加中间件
- 📝 查看 [Context API](context.md) - 了解请求和响应处理
- 🚀 探索 [高级功能](../advanced/websocket.md) - WebSocket、JSON-RPC 等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
