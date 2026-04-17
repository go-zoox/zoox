# 配置管理

Zoox 提供了灵活的配置管理机制，支持环境变量、代码配置和默认值。

## 配置结构

Zoox 的配置结构定义在 `config.Config` 中：

```go
type Config struct {
	Protocol  string
	Host      string
	Port      int
	HTTPSPort int

	EnableH2C         bool // 明文 HTTP/2（h2c），仅建议在可信网络使用
	EnableHTTP3       bool // 与 HTTPS 同时启用时在 UDP 上提供 HTTP/3
	HTTP3Port         int  // HTTP/3 UDP 端口，0 表示与 HTTPSPort 相同
	HTTP3AltSvcMaxAge int  // HTTPS 响应 Alt-Svc 的 ma= 秒数；0 为默认 86400；负数关闭 Alt-Svc
	
	BodySizeLimit int64
	
	NetworkType      string
	UnixDomainSocket string
	
	// TLS
	TLSCertFile string
	TLSKeyFile  string
	TLSCaCertFile string
	TLSCert     string
	TLSKey      string
	
	LogLevel  string
	SecretKey string
	
	Session session.Config
	Cache   cache.Config
	Redis   Redis
	
	Banner  string
	Monitor Monitor
	Logger  Logger
}
```

**说明**: 配置结构参考 `config/config.go:8-50`。

## 基本配置

### 服务器配置

```go
app := zoox.New()

// 协议（http 或 https）
app.Config.Protocol = "http"

// 主机地址
app.Config.Host = "0.0.0.0"  // 默认值

// 端口
app.Config.Port = 8080       // 默认值

// HTTPS 端口（如果设置了，会同时启动 HTTPS 服务）
app.Config.HTTPSPort = 8443
```

### 启动服务器

```go
// 方式1：使用配置的端口
app.Run()

// 方式2：指定端口
app.Run(":3000")

// 方式3：指定主机和端口
app.Run("127.0.0.1:3000")

// 方式4：使用 HTTP URL
app.Run("http://127.0.0.1:3000")

// 方式5：使用 Unix Domain Socket
app.Run("unix:///tmp/app.sock")
```

**说明**: Run 方法参考 `application.go:297-330`，地址解析参考 `application.go:580-623`。

## 环境变量配置

Zoox 支持通过环境变量配置，会自动从环境变量读取配置：

### 内置环境变量

```bash
# 服务器配置
export PORT=8080
export HTTPS_PORT=8443
export ENABLE_H2C=true
export ENABLE_HTTP3=true
export HTTP3_PORT=8443
export HTTP3_ALTSVC_MAX_AGE=86400
export MODE=production

# 日志配置
export LOG_LEVEL=info

# 密钥配置
export SECRET_KEY=your-secret-key

# Session 配置
export SESSION_MAX_AGE=24h

# Redis 配置
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_USER=root
export REDIS_PASS=password
export REDIS_DB=0

# 监控配置
export MONITOR_PROMETHEUS_ENABLED=true
export MONITOR_PROMETHEUS_PATH=/metrics
export MONITOR_SENTRY_ENABLED=true
export MONITOR_SENTRY_DSN=your-sentry-dsn
```

**说明**: 环境变量常量参考 `constants.go:38-65`，环境变量解析参考 `application.go:217-281`。

### 使用 .env 文件

创建 `.env` 文件：

```env
PORT=8080
LOG_LEVEL=info
SECRET_KEY=your-secret-key
REDIS_HOST=localhost
REDIS_PORT=6379
```

在代码中加载：

```go
import "github.com/joho/godotenv"

func main() {
	// 加载 .env 文件
	godotenv.Load()
	
	app := zoox.New()
	app.Run()
}
```

## TLS/HTTPS 配置

### 方式1：使用文件

```go
app := zoox.New()

app.Config.TLSCertFile = "/path/to/cert.pem"
app.Config.TLSKeyFile = "/path/to/key.pem"
app.Config.HTTPSPort = 8443

app.Run()
```

### 方式2：使用内存中的证书

```go
app := zoox.New()

app.Config.TLSCert = "-----BEGIN CERTIFICATE-----\n..."
app.Config.TLSKey = "-----BEGIN PRIVATE KEY-----\n..."
app.Config.HTTPSPort = 8443

app.Run()
```

### 方式3：使用 SNI（动态证书加载）

```go
app := zoox.New()

app.SetTLSCertLoader(func(sni string) (key, cert string, err error) {
	// 根据 SNI 动态加载证书
	key = loadKeyForDomain(sni)
	cert = loadCertForDomain(sni)
	return key, cert, nil
})

app.Config.HTTPSPort = 8443
app.Run()
```

### 客户端证书验证

```go
app := zoox.New()

app.Config.TLSCaCertFile = "/path/to/ca.pem"  // CA 证书
app.Config.TLSCertFile = "/path/to/cert.pem"
app.Config.TLSKeyFile = "/path/to/key.pem"
app.Config.HTTPSPort = 8443

app.Run()
```

**说明**: TLS 与 HTTPS 服务实现参考 `application.go` 中的 `buildTLSConfig`、`serveHTTPS`。

### HTTP/2（HTTPS）

在已配置 `HTTPSPort` 与 TLS 证书时，HTTPS 监听器会通过 ALPN 同时提供 **HTTP/2**（`h2`）与 HTTP/1.1。框架会调用 `golang.org/x/net/http2` 的 `ConfigureServer` 注册 HTTP/2，并在 TLS 配置中声明 `h2`、`http/1.1` 的 ALPN。

可用 `curl -I --http2 https://<host>:<HTTPS端口>` 或浏览器开发者工具中的 Protocol 列验证。

### 明文 HTTP/2（h2c）

