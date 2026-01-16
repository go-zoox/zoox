# RESTful API 示例

这是一个完整的 RESTful API 示例，展示了如何使用 Zoox 构建一个用户管理 API。

## 项目结构

```
rest-api/
├── main.go
├── go.mod
├── handlers/
│   └── user.go
├── models/
│   └── user.go
└── middleware/
    └── auth.go
```

## 完整代码

### main.go

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
	"rest-api/handlers"
	"rest-api/middleware"
)

func main() {
	app := zoox.New()
	
	// 全局中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	app.Use(middleware.RequestID())
	
	// 健康检查
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"status": "ok",
		})
	})
	
	// API 路由组
	api := app.Group("/api/v1")
	
	// 公共路由（不需要认证）
	api.Post("/login", handlers.Login)
	api.Post("/register", handlers.Register)
	
	// 受保护的路由（需要认证）
	protected := api.Group("")
	protected.Use(auth.RequireAuth())
	
	protected.Get("/users", handlers.GetUsers)
	protected.Get("/users/:id", handlers.GetUser)
	protected.Post("/users", handlers.CreateUser)
	protected.Put("/users/:id", handlers.UpdateUser)
	protected.Delete("/users/:id", handlers.DeleteUser)
	
	// 启动服务器
	app.Run(":8080")
}
```

### models/user.go

```go
package models

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"` // 不序列化密码
}

// 模拟数据库
var users = []*User{
	{ID: 1, Name: "Alice", Email: "alice@example.com", Password: "password1"},
	{ID: 2, Name: "Bob", Email: "bob@example.com", Password: "password2"},
}

var nextID = 3

func GetAllUsers() []*User {
	return users
}

func GetUserByID(id int) *User {
	for _, user := range users {
		if user.ID == id {
			return user
		}
	}
	return nil
}

func GetUserByEmail(email string) *User {
	for _, user := range users {
		if user.Email == email {
			return user
		}
	}
	return nil
}

func CreateUser(user *User) *User {
	user.ID = nextID
	nextID++
	users = append(users, user)
	return user
}

func UpdateUser(id int, user *User) *User {
	for i, u := range users {
		if u.ID == id {
			user.ID = id
			users[i] = user
			return user
		}
	}
	return nil
}

func DeleteUser(id int) bool {
	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			return true
		}
	}
	return false
}
```

### handlers/user.go

```go
package handlers

import (
	"strconv"
	"time"
	
	"github.com/go-zoox/zoox"
	"rest-api/models"
)

