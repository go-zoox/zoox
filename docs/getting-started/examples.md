# 常见场景快速示例

本文档提供了 Zoox 在实际开发中常见场景的快速示例，帮助你快速应用到项目中。

## 场景 1: RESTful API

创建一个完整的 RESTful API，包含 CRUD 操作：

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

// 模拟数据存储
var users = []zoox.H{
	{"id": "1", "name": "Alice", "email": "alice@example.com"},
	{"id": "2", "name": "Bob", "email": "bob@example.com"},
}

func main() {
	app := zoox.New()
	
	// 中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	
	// API 路由组
	api := app.Group("/api/v1")
	
	// GET /api/v1/users - 获取所有用户
	api.Get("/users", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"users": users,
		})
	})
	
	// GET /api/v1/users/:id - 获取单个用户
	api.Get("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		for _, user := range users {
			if user["id"] == id {
				ctx.JSON(200, user)
				return
			}
		}
		ctx.Error(404, "User not found")
	})
	
	// POST /api/v1/users - 创建用户
	api.Post("/users", func(ctx *zoox.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		
		if err := ctx.BindJSON(&user); err != nil {
			ctx.Fail(err, 4000001, "Invalid JSON", 400)
			return
		}
		
		newUser := zoox.H{
			"id":    len(users) + 1,
			"name":  user.Name,
			"email": user.Email,
		}
		users = append(users, newUser)
		
		ctx.JSON(201, newUser)
	})
	
	// PUT /api/v1/users/:id - 更新用户
	api.Put("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		
		if err := ctx.BindJSON(&user); err != nil {
			ctx.Fail(err, 4000001, "Invalid JSON", 400)
			return
		}
		
		for i, u := range users {
			if u["id"] == id {
				users[i] = zoox.H{
					"id":    id,
					"name":  user.Name,
					"email": user.Email,
				}
				ctx.JSON(200, users[i])
				return
			}
		}
		
		ctx.Error(404, "User not found")
	})
	
	// DELETE /api/v1/users/:id - 删除用户
	api.Delete("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		for i, user := range users {
			if user["id"] == id {
				users = append(users[:i], users[i+1:]...)
				ctx.JSON(200, zoox.H{"message": "User deleted"})
				return
			}
		}
		ctx.Error(404, "User not found")
	})
	
	app.Run(":8080")
}
```

**测试**:

```bash
# 获取所有用户
curl http://localhost:8080/api/v1/users

# 获取单个用户
curl http://localhost:8080/api/v1/users/1

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'

# 更新用户
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.updated@example.com"}'

# 删除用户
curl -X DELETE http://localhost:8080/api/v1/users/1
```

## 场景 2: 静态文件服务

提供静态文件服务（如前端应用）：

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	app := zoox.New()
	
	// 提供静态文件服务
	// 访问 http://localhost:8080/static/ 会映射到 ./public/ 目录
	app.Static("/static", "./public")
	
	// 或者使用 StaticFS 提供自定义文件系统
	// app.StaticFS("/static", http.Dir("./public"))
	
	app.Run(":8080")
}
```

**说明**:
- `app.Static("/static", "./public")` - 将 `/static` 路径映射到 `./public` 目录（参考: `group.go:345-401`）
- 支持自动 MIME 类型识别
- 支持缓存控制

## 场景 3: 表单处理

处理 HTML 表单提交和文件上传：

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	app.Use(middleware.Logger())
	
	// 处理表单提交
	app.Post("/submit", func(ctx *zoox.Context) {
		// 获取表单字段
		name := ctx.Form().Get("name")
		email := ctx.Form().Get("email")
		
		ctx.JSON(200, zoox.H{
			"name":  name,
			"email": email,
		})
	})
	
	// 处理文件上传
	app.Post("/upload", func(ctx *zoox.Context) {
		file, fileHeader, err := ctx.File("file")
		if err != nil {
			ctx.Error(400, "No file uploaded")
			return
		}
		defer file.Close()
		
		ctx.JSON(200, zoox.H{
			"filename": fileHeader.Filename,
			"size":     fileHeader.Size,
		})
	})
	
	// 使用 BindForm 绑定表单到结构体
	app.Post("/register", func(ctx *zoox.Context) {
		var user struct {
			Name  string `form:"name"`
			Email string `form:"email"`
		}
		
		if err := ctx.BindForm(&user); err != nil {
			ctx.Error(400, "Invalid form data")
			return
		}
		
		ctx.JSON(200, user)
	})
	
	app.Run(":8080")
}
```

**测试**:

```bash
# 表单提交
curl -X POST http://localhost:8080/submit \
  -d "name=Alice&email=alice@example.com"

# 文件上传
curl -X POST http://localhost:8080/upload \
  -F "file=@/path/to/file.txt"