将 `app.Config.EnableH2C = true`（或环境变量 `ENABLE_H2C=true`）时，在 **TCP** 明文 HTTP 端口上启用 h2c。仅建议在受信任网络或反向代理之后使用；公网暴露风险较高。

### HTTP/3（QUIC）

1. 配置 `HTTPSPort` 与 TLS（HTTP/3 与 HTTPS 共用证书逻辑）。
2. 设置 `app.Config.EnableHTTP3 = true`（或 `ENABLE_HTTP3=true`）。
3. 可选：`app.Config.HTTP3Port`（或 `HTTP3_PORT`）指定 UDP 端口；为 `0` 时使用与 `HTTPSPort` 相同的端口号（生产环境常见：TCP 与 UDP 同端口，如 443）。
4. 防火墙与安全组需 **放行对应 UDP 端口**。
5. 启用 HTTP/3 时，HTTPS 响应会默认携带 `Alt-Svc`（引导浏览器使用 HTTP/3）。可通过 `HTTP3AltSvcMaxAge` / `HTTP3_ALTSVC_MAX_AGE` 调整 `ma=`；设为 **负数** 可关闭 `Alt-Svc` 头。

若由 Nginx、Caddy 等在边缘终止 TLS 并处理 HTTP/2/HTTP/3，可不在应用内开启 HTTP/3。

## Session 配置

```go
app := zoox.New()

// Session 最大存活时间
app.Config.Session.MaxAge = 24 * time.Hour

// 其他 Session 配置
app.Config.Session.HttpOnly = true
app.Config.Session.Secure = true
app.Config.Session.SameSite = "Lax"
```

**说明**: Session 配置参考 `config/config.go:38`。

## Cache 配置

```go
app := zoox.New()

// 使用内存缓存
app.Config.Cache = kv.Config{
	Engine: "memory",
}

// 使用 Redis 缓存
app.Config.Cache = kv.Config{
	Engine: "redis",
	Config: &redis.Config{
		Host:     "localhost",
		Port:     6379,
		Password: "password",
		DB:       0,
	},
}
```

**说明**: Cache 配置参考 `config/config.go:40`。

## Redis 配置

```go
app := zoox.New()

app.Config.Redis.Host = "localhost"
app.Config.Redis.Port = 6379
app.Config.Redis.Username = "root"
app.Config.Redis.Password = "password"
app.Config.Redis.DB = 0
```

**说明**: Redis 配置参考 `config/redis.go`。

## 日志配置

```go
app := zoox.New()

// 日志级别
app.Config.LogLevel = "info"  // debug, info, warn, error

// Logger 配置
app.Config.Logger.Level = "info"
app.Config.Logger.Middleware.Disabled = false
```

**说明**: Logger 配置参考 `config/logger.go`。

## 监控配置

### Prometheus

```go
app := zoox.New()

app.Config.Monitor.Prometheus.Enabled = true
app.Config.Monitor.Prometheus.Path = "/metrics"
```

### Sentry

```go
app := zoox.New()

app.Config.Monitor.Sentry.Enabled = true
app.Config.Monitor.Sentry.DSN = "your-sentry-dsn"
app.Config.Monitor.Sentry.Debug = false
app.Config.Monitor.Sentry.WaitForDelivery = false
app.Config.Monitor.Sentry.Timeout = 5 * time.Second
```

**说明**: 监控配置参考 `config/monitor.go`。

## 默认配置

Zoox 会自动应用默认配置：

```go
// 默认配置值
Protocol: "http"
Host: "0.0.0.0"
Port: 8080
SecretKey: random.String(16)  // 随机生成
Session.MaxAge: 24 * time.Hour
```

**说明**: 默认配置应用参考 `application.go:169-215`。

## 配置优先级

配置优先级（从高到低）：

1. **代码配置** - 直接在代码中设置
2. **环境变量** - 从环境变量读取
3. **默认值** - 框架提供的默认值

```go
app := zoox.New()

// 1. 代码配置（最高优先级）
app.Config.Port = 3000

// 2. 环境变量（如果代码未设置）
// export PORT=8080

// 3. 默认值（如果环境变量也未设置）
// 默认端口: 8080
```

## 配置最佳实践

### 1. 使用环境变量

```go
// 推荐：使用环境变量
// export PORT=8080
app := zoox.New()
app.Run()  // 自动从环境变量读取

// 不推荐：硬编码配置
app.Config.Port = 8080
```

### 2. 区分开发和生产环境

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

### 3. 使用配置文件

```go
import "gopkg.in/yaml.v3"

type AppConfig struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
}

func main() {
	data, _ := os.ReadFile("config.yaml")
	var cfg AppConfig
	yaml.Unmarshal(data, &cfg)
	
	app := zoox.New()
	app.Config.Port = cfg.Server.Port
	app.Run()
}
```

### 4. 密钥管理

```go
// 推荐：从环境变量读取密钥
// export SECRET_KEY=your-secret-key
app := zoox.New()

// 不推荐：硬编码密钥
app.Config.SecretKey = "hardcoded-secret"
```

## 生命周期钩子

### BeforeReady

在服务器启动前执行：

```go
app := zoox.New()

app.SetBeforeReady(func() {
	// 初始化数据库连接
	// 加载配置
	// 注册中间件
})
```

### BeforeDestroy

在服务器关闭前执行：

```go
app := zoox.New()

app.SetBeforeDestroy(func() {
	// 关闭数据库连接
	// 清理资源
})
```

**说明**: 生命周期钩子参考 `application.go:357-365`。

## 下一步

- 🛣️ 学习 [路由系统](routing.md)
- 🔌 查看 [中间件使用](middleware.md)
- 📝 了解 [Context API](context.md)

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
