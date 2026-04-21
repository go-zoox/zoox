# Context API 详解

Context 是 Zoox 框架的核心，它封装了 HTTP 请求和响应，提供了丰富的 API 来处理请求和生成响应。

## Context 概述

每个请求都会创建一个新的 Context 实例，它包含：

- 请求信息（方法、路径、参数、查询、表单、Body 等）
- 响应方法（JSON、HTML、String、Redirect 等）
- 工具方法（Cookie、Session、JWT、Cache 等）

**说明**: Context 结构定义参考 `context.go:59-152`。

## 请求数据获取

### 查询参数 (Query)

```go
// 获取单个查询参数
query := ctx.Query().Get("q")

// 获取整数查询参数（带默认值）
page := ctx.Query().GetInt("page", 1)

// 获取所有查询参数
queries := ctx.Queries()
```

**说明**: Query 方法参考 `context.go:215-222`。

### 路由参数 (Param)

```go
// 获取路由参数
id := ctx.Param().Get("id")

// 获取所有路由参数
params := ctx.Params()
```

**说明**: Param 方法参考 `context.go:224-227`。

### 请求体 (Body)

```go
// 获取原始 Body 字节
bodyBytes, err := ctx.BodyBytes()

// 获取 Body 对象
body := ctx.Body()

// 获取 Body 数据（解析为 map）
bodies := ctx.Bodies()
```

**说明**: Body 方法参考 `context.go:243-250`。

### 表单数据 (Form)

```go
// 获取表单字段
name := ctx.Form().Get("name")

// 获取所有表单数据
forms, err := ctx.Forms()
```

**说明**: Form 方法参考 `context.go:234-241`。

### 请求头 (Header)

```go
// 获取请求头
auth := ctx.Header().Get("Authorization")

// 获取所有请求头
headers := ctx.Headers()

// 常用请求头快捷方法
userAgent := ctx.UserAgent()
contentType := ctx.ContentType()
authorization := ctx.Authorization()
```

**说明**: Header 方法参考 `context.go:229-232`。

### 文件上传

```go
// 获取上传的文件
file, fileHeader, err := ctx.File("file")
if err != nil {
	ctx.Error(400, "No file uploaded")
	return
}
defer file.Close()

// 获取所有上传的文件
files := ctx.Files()
```

## 数据绑定

Zoox 提供了强大的数据绑定功能，可以将请求数据自动绑定到结构体。

### 绑定 JSON

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

app.Post("/users", func(ctx *zoox.Context) {
	var user User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Error(400, "Invalid JSON")
		return
	}
	
	ctx.JSON(200, user)
})
```

**说明**: BindJSON 方法参考 `context.go:766-801`。

### 绑定查询参数

```go
type QueryParams struct {
	Page  int    `query:"page"`
	Size  int    `query:"size"`
	Order string `query:"order"`
}

app.Get("/search", func(ctx *zoox.Context) {
	var params QueryParams
	if err := ctx.BindQuery(&params); err != nil {
		ctx.Error(400, "Invalid query parameters")
		return
	}
	
	ctx.JSON(200, params)
})
```

**说明**: BindQuery 方法参考 `context.go:865-876`。

### 与 gormx 查询参数适配

当你在 Zoox 中使用 `github.com/go-zoox/gormx` 时，推荐使用 Zoox 侧适配包：

```go
import (
	"github.com/go-zoox/zoox"
	contextgormx "github.com/go-zoox/zoox/components/context/gormx"
)

app.Get("/users/:id", func(ctx *zoox.Context) {
	params := contextgormx.NewParams(ctx)

	list, err := params.GetList()
	if err != nil {
		ctx.Error(400, err.Error())
		return
	}

	whereQuery, whereArgs, _ := list.Where.Build()
	orderBy := list.OrderBy.Build()

	ctx.JSON(200, zoox.H{
		"page":       list.Page,
		"page_size":  list.PageSize,
		"whereQuery": whereQuery,
		"whereArgs":  whereArgs,
		"orderBy":    orderBy,
	})
})
```

常见查询参数示例：

- `?page=2&pageSize=20`
- `?orderBy=created_at:desc,id:asc`
- `?age=18,65:range[)`（区间语法）

这样可以保持 `gormx` 和 `zoox` 的解耦：`gormx` 只依赖通用接口，`zoox` 提供上下文适配实现。

### 绑定表单数据

```go
type FormData struct {
	Name  string `form:"name"`
	Email string `form:"email"`
}

