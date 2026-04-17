# Agent 指引（Zoox）

面向在本仓库内改代码的自动化助手；与 HTTP 服务协议相关的实现经验如下。

## 文档

- 用户可见的配置与协议说明：[docs/guides/configuration.md](docs/guides/configuration.md)（含 HTTP/2、h2c、HTTP/3、环境变量）。

### 文档站点（VitePress / Node）

- **不要**在仓库根目录找 `package.json`：VitePress 与 pnpm 相关文件在 **`docs/`** 下，包括 `docs/package.json`、`docs/pnpm-lock.yaml`、`docs/tsconfig.json`。
- 本地开发：先 `cd docs`，再 `pnpm install`、`pnpm run docs:dev` / `docs:build` / `docs:preview`（脚本里 VitePress 根目录为当前目录 `.`）。
- CI： [.github/workflows/docs.yml](.github/workflows/docs.yml) 在 **`working-directory: docs`** 下执行 `pnpm install` 与 `pnpm run docs:build`；`actions/setup-node` 的 `cache-dependency-path` 为 **`docs/pnpm-lock.yaml`**。
- 人类可读说明：[README_DOCS.md](README_DOCS.md)。

## 服务入口与协议（实现位置）

- 启动与并发：`application.go` 中 `serve()` 使用 `errgroup` 并行跑 `serveHTTP`、`serveHTTPS`、`serveHTTP3`。
- TLS 构建：`buildTLSConfig()`；HTTPS 与 HTTP/3 通过闭包内 `sync.Once` 共享同一次构建结果，避免重复读盘与竞态。
- **HTTP/2（TLS）**：`serveHTTPS` 在 `http2.ConfigureServer` 之后 `tls.NewListener` + `Serve`。若 `tls.Config.NextProtos` 为空，`buildTLSConfig` 会设为 `h2`、`http/1.1`，否则在仅依赖 `tls.NewListener` 时客户端可能一直落到 HTTP/1.1。
- **h2c**：`serveHTTP` 在 `EnableH2C` 且 `NetworkType == "tcp"` 时用 `h2c.NewHandler`；不要在公网裸奔。
- **HTTP/3**：`serveHTTP3` 使用 `github.com/quic-go/quic-go/http3`，`http3.ConfigureTLSConfig` 从共享 TLS 派生 QUIC 用配置；需 `EnableHTTP3`、`HTTPSPort`、有效证书，且当前仅 **TCP** 网络（非 unix socket）。HTTPS 侧可通过 `wrapWithAltSvc` 发 `Alt-Svc`（`HTTP3AltSvcMaxAge < 0` 可关闭）。

## 配置字段与环境变量

- 见 [config/config.go](config/config.go)：`EnableH2C`、`EnableHTTP3`、`HTTP3Port`、`HTTP3AltSvcMaxAge`。
- 环境变量见 [constants.go](constants.go)：`ENABLE_H2C`、`ENABLE_HTTP3`、`HTTP3_PORT`、`HTTP3_ALTSVC_MAX_AGE`。

## 测试

- 协议相关测试：[application_protocol_test.go](application_protocol_test.go)（HTTP/2 over TLS、Alt-Svc 字符串、`Application` 作为 HTTP/3 handler）。

## 兼容性注意

- `responseWriter` 实现 `Hijack`；在 **HTTP/2** 连接上 Hijack 不可用（与 `net/http` 行为一致）。WebSocket/原始 TCP 场景需单独考虑协议与升级路径。

## 本地构建

- 若在父目录启用了 Go workspace 且与单模块开发冲突，可对仅针对本模块的命令使用 `GOWORK=off`（以本机环境为准）。
