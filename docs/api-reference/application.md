# Application API 参考

Application 是 Zoox 框架的核心，代表整个 Web 应用。

## 创建应用

### New()

创建一个新的应用实例。

```go
app := zoox.New()
```

**说明**: 实现参考 `application.go:132-157`。

### Default()

创建一个带默认中间件的应用实例。

```go
app := zoox.Default()
```

默认中间件包括：
- Recovery
- RequestID
- RealIP
- Logger
- HealthCheck
- Runtime

**说明**: 实现参考 `defaults/defaults.go:12-115`。

## 启动服务器

### Run(addr ...string)

启动 HTTP 服务器。

```go
// 使用默认配置
app.Run()

// 指定端口
app.Run(":8080")

// 指定主机和端口
app.Run("127.0.0.1:8080")

// 使用 HTTP URL
app.Run("http://127.0.0.1:8080")

// 使用 Unix Domain Socket
app.Run("unix:///tmp/app.sock")
```

**说明**: 实现参考 `application.go:297-330`。

### Listen(port int)

启动服务器（必须指定端口）。

```go
app.Listen(8080)
```

**说明**: 实现参考 `application.go:332-337`。

## 路由注册

### Get/Post/Put/Patch/Delete/Head/Options/Connect

注册 HTTP 方法路由。

```go
app.Get("/", handler)
app.Post("/users", handler)
app.Put("/users/:id", handler)
app.Patch("/users/:id", handler)
app.Delete("/users/:id", handler)
```

**说明**: 实现参考 `group.go:79-125`。

### Any(path string, handler ...HandlerFunc)

注册所有 HTTP 方法的路由。

```go
app.Any("/all", handler)
```

### Group(prefix string, cb ...GroupFunc)

创建路由组。

```go
api := app.Group("/api/v1")
api.Get("/users", handler)
```

**说明**: 实现参考 `group.go:39-50`。

## 中间件

### Use(middlewares ...HandlerFunc)

注册全局中间件。

```go
app.Use(middleware.Logger())
app.Use(middleware.Recovery())
```

**说明**: 实现参考 `group.go:219-222`。

## 静态文件

### Static(basePath, rootDir string, options ...*StaticOptions)

提供静态文件服务。

```go
app.Static("/static", "./public")
```

**说明**: 实现参考 `group.go:345-401`。

### StaticFS(relativePath string, fs http.FileSystem)

使用自定义文件系统提供静态文件服务。

```go
app.StaticFS("/static", http.Dir("./public"))
```

## 404 处理

### NotFound(h HandlerFunc)

设置自定义 404 处理函数。

```go
app.NotFound(func(ctx *zoox.Context) {
	ctx.JSON(404, zoox.H{"error": "Not Found"})
})
```

### Fallback(h HandlerFunc)

NotFound 的别名。

```go
app.Fallback(func(ctx *zoox.Context) {
	ctx.JSON(404, zoox.H{"error": "Not Found"})
})
```

**说明**: 实现参考 `application.go:159-167`。

## 模板

### SetTemplates(dir string, fns ...template.FuncMap)

设置模板目录和自定义函数。

```go
app.SetTemplates("./templates/*", template.FuncMap{
	"upper": strings.ToUpper,
})
```

**说明**: 实现参考 `application.go:343-350`。

## 配置

### Config

应用配置对象。

```go
app.Config.Port = 8080
app.Config.Host = "0.0.0.0"
app.Config.SecretKey = "your-secret-key"
app.Config.LogLevel = "info"
```

**说明**: 配置结构参考 `config/config.go:8-50`。

## 组件访问

### Cache() cache.Cache

获取缓存实例。

```go
cache := app.Cache()
cache.Set("key", "value", time.Hour)
```

`app.Cache()` 首次初始化后，会自动注册到全局入口 `components/application/cache`，便于在非请求链路代码中复用同一个实例：

```go
import appcache "github.com/go-zoox/zoox/components/application/cache"

cache := appcache.Get()
```