app.Post("/register", func(ctx *zoox.Context) {
	var data FormData
	if err := ctx.BindForm(&data); err != nil {
		ctx.Error(400, "Invalid form data")
		return
	}
	
	ctx.JSON(200, data)
})
```

**说明**: BindForm 方法参考 `context.go:822-837`。

### 绑定路由参数

```go
type RouteParams struct {
	ID string `param:"id"`
}

app.Get("/users/:id", func(ctx *zoox.Context) {
	var params RouteParams
	if err := ctx.BindParams(&params); err != nil {
		ctx.Error(400, "Invalid parameters")
		return
	}
	
	ctx.JSON(200, params)
})
```

**说明**: BindParams 方法参考 `context.go:839-850`。

### 绑定请求头

```go
type Headers struct {
	Authorization string `header:"Authorization"`
	UserAgent     string `header:"User-Agent"`
}

app.Get("/info", func(ctx *zoox.Context) {
	var headers Headers
	if err := ctx.BindHeader(&headers); err != nil {
		ctx.Error(400, "Invalid headers")
		return
	}
	
	ctx.JSON(200, headers)
})
```

**说明**: BindHeader 方法参考 `context.go:852-863`。

### 绑定 Body（通用）

```go
type BodyData struct {
	Name  string `body:"name"`
	Email string `body:"email"`
}

app.Post("/data", func(ctx *zoox.Context) {
	var data BodyData
	if err := ctx.BindBody(&data); err != nil {
		ctx.Error(400, "Invalid body")
		return
	}
	
	ctx.JSON(200, data)
})
```

**说明**: BindBody 方法参考 `context.go:878-889`。

## 响应方法

### JSON 响应

```go
// 基本 JSON 响应
ctx.JSON(200, zoox.H{
	"message": "Success",
	"data":    data,
})

// 使用结构体
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

ctx.JSON(200, Response{
	Code:    200,
	Message: "Success",
})
```

**说明**: JSON 方法参考 `context.go:402-413`。

### HTML 响应

```go
// 直接 HTML 字符串
ctx.HTML(200, "<h1>Hello, Zoox!</h1>")

// 带数据的 HTML（使用模板）
ctx.HTML(200, "<h1>{{.Title}}</h1>", zoox.H{
	"Title": "Welcome",
})
```

**说明**: HTML 方法参考 `context.go:423-432`。

### 字符串响应

```go
ctx.String(200, "Plain text response")
```

**说明**: String 方法参考 `context.go:396-400`。

### 数据响应

```go
ctx.Data(200, "application/octet-stream", []byte("binary data"))
```

**说明**: Data 方法参考 `context.go:416-421`。

### 重定向

```go
// 临时重定向（302）
ctx.Redirect("http://example.com")

// 永久重定向（301）
ctx.RedirectPermanent("http://example.com")

// 临时重定向（307）
ctx.RedirectTemporary("http://example.com")

// 查看其他（303）
ctx.RedirectSeeOther("http://example.com")
```

**说明**: Redirect 方法参考 `context.go:548-583`。

### 成功响应

```go
ctx.Success(zoox.H{
	"users": []zoox.H{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	},
})
```

响应格式：
```json
{
  "code": 200,
  "message": "success",
  "result": {...}
}
```

**说明**: Success 方法参考 `context.go:499-506`。

### 错误响应

```go
// 系统错误
ctx.Error(404, "Not Found")

// 业务错误
ctx.Fail(err, 4000001, "Invalid parameter", 400)
```

**说明**: Error 和 Fail 方法参考 `context.go:477-497, 512-541`。

## Cookie 和 Session

### Cookie 操作

```go
// 设置 Cookie
ctx.Cookie().Set("username", "alice", 3600)  // 1小时过期

// 获取 Cookie
username := ctx.Cookie().Get("username")

// 删除 Cookie
ctx.Cookie().Delete("username")

// 获取所有 Cookie
cookies := ctx.Cookies()
```

**说明**: Cookie 方法参考 `context.go:1009-1019`。

### Session 操作

```go
// 设置 Session
ctx.Session().Set("user_id", 123)
ctx.Session().Set("username", "alice")

// 获取 Session
userID := ctx.Session().Get("user_id")
username := ctx.Session().Get("username")

// 删除 Session
ctx.Session().Delete("user_id")

// 清除所有 Session
ctx.Session().Clear()
```

**说明**: Session 方法参考 `context.go:1021-1033`。

## JWT 操作

```go
// 生成 JWT Token
jwt := ctx.Jwt()
token, err := jwt.Sign(map[string]interface{}{
	"user_id":  1,
	"username": "alice",
	"exp":      time.Now().Add(24 * time.Hour).Unix(),
})

