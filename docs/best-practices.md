# 最佳实践

本文档总结了使用 Zoox 框架的最佳实践和建议。

## 项目结构

### 推荐的项目结构

```
my-zoox-app/
├── main.go              # 应用入口
├── go.mod
├── go.sum
├── config/
│   └── config.go        # 配置管理
├── handlers/
│   ├── user.go          # 用户相关处理
│   └── post.go           # 文章相关处理
├── models/
│   ├── user.go          # 数据模型
│   └── post.go
├── middleware/
│   └── auth.go          # 自定义中间件
├── utils/
│   └── helpers.go       # 工具函数
└── templates/            # 模板文件（如果使用）
    └── *.html
```

### 组织路由

```go
// main.go
package main

import (
	"github.com/go-zoox/zoox"
	"my-app/handlers"
	"my-app/middleware"
)

func main() {
	app := zoox.New()
	
	// 全局中间件
	setupMiddleware(app)
	
	// 路由
	setupRoutes(app)
	
	app.Run(":8080")
}

func setupMiddleware(app *zoox.Application) {
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
}

func setupRoutes(app *zoox.Application) {
	// 公共路由
	app.Get("/", handlers.Home)
	app.Get("/health", handlers.Health)
	
	// API 路由
	api := app.Group("/api/v1")
	api.Use(middleware.JWT())
	
	api.Get("/users", handlers.GetUsers)
	api.Post("/users", handlers.CreateUser)
	api.Get("/users/:id", handlers.GetUser)
}
```

## 错误处理

### 统一错误格式

```go
// 定义错误码
const (
	ErrCodeInvalidParam = 4000001
	ErrCodeNotFound     = 4040001
	ErrCodeUnauthorized = 4010001
	ErrCodeInternal     = 5000001
)

// 使用 ctx.Fail() 返回业务错误
app.Post("/users", func(ctx *zoox.Context) {
	var user User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Fail(err, ErrCodeInvalidParam, "Invalid JSON", 400)
		return
	}
	
	if user.Name == "" {
		ctx.Fail(nil, ErrCodeInvalidParam, "Name is required", 400)
		return
	}
	
	// 业务逻辑
	ctx.Success(user)
})
```

### 错误处理中间件

```go
func ErrorHandler() zoox.Middleware {
	return func(ctx *zoox.Context) {
		ctx.Next()
		
		// 检查是否有错误
		if ctx.StatusCode() >= 400 {
			ctx.Logger.Errorf("Error: %d - %s", ctx.StatusCode(), ctx.Path)
		}
	}
}
```

## 配置管理

### 使用环境变量

```go
// config/config.go
package config

import "os"

type Config struct {
	Port      string
	SecretKey string
	DBHost    string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		SecretKey: getEnv("SECRET_KEY", ""),
		DBHost:    getEnv("DB_HOST", "localhost"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

### 区分环境

```go
app := zoox.New()

if app.IsProd() {
	// 生产环境配置
	app.Config.LogLevel = "info"
	app.Config.Monitor.Sentry.Enabled = true
} else {
	// 开发环境配置
	app.Config.LogLevel = "debug"
	app.Config.Monitor.Sentry.Enabled = false
}
```

## 性能优化

### 1. 使用缓存

```go
app.Get("/user/:id", func(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	cache := ctx.Cache()
	cacheKey := "user:" + id
	
	var user User
	if cache.Get(cacheKey, &user) == nil {
		ctx.JSON(200, user)
		return
	}
	
	// 从数据库获取
	user = getUserFromDB(id)
	cache.Set(cacheKey, user, time.Hour)
	
	ctx.JSON(200, user)
})
```

### 2. 启用 Gzip 压缩

```go
app.Use(middleware.Gzip())
```

### 3. 使用连接池

对于数据库连接，使用连接池：

```go
// 初始化数据库连接池
db, err := sql.Open("mysql", dsn)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
```

### 4. 异步处理

对于耗时操作，使用异步处理：

```go
app.Post("/email", func(ctx *zoox.Context) {
	var email Email
	ctx.BindJSON(&email)
	
	// 立即返回响应
	ctx.JSON(200, zoox.H{"message": "Email queued"})
	
	// 异步发送邮件
	go sendEmail(email)
})
```

## 安全建议

### 1. 使用 HTTPS

```go
app.Config.TLSCertFile = "/path/to/cert.pem"
app.Config.TLSKeyFile = "/path/to/key.pem"
app.Config.HTTPSPort = 8443
```

### 2. 设置安全响应头

```go
app.Use(middleware.Helmet(nil))
```

### 3. 验证和清理输入

```go
app.Post("/users", func(ctx *zoox.Context) {
	var user User
	ctx.BindJSON(&user)
	
	// 验证输入
	if err := validateUser(user); err != nil {
		ctx.Fail(err, ErrCodeInvalidParam, err.Error(), 400)
		return
	}
	
	// 清理输入（防止 XSS）
	user.Name = sanitize(user.Name)
	user.Email = sanitize(user.Email)
	
	// 处理
})
```

### 4. 使用强密钥

```go
// 推荐：使用随机生成的强密钥
import "github.com/go-zoox/random"

app.Config.SecretKey = random.String(32)

// 不推荐：使用弱密钥
app.Config.SecretKey = "123456"
```

### 5. 限制请求速率

```go
app.Use(middleware.RateLimit(&middleware.RateLimitConfig{
	Period: time.Minute,
	Limit:  100,
}))
```

## 日志记录

### 结构化日志

```go
app.Get("/users/:id", func(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	
	ctx.Logger.Infof("Getting user: %s", id)
	
	user := getUser(id)
	if user == nil {
		ctx.Logger.Warnf("User not found: %s", id)
		ctx.Error(404, "User not found")
		return
	}
	
	ctx.Logger.Infof("User found: %s", id)
	ctx.JSON(200, user)
})
```

### 请求追踪

```go
app.Use(middleware.RequestID())

app.Get("/users", func(ctx *zoox.Context) {
	requestID := ctx.RequestID()
	ctx.Logger.Infof("[%s] Getting users", requestID)
	
	// 处理逻辑
})
```

## 测试

### 单元测试

```go
func TestGetUser(t *testing.T) {
	app := zoox.New()
	app.Get("/users/:id", func(ctx *zoox.Context) {
		id := ctx.Param().Get("id")
		ctx.JSON(200, zoox.H{"id": id})
	})
	
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"123"`)
}
```

## 部署建议

### 1. 使用环境变量

```bash
export PORT=8080
export SECRET_KEY=your-secret-key
export LOG_LEVEL=info
```

### 2. 健康检查

```go
app.Get("/health", func(ctx *zoox.Context) {
	ctx.JSON(200, zoox.H{
		"status": "ok",
		"timestamp": time.Now(),
	})
})
```

### 3. 优雅关闭

```go
func main() {
	app := zoox.New()
	
	// 设置关闭钩子
	app.SetBeforeDestroy(func() {
		// 清理资源
		closeDB()
		closeRedis()
	})
	
	app.Run(":8080")
}
```

## 下一步

- 📚 查看 [完整文档索引](README.md)
- 🛣️ 学习 [路由系统](guides/routing.md)
- 🔌 了解 [中间件使用](guides/middleware.md)

---

**需要更多帮助？** 👉 [完整文档索引](README.md)