```

## 场景 4: Session 和 Cookie

使用 Session 和 Cookie 管理用户状态：

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	app.Use(middleware.Logger())
	
	// 设置 Cookie
	app.Get("/set-cookie", func(ctx *zoox.Context) {
		ctx.Cookie().Set("username", "alice", 3600) // 1小时过期
		ctx.JSON(200, zoox.H{"message": "Cookie set"})
	})
	
	// 读取 Cookie
	app.Get("/get-cookie", func(ctx *zoox.Context) {
		username := ctx.Cookie().Get("username")
		ctx.JSON(200, zoox.H{"username": username})
	})
	
	// 使用 Session
	app.Get("/login", func(ctx *zoox.Context) {
		// 设置 Session
		ctx.Session().Set("user_id", 123)
		ctx.Session().Set("username", "alice")
		
		ctx.JSON(200, zoox.H{"message": "Logged in"})
	})
	
	app.Get("/profile", func(ctx *zoox.Context) {
		// 读取 Session
		userID := ctx.Session().Get("user_id")
		username := ctx.Session().Get("username")
		
		if userID == nil {
			ctx.Error(401, "Not authenticated")
			return
		}
		
		ctx.JSON(200, zoox.H{
			"user_id":  userID,
			"username": username,
		})
	})
	
	app.Get("/logout", func(ctx *zoox.Context) {
		// 清除 Session
		ctx.Session().Clear()
		ctx.JSON(200, zoox.H{"message": "Logged out"})
	})
	
	app.Run(":8080")
}
```

**说明**:
- `ctx.Cookie()` - Cookie 操作（参考: `context.go:1009-1019`）
- `ctx.Session()` - Session 操作（参考: `context.go:1021-1033`）
- Session 需要配置 `SecretKey`（参考: `application.go:175-177`）

## 场景 5: JWT 认证

使用 JWT 进行身份认证：

```go
package main

import (
	"time"
	
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	app.Use(middleware.Logger())
	
	// 登录接口 - 生成 JWT
	app.Post("/login", func(ctx *zoox.Context) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		
		if err := ctx.BindJSON(&creds); err != nil {
			ctx.Error(400, "Invalid JSON")
			return
		}
		
		// 验证用户名密码（示例）
		if creds.Username == "admin" && creds.Password == "password" {
			// 生成 JWT Token
			jwt := ctx.Jwt()
			token, err := jwt.Sign(map[string]interface{}{
				"user_id":  1,
				"username": "admin",
				"exp":      time.Now().Add(24 * time.Hour).Unix(),
			})
			
			if err != nil {
				ctx.Error(500, "Failed to generate token")
				return
			}
			
			ctx.JSON(200, zoox.H{
				"token": token,
			})
		} else {
			ctx.Error(401, "Invalid credentials")
		}
	})
	
	// 受保护的路由组
	protected := app.Group("/api")
	protected.Use(middleware.JWT()) // JWT 中间件
	
	protected.Get("/profile", func(ctx *zoox.Context) {
		// 从 JWT 中获取用户信息
		jwt := ctx.Jwt()
		token, _ := ctx.BearerToken()
		
		claims, err := jwt.Verify(token)
		if err != nil {
			ctx.Error(401, "Invalid token")
			return
		}
		
		ctx.JSON(200, zoox.H{
			"user_id":  claims["user_id"],
			"username": claims["username"],
		})
	})
	
	app.Run(":8080")
}
```

**测试**:

```bash
# 登录获取 Token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# 使用 Token 访问受保护的路由
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**说明**:
- `ctx.Jwt()` - JWT 操作（参考: `context.go:1035-1047`）
- `middleware.JWT()` - JWT 认证中间件（参考: `middleware/jwt.go`）
- 需要配置 `SecretKey`（参考: `application.go:175-177`）

## 场景 6: 查询参数和分页

处理查询参数和实现分页：

```go
package main

import (
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()
	
	// 使用查询参数
	app.Get("/search", func(ctx *zoox.Context) {
		query := ctx.Query().Get("q")
		page := ctx.Query().GetInt("page", 1)      // 默认值 1
		size := ctx.Query().GetInt("size", 10)     // 默认值 10
		
		ctx.JSON(200, zoox.H{
			"query": query,
			"page":  page,
			"size":   size,
		})
	})
	
	// 使用 BindQuery 绑定查询参数到结构体
	app.Get("/users", func(ctx *zoox.Context) {
		var params struct {
			Page  int    `query:"page"`
			Size  int    `query:"size"`
			Order string `query:"order"`
		}
		
		if err := ctx.BindQuery(&params); err != nil {
			ctx.Error(400, "Invalid query parameters")
			return
		}
		
		ctx.JSON(200, params)
	})
	
	app.Run(":8080")
}
```

**测试**:

```bash
curl "http://localhost:8080/search?q=test&page=2&size=20"
curl "http://localhost:8080/users?page=1&size=10&order=desc"
```

## 场景 7: 错误处理和统一响应格式

实现统一的错误处理和响应格式：

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	
	// 使用 Success 返回成功响应
	app.Get("/api/data", func(ctx *zoox.Context) {
		data := []zoox.H{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
		}
		ctx.Success(data)
	})
	
	// 使用 Fail 返回业务错误
	app.Get("/api/user/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		if id == "" {
			ctx.Fail(nil, 4000001, "User ID is required", 400)
			return
		}
		
		// 模拟查找用户
		if id == "999" {
			ctx.Fail(nil, 4040001, "User not found", 404)
			return
		}
		
		ctx.Success(zoox.H{
			"id":   id,
			"name": "User " + id,
		})
	})
	
	app.Run(":8080")
}
```

**响应格式**:

成功响应：
```json
{
  "code": 200,
  "message": "success",
  "result": {...}
}
```

错误响应：
```json
{
  "code": 4000001,
  "message": "User ID is required"
}
```

## 下一步

- 📚 深入学习 [路由系统](../guides/routing.md)
- 🔌 了解 [中间件系统](../guides/middleware.md)
- 📝 查看 [Context API](../guides/context.md)
- 🚀 探索 [高级功能](../advanced/websocket.md)

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
