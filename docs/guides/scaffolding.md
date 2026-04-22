# 业务脚手架 CLI

Zoox 在仓库内提供官方命令行工具，用于生成与 **go-idp / eunomia 风格**相近的分层布局：**api**、**services**、**models**，并附带 **migrate**、**config**、**middlewares**、**utils**、**router**。模板通过 `embed` 内嵌在二进制中。

## 安装

```bash
go install github.com/go-zoox/zoox/cmd/zoox@latest
```

确认 `$(go env GOPATH)/bin` 在 `PATH` 中后执行 `zoox --help`。

## 命令一览

| 命令 | 说明 |
|------|------|
| `zoox new` | 在空目录生成完整项目骨架（见下文目录树）。 |
| `zoox gen module <name>` | 在已有项目中生成 `api/v1/<name>`、`services/v1/<name>`、`models/v1/<name>`，并改写 `router/rest.go` 与 `models/register.go`。 |

当前生成器固定使用 **`v1`** 作为 URL 与目录版本段（与 `scaffold.DefaultAPIVersion` 一致）。

### `zoox new`

```text
zoox new [--module 模块路径] [--go 版本] <目录>
```

- **`--module` / `-m`**：`go.mod` 的 module 路径；目录名为 `.` 等无法推断时需显式指定。
- **`--go`**：写入 `go.mod` 的 `go` 行，默认 `1.22`。
- 依赖：`go.mod` 中带 `require github.com/go-zoox/zoox v1.18.2`（可按需 `go get` 升级）。

```bash
zoox new ./my-api --module github.com/acme/my-api
cd my-api
go mod tidy
go run ./cmd/server
```

默认 `PORT` 未设置时监听 `:8080`，`GET /health` 返回 JSON。

### `zoox gen module`

```text
zoox gen module [--dir 路径] <name>
```

- **`<name>`**：`[a-z][a-z0-9]*`（如 `user`）。
- 资源路径规则与此前一致：默认 `name + "s"`（已以 `s` 结尾则不变），例如 `user` → 挂载 `/api/v1/users`。

**标记行（不要删）：**

- `router/rest.go`：`// zoox:register-api-imports`、`// zoox:register-routes`（行前为 `\t`）
- `models/register.go`：`// zoox:register-model-imports`

`gen module` 会在上述标记**上方**追加 import 与 `r.Group("/api/v1/...", v1Xxx.Router())`，并在 `models/register.go` 中追加对 `models/v1/<name>` 的空白导入。

## `zoox new` 生成的目录结构

```text
<project>/
├── go.mod
├── cmd/server/main.go    # config.Load、migrate.Run、middlewares.Setup、router.Register
├── config/
│   ├── config.go
│   └── load.go           # 读取 PORT 等环境变量
├── migrate/
│   └── migrate.go        # 调用 models.Register()
├── models/
│   └── register.go       # 含 // zoox:register-model-imports
├── router/
│   └── rest.go           # 含 // zoox:register-api-imports、// zoox:register-routes
├── middlewares/
│   └── middlewares.go    # Recovery、RequestID；Health
└── utils/
    └── utils.go          # 小工具占位（如 Ptr）
```

## `zoox gen module` 追加的结构（示例 `user`）

```text
api/v1/user/
├── api.go        # 接口 + Get() 单例
├── router.go     # Router() -> zoox.RouterGroup 回调
└── impl.go       # List / Retrieve 示例

services/v1/user/
├── service.go    # 接口 + Get() 单例
└── impl.go       # 调用 model 层

models/v1/user/
├── model.go
├── dto.go
└── impl.go       # Store + 内存示例数据
```

同时 `router/rest.go` 会增加类似：

```text
r.Group("/api/v1/users", v1User.Router())
```

## 仓库内模板路径（源码）

```text
cmd/zoox/scaffold/templates/
├── new/
│   ├── go.mod.tmpl
│   ├── cmd/server/main.go.tmpl
│   ├── config/
│   ├── migrate/
│   ├── models/register.go.tmpl
│   ├── router/rest.go.tmpl
│   ├── middlewares/
│   └── utils/
└── module/
    ├── api.go.tmpl
    ├── router.go.tmpl
    ├── api_impl.go.tmpl
    ├── service.go.tmpl
    ├── service_impl.go.tmpl
    ├── model.go.tmpl
    ├── dto.go.tmpl
    └── model_impl.go.tmpl
```

## 模板变量（module 模板）

```text
{{.Module}}       — go.mod 的 module 路径
{{.APIVersion}}   — 当前固定 v1
{{.Name}}         — 命令行资源名（小写）
{{.Exported}}     — 导出前缀，如 user → User
{{.ResourcePath}} — URL 段，如 users
```

## 典型工作流

1. `zoox new ./app --module example.com/app`
2. `cd app && go mod tidy && go run ./cmd/server`
3. `zoox gen module order`
4. 访问 `GET /api/v1/orders`、`GET /api/v1/orders/:id`（示例 JSON）。

生成代码仅为起点；真实项目可接入 gormx、统一错误包、与 eunomia 一致的 `api.Api` 门面等，再逐步替换。

## 与文档示例的关系

[RESTful API 示例](/examples/rest-api) 仍以手写结构说明为主；CLI 提供的是**固定目录约定**与可重复生成的模块桩。
