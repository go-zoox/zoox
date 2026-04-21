# Context API 参考

Context 封装了 HTTP 请求和响应，提供了丰富的 API。

## 请求数据

### Query() query.Query

获取查询参数。

```go
query := ctx.Query().Get("page")
page := ctx.Query().GetInt("page", 1)
```

### Param() param.Param

获取路由参数。

```go
id := ctx.Param().Get("id")
```

### Form() form.Form

获取表单数据。

```go
name := ctx.Form().Get("name")
```

### Body() body.Body

获取请求体。

```go
body := ctx.Body()
bodyBytes, _ := ctx.BodyBytes()
```

### Header() http.Header

获取请求头。

```go
auth := ctx.Header().Get("Authorization")
```

### File(key string) (multipart.File, *multipart.FileHeader, error)

获取上传的文件。

```go
file, fileHeader, err := ctx.File("file")
```

## 数据绑定

### BindJSON(obj interface{}) error

绑定 JSON 请求体到结构体。

```go
var user User
ctx.BindJSON(&user)
```

### BindQuery(obj interface{}) error

绑定查询参数到结构体。

```go
var params QueryParams
ctx.BindQuery(&params)
```

### gormx 查询参数适配

如果你使用 `gormx`，可以通过 Zoox 适配层直接复用 `gormx.Params`：

```go
import contextgormx "github.com/go-zoox/zoox/components/context/gormx"

params := contextgormx.NewParams(ctx)
list, err := params.GetList()
```

更多示例见 [gormx 组件文档](../components/gormx.md)。

### BindForm(obj interface{}) error

绑定表单数据到结构体。

```go
var data FormData
ctx.BindForm(&data)
```

### BindParams(obj interface{}) error

绑定路由参数到结构体。

```go
var params RouteParams
ctx.BindParams(&params)
```

### BindHeader(obj interface{}) error

绑定请求头到结构体。

```go
var headers Headers
ctx.BindHeader(&headers)
```

### BindBody(obj interface{}) error

绑定请求体到结构体。

```go
var data BodyData
ctx.BindBody(&data)
```

## 响应方法

### JSON(status int, obj interface{})

返回 JSON 响应。

```go
ctx.JSON(200, zoox.H{"message": "Success"})
```

### HTML(status int, html string, data ...any)

返回 HTML 响应。

```go
ctx.HTML(200, "<h1>Hello</h1>")
```

### String(status int, text string)

返回字符串响应。

```go
ctx.String(200, "Hello, World")
```

### Data(status int, contentType string, data []byte)

返回数据响应。

```go
ctx.Data(200, "application/octet-stream", []byte("data"))
```

### Redirect(url string, status ...int)

重定向。

```go
ctx.Redirect("http://example.com")
ctx.RedirectPermanent("http://example.com")
ctx.RedirectTemporary("http://example.com")
```

### Success(result interface{})

返回成功响应。

```go
ctx.Success(zoox.H{"data": data})
```

### Error(status int, message string)

返回错误响应。

```go
ctx.Error(404, "Not Found")
```

### Fail(err error, code int, message string, status ...int)

返回业务错误。

```go
ctx.Fail(err, 4000001, "Invalid parameter", 400)
```

## Cookie 和 Session

### Cookie() cookie.Cookie

获取 Cookie 实例。

```go
ctx.Cookie().Set("key", "value", 3600)
value := ctx.Cookie().Get("key")
```

### Session() session.Session

获取 Session 实例。

```go
ctx.Session().Set("user_id", 123)
userID := ctx.Session().Get("user_id")
```

## JWT

### Jwt() jwt.Jwt

获取 JWT 实例。

```go
jwt := ctx.Jwt()
token, _ := jwt.Sign(claims)
claims, _ := jwt.Verify(token)
```

## 客户端信息

### IP() string

获取客户端 IP。

```go
ip := ctx.IP()
```

### IPs() []string

获取所有 IP（X-Forwarded-For）。

```go
ips := ctx.IPs()
```

### Hostname() string

获取主机名。

```go
hostname := ctx.Hostname()
```

### UserAgent() useragent.UserAgent

获取 User Agent。

```go
ua := ctx.UserAgent()
browser := ua.Browser()
```

## 工具方法

### RequestID() string

获取请求 ID。

```go
requestID := ctx.RequestID()
```

### Status(status int)

设置状态码。

```go
ctx.Status(200)
```

### StatusCode() int

获取状态码。

```go
status := ctx.StatusCode()
```

### SetHeader(key, value string)

设置响应头。

```go
ctx.SetHeader("Content-Type", "application/json")
```

### Next()

执行下一个中间件或处理函数。

```go
ctx.Next()
```

## 组件访问

### Cache() cache.Cache

获取缓存实例。

```go
cache := ctx.Cache()
```

### Cron() cron.Cron

获取定时任务实例。

```go
cron := ctx.Cron()
```

### JobQueue() jobqueue.JobQueue

获取任务队列实例。

```go
queue := ctx.JobQueue()
```

### Logger() *logger.Logger

获取日志实例。

```go
logger := ctx.Logger()
```

### Debug() debug.Debug

获取调试实例。

```go
debug := ctx.Debug()
```

### Env() env.Env

获取环境变量实例。

```go
env := ctx.Env()
```

## 完整示例

```go
app.Get("/users/:id", func(ctx *zoox.Context) {
	// 获取路由参数
	id := ctx.Param().Get("id")
	
	// 获取查询参数
	page := ctx.Query().GetInt("page", 1)
	
	// 从缓存获取
	cache := ctx.Cache()
	var user User
	if cache.Get("user:"+id, &user) == nil {
		ctx.JSON(200, user)
		return
	}
	
	// 从数据库获取
	user = getUserFromDB(id)
	
	// 写入缓存
	cache.Set("user:"+id, user, time.Hour)
	
	// 返回响应
	ctx.Success(user)
})
```

## 下一步

- 📝 查看 [Application API](application.md) - 应用方法参考
- 🛣️ 学习 [Router API](router.md) - 路由相关方法
- 🔌 了解 [中间件列表](middleware-list.md) - 所有内置中间件

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
