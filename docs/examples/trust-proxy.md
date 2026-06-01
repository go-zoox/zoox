# Trust Proxy 示例

这个示例展示如何在 Zoox 作为中间代理时保留上游 `X-Forwarded-*`。

## 场景

```text
Client (HTTPS)
  -> Ingress/Nginx (设置 X-Forwarded-Proto/Port)
  -> Zoox
  -> Backend
```

## 代码

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	app := zoox.New()

	// 一键开启 trust proxy（全局）
	app.Config.TrustProxy = true

	app.Proxy("/api", "http://backend:8080")

	app.Run(":8080")
}
```

## 路由级配置

```go
app.Proxy("/api", "http://backend:8080", func(cfg *zoox.ProxyConfig) {
	cfg.TrustProxy = true
})
```

## 验证建议

后端回显请求头时，确认以下值与上游一致：

- `X-Forwarded-Proto`
- `X-Forwarded-Port`
- `X-Forwarded-Host`