// 验证 JWT Token
claims, err := jwt.Verify(token)
if err != nil {
	ctx.Error(401, "Invalid token")
	return
}

userID := claims["user_id"]
```

**说明**: JWT 方法参考 `context.go:1035-1047`。

## 客户端信息

### IP 地址

```go
// 获取客户端 IP（自动处理代理）
ip := ctx.IP()

// 获取所有 IP（X-Forwarded-For）
ips := ctx.IPs()

// 客户端 IP（别名）
clientIP := ctx.ClientIP()
```

**说明**: IP 方法参考 `context.go:614-648`。

### 主机信息

```go
// 获取主机名
hostname := ctx.Hostname()

// 获取主机（包含端口）
host := ctx.Host()

// 获取协议
protocol := ctx.Protocol()

// 获取完整 URL
url := ctx.URL()
```

**说明**: 主机信息方法参考 `context.go:585-611`。

### User Agent

```go
ua := ctx.UserAgent()

// User Agent 信息
browser := ua.Browser()
os := ua.OS()
device := ua.Device()
```

**说明**: UserAgent 方法参考 `context.go:331-334`。

## 其他工具方法

### 请求 ID

```go
requestID := ctx.RequestID()
```

### 状态码

```go
// 设置状态码
ctx.Status(200)

// 获取状态码
status := ctx.StatusCode()
```

**说明**: Status 方法参考 `context.go:252-260`。

### 响应头

```go
// 设置响应头
ctx.SetHeader("Content-Type", "application/json")

// 添加响应头
ctx.AddHeader("X-Custom-Header", "value")

// 快捷方法
ctx.Set("Content-Type", "application/json")
```

**说明**: SetHeader 方法参考 `context.go:272-280`。

### 内容类型检测

```go
// 检测是否接受 JSON
if ctx.AcceptJSON() {
	ctx.JSON(200, data)
}

// 检测是否接受 HTML
if ctx.AcceptHTML() {
	ctx.HTML(200, html)
}
```

**说明**: Accept 方法参考 `context.go:912-926`。

### 连接升级检测

```go
// 检测是否为 WebSocket 升级请求
if ctx.IsConnectionUpgrade() {
	// 处理 WebSocket
}
```

## 应用组件访问

### Cache

```go
cache := ctx.Cache()

// 设置缓存
cache.Set("key", "value", time.Hour)

// 获取缓存
var value string
cache.Get("key", &value)
```

### Cron

```go
cron := ctx.Cron()

// 添加定时任务
cron.AddJob("daily-cleanup", "0 0 * * *", func() error {
	// 每天午夜执行
	return nil
})
```

### JobQueue

```go
queue := ctx.JobQueue()

// 添加任务
queue.Add("task-name", data)
```

### Logger

```go
ctx.Logger.Info("Info message")
ctx.Logger.Error("Error message")
ctx.Logger.Debug("Debug message")
```

### Debug

```go
if ctx.Debug().IsDebugMode() {
	// 调试模式下的逻辑
}
```

### I18n

```go
i18n := ctx.I18n()

// 翻译文本
text := i18n.T("hello")
```

## Context 最佳实践

### 1. 错误处理

```go
app.Post("/users", func(ctx *zoox.Context) {
	var user User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Fail(err, 4000001, "Invalid JSON", 400)
		return
	}
	
	// 业务逻辑
	ctx.Success(user)
})
```

### 2. 数据验证

```go
app.Post("/users", func(ctx *zoox.Context) {
	var user User
	if err := ctx.BindJSON(&user); err != nil {
		ctx.Fail(err, 4000001, "Invalid JSON", 400)
		return
	}
	
	// 验证数据
	if user.Name == "" {
		ctx.Fail(nil, 4000002, "Name is required", 400)
		return
	}
	
	ctx.Success(user)
})
```

### 3. 统一响应格式

```go
// 成功响应
ctx.Success(data)

// 业务错误
ctx.Fail(err, code, message, status)

// 系统错误
ctx.Error(status, message)
```

### 4. 使用绑定而非手动解析

```go
// 推荐：使用绑定
var user User
ctx.BindJSON(&user)

// 不推荐：手动解析
body, _ := ctx.BodyBytes()
json.Unmarshal(body, &user)
```

## 下一步

- 🛣️ 学习 [路由系统](routing.md) - 了解路由参数和路由组
- 🔌 查看 [中间件使用](middleware.md) - 在中间件中使用 Context
- 🚀 探索 [高级功能](../advanced/websocket.md) - WebSocket、SSE 等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