// Login 处理登录
func Login(ctx *zoox.Context) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	if err := ctx.BindJSON(&creds); err != nil {
		ctx.Fail(err, 4000001, "Invalid JSON", 400)
		return
	}
	
	user := models.GetUserByEmail(creds.Email)
	if user == nil || user.Password != creds.Password {
		ctx.Fail(nil, 4010001, "Invalid credentials", 401)
		return
	}
	
	// 生成 JWT Token
	jwt := ctx.Jwt()
	token, err := jwt.Sign(map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	
	if err != nil {
		ctx.Fail(err, 5000001, "Failed to generate token", 500)
		return
	}
	
	ctx.JSON(200, zoox.H{
		"token": token,
		"user": zoox.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// Register 处理注册
func Register(ctx *zoox.Context) {
	var user struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Fail(err, 4000001, "Invalid JSON", 400)
		return
	}
	
	if user.Name == "" || user.Email == "" || user.Password == "" {
		ctx.Fail(nil, 4000002, "Name, email and password are required", 400)
		return
	}
	
	// 检查邮箱是否已存在
	if models.GetUserByEmail(user.Email) != nil {
		ctx.Fail(nil, 4000003, "Email already exists", 400)
		return
	}
	
	// 创建用户
	newUser := &models.User{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}
	createdUser := models.CreateUser(newUser)
	
	ctx.JSON(201, zoox.H{
		"id":    createdUser.ID,
		"name":  createdUser.Name,
		"email": createdUser.Email,
	})
}

// GetUsers 获取所有用户
func GetUsers(ctx *zoox.Context) {
	users := models.GetAllUsers()
	
	// 移除密码字段
	result := make([]zoox.H, len(users))
	for i, user := range users {
		result[i] = zoox.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		}
	}
	
	ctx.Success(result)
}

// GetUser 获取单个用户
func GetUser(ctx *zoox.Context) {
	idStr := ctx.Param().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.Fail(err, 4000001, "Invalid user ID", 400)
		return
	}
	
	user := models.GetUserByID(id)
	if user == nil {
		ctx.Fail(nil, 4040001, "User not found", 404)
		return
	}
	
	ctx.Success(zoox.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

// CreateUser 创建用户
func CreateUser(ctx *zoox.Context) {
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Fail(err, 4000001, "Invalid JSON", 400)
		return
	}
	
	if user.Name == "" || user.Email == "" {
		ctx.Fail(nil, 4000002, "Name and email are required", 400)
		return
	}
	
	// 检查邮箱是否已存在
	if models.GetUserByEmail(user.Email) != nil {
		ctx.Fail(nil, 4000003, "Email already exists", 400)
		return
	}
	
	newUser := &models.User{
		Name:  user.Name,
		Email: user.Email,
	}
	createdUser := models.CreateUser(newUser)
	
	ctx.JSON(201, zoox.H{
		"id":    createdUser.ID,
		"name":  createdUser.Name,
		"email": createdUser.Email,
	})
}

// UpdateUser 更新用户
func UpdateUser(ctx *zoox.Context) {
	idStr := ctx.Param().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.Fail(err, 4000001, "Invalid user ID", 400)
		return
	}
	
	var user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Fail(err, 4000001, "Invalid JSON", 400)
		return
	}
	
	existingUser := models.GetUserByID(id)
	if existingUser == nil {
		ctx.Fail(nil, 4040001, "User not found", 404)
		return
	}
	
	updatedUser := &models.User{
		ID:    id,
		Name:  user.Name,
		Email: user.Email,
	}
	result := models.UpdateUser(id, updatedUser)
	
	ctx.Success(zoox.H{
		"id":    result.ID,
		"name":  result.Name,
		"email": result.Email,
	})
}

// DeleteUser 删除用户
func DeleteUser(ctx *zoox.Context) {
	idStr := ctx.Param().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.Fail(err, 4000001, "Invalid user ID", 400)
		return
	}
	
	if !models.DeleteUser(id) {
		ctx.Fail(nil, 4040001, "User not found", 404)
		return
	}
	
	ctx.JSON(200, zoox.H{"message": "User deleted"})
}
```

### middleware/auth.go

```go
package auth

import (
	"github.com/go-zoox/zoox"
)

// RequireAuth 要求认证的中间件
func RequireAuth() zoox.Middleware {
	return func(ctx *zoox.Context) {
		token, ok := ctx.BearerToken()
		if !ok {
			ctx.Fail(nil, 4010001, "Token required", 401)
			return
		}
		
		jwt := ctx.Jwt()
		claims, err := jwt.Verify(token)
		if err != nil {
			ctx.Fail(err, 4010002, "Invalid token", 401)
			return
		}
		
		// 将用户信息存储到 Context
		ctx.State().Set("user_id", claims["user_id"])
		ctx.State().Set("email", claims["email"])
		
		ctx.Next()
	}
}
```

## 运行和测试

### 启动服务器

```bash
go mod init rest-api
go get github.com/go-zoox/zoox
go run main.go
```

### 测试 API

```bash
# 注册用户
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"password123"}'

# 登录
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}'

# 获取所有用户（需要 Token）
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN"

# 获取单个用户
curl http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer YOUR_TOKEN"

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'

# 更新用户
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.updated@example.com"}'

# 删除用户
curl -X DELETE http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 特性说明

1. **RESTful 设计** - 遵循 REST 规范
2. **JWT 认证** - 使用 JWT 进行身份认证
3. **错误处理** - 统一的错误响应格式
4. **中间件** - 使用中间件进行认证和日志记录
5. **路由组** - 使用路由组组织代码

## 扩展建议

1. **数据库集成** - 替换内存存储为真实数据库
2. **密码加密** - 使用 bcrypt 加密密码
3. **输入验证** - 添加更严格的输入验证
4. **分页** - 为列表接口添加分页功能
5. **缓存** - 使用缓存提高性能

## 下一步

- 📡 查看 [实时应用示例](real-time-app.md) - WebSocket 应用
- 🏗️ 学习 [微服务示例](microservice.md) - 微服务架构
- 📚 阅读 [最佳实践](../best-practices.md) - 开发建议

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