**说明**: 实现参考 `application.go` 和 `components/application/cache/cache.go`。

### Cron() cron.Cron

获取定时任务实例。

```go
cron := app.Cron()
cron.AddJob("task", "0 * * * *", handler)
```

**说明**: 实现参考 `application.go:448-455`。

### JobQueue() jobqueue.JobQueue

获取任务队列实例。

```go
queue := app.JobQueue()
queue.Add("task", data)
```

**说明**: 实现参考 `application.go:457-464`。

### JSONRPCRegistry() jsonrpcServer.Server

获取 JSON-RPC 注册表。

```go
registry := app.JSONRPCRegistry()
registry.Register("method", handler)
```

**说明**: 实现参考 `application.go:392-399`。

### PubSub() pubsub.PubSub

获取发布订阅实例。

```go
pubsub := app.PubSub()
pubsub.Publish("channel", "message")
```

**说明**: 实现参考 `application.go:401-418`。

### MQ() mq.MQ

获取消息队列实例。

```go
mq := app.MQ()
mq.Publish("queue", "message")
```

**说明**: 实现参考 `application.go:420-437`。

### Logger() *logger.Logger

获取日志实例。

```go
logger := app.Logger()
logger.Info("Message")
```

**说明**: 实现参考 `application.go:493-503`。

### Env() env.Env

获取环境变量实例。

```go
env := app.Env()
mode := env.Get("MODE")
```

**说明**: 实现参考 `application.go:484-491`。

### Debug() debug.Debug

获取调试实例。

```go
debug := app.Debug()
if debug.IsDebugMode() {
	// 调试逻辑
}
```

**说明**: 实现参考 `application.go:505-512`。

### Runtime() runtime.Runtime

获取运行时信息实例。

```go
runtime := app.Runtime()
runtime.Print()
```

**说明**: 实现参考 `application.go:514-521`。

## 生命周期钩子

### SetBeforeReady(fn func())

设置服务器启动前的回调。

```go
app.SetBeforeReady(func() {
	// 初始化数据库
	// 加载配置
})
```

**说明**: 实现参考 `application.go:357-360`。

### SetBeforeDestroy(fn func())

设置服务器关闭前的回调。

```go
app.SetBeforeDestroy(func() {
	// 关闭数据库连接
	// 清理资源
})
```

**说明**: 实现参考 `application.go:362-365`。

## TLS 配置

### SetTLSCertLoader(loader func(sni string) (key, cert string, err error))

设置动态 TLS 证书加载器（支持 SNI）。

```go
app.SetTLSCertLoader(func(sni string) (key, cert string, err error) {
	// 根据 SNI 加载证书
	return loadCertForDomain(sni)
})
```

**说明**: 实现参考 `application.go:382-385`。

## 工具方法

### IsProd() bool

检查是否为生产环境。

```go
if app.IsProd() {
	// 生产环境逻辑
}
```

**说明**: 实现参考 `application.go:387-390`。

### Address() string

获取服务器地址。

```go
addr := app.Address()  // "0.0.0.0:8080"
```

### AddressHTTPS() string

获取 HTTPS 服务器地址。

```go
addr := app.AddressHTTPS()  // "0.0.0.0:8443"
```

## 完整示例

```go
package main

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	// 创建应用
	app := zoox.New()
	
	// 配置
	app.Config.Port = 8080
	app.Config.SecretKey = "your-secret-key"
	
	// 中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	
	// 路由
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{"message": "Hello"})
	})
	
	// 生命周期钩子
	app.SetBeforeReady(func() {
		app.Logger().Info("Server starting...")
	})
	
	app.SetBeforeDestroy(func() {
		app.Logger().Info("Server shutting down...")
	})
	
	// 启动服务器
	app.Run(":8080")
}
```

## 下一步

- 📝 查看 [Context API](context.md) - Context 方法参考
- 🛣️ 学习 [Router API](router.md) - 路由相关方法
- 🔌 了解 [中间件列表](middleware-list.md) - 所有内置中间件

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
