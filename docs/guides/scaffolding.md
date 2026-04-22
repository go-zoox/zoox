# 业务脚手架 CLI

Zoox 在仓库内提供官方命令行工具 `zoox`：在空目录**创建项目**（`new`）、在已有项目中**生成领域模块**（`gen module`），并配套 **`install` / `dev` / `build`** 等本地开发命令。模板通过 `embed` 内嵌在二进制中；项目元数据与可覆盖的默认行为写在 **`.zoox/config.yaml`**（见下文）。

## 安装

```bash
go install github.com/go-zoox/zoox/cmd/zoox@latest
```

确认 `$(go env GOPATH)/bin` 在 `PATH` 中后执行 `zoox --help`。

## 命令一览

| 命令 | 说明 |
|------|------|
| `zoox new` | 在空目录生成完整项目骨架（含 `.zoox/config.yaml`）。 |
| `zoox gen module <name>` | 在已有项目中生成 `api/v1/<name>`、`services/v1/<name>`、`models/v1/<name>`，并改写 `router/rest.go` 与 `models/register.go`。 |
| `zoox install` | 在指定项目根目录执行 `go mod tidy`（默认当前目录，可用 `-C` 指定根目录）。 |
| `zoox dev` | 先 `go mod tidy`，再监听文件变化并重新 **本机构建 + 运行** 主包（默认 `./cmd/server`）。 |
| `zoox build` | 先 `go mod tidy`，再对主包执行 `go build`（默认可执行文件为 `./bin/server`）。 |
| `zoox database migrate` | 先 `go mod tidy`，再 `go run ./cmd/migrate`（只执行 `migrate.Run()`，不启动 HTTP）。 |

**说明**：`zoox install`、`zoox dev`、`zoox build`、`zoox database migrate` 仅在**项目根存在**由 `zoox new` 生成的 **`.zoox/config.yaml`**（且 `version >= 1`）时可用；普通 Go 仓库或手抄结构但无该文件时，上述命令会报错「不是 Zoox 项目」类提示。`zoox gen module` 仍只需能定位到 `go.mod` 且保留路由/模型中的标记行。

当前 `gen module` 固定使用 **`v1`** 作为 URL 与目录版本段（与 `scaffold.DefaultAPIVersion` 一致）。

## `.zoox/config.yaml`

`zoox new` 会在**项目根**创建 `.zoox/config.yaml`。**Go 的 module 路径只写在 `go.mod`**，本文件不存 `module` 字段。

| 键 | 说明 |
|----|------|
| `version` | 文件格式版本，当前为 `1`。 |
| `name` | 项目展示名。 |
| `author` | 作者（可为空）。 |
| `description` | 项目描述（可为空）。 |
| `created_at` | 创建时间（UTC，RFC3339），由 `zoox new` 自动写入。 |
| `cli` | 执行 `zoox new` 时使用的 **zoox CLI 版本号**。 |
| `app.entry` | 主包路径，缺省为 `./cmd/server`（可手工修改；未来可与 `dev`/`build` 对齐读取）。 |
| `build.output` | 可执行文件输出路径，缺省为 `./bin/server`。 |
| `dev.ignores` | 开发监听时额外忽略的路径（列表），可为空。 |

新建项目后文件形状示例：

```yaml
version: 1
name: my-api
author: team@example.com
description: 订单与库存演示服务
created_at: "2026-01-15T10:00:00Z"
cli: "1.0.0"
app:
  entry: ./cmd/server
build:
  output: ./bin/server
dev:
  ignores: []
```

### `zoox new`

```text
zoox new [--module 模块路径] [--go 版本] [--name 名称] [--author 作者] [--description 描述] <目录>
```

- **`--module` / `-m`**：`go.mod` 的 module 路径；目录名为 `.` 等无法推断时需显式指定。  
- **`--go`**：写入 `go.mod` 的 `go` 行，默认 `1.22`。  
- **`--name`**：写入 `.zoox/config.yaml` 的 `name`；省略则用 `<目录>` 的目录名。  
- **`--author`**、**`--description`**：写入同文件顶层的 `author`、`description`。  
- 依赖：`go.mod` 中由模板生成 `require github.com/go-zoox/zoox`（版本以模板为准，可按需 `go get` 升级）。

**最小示例：**

```bash
mkdir -p /tmp/zoox-demo && zoox new /tmp/zoox-demo --module github.com/acme/zoox-demo
```

**带元数据示例：**

```bash
zoox new ./my-api \
  --module github.com/acme/my-api \
  --name "Acme API" \
  --author "you@example.com" \
  --description "内部 REST 网关"
cd my-api
zoox install
zoox dev
```

默认未设置 `PORT` 时监听 `:8080`，`GET /health` 返回 JSON。

### `zoox install` / `zoox dev` / `zoox build`

- 三者均可使用 **`-C` / `--context`** 指定**项目根目录**（含 `go.mod`）；不指定则为**当前工作目录**。  
- **`dev`**：适合本地改代码后自动重编、重启；仅针对**本机**平台。  
- **`build`**：生成可执行文件；`--output` / `-o`、`--entry` / `-e` 可覆盖路径（含义见 `zoox build --help`），相对 `-C` 目录解析。  

```bash
cd my-api
zoox install
zoox build
./bin/server
```

在**另一 shell** 或从**子目录**对项目根执行时：

```bash
zoox dev -C /path/to/my-api
# 或
zoox build -C /path/to/my-api -o dist/server
```

### `zoox database migrate`

在**项目根**执行 `go mod tidy`，再 `go run ./cmd/migrate`（入口会 `config.Load()` 并调用 `migrate/migrate.go` 里的 `Run()`），**不启动 HTTP**。需由 `zoox new` 生成或自行提供 `cmd/migrate/main.go`。

```bash
cd my-api
zoox database migrate
zoox database migrate -C /path/to/my-api
```

## `zoox gen module`

```text
zoox gen module [--dir 路径] <name>
```

- **`<name>`**：`[a-z][a-z0-9]*`（如 `user`）。  
- **`--dir` / `-C`**：从该路径**向上**查找 `go.mod`；不指定则当前目录。  
- 资源路径规则：默认 `name + "s"`（已以 `s` 结尾则不变），例如 `user` → 挂载 `/api/v1/users`。

**标记行（不要删）：**

- `router/rest.go`：`// zoox:register-api-imports`、`// zoox:register-routes`（行前为 `\t`）  
- `models/register.go`：`// zoox:register-model-imports`  

`gen module` 会在上述标记**上方**追加 import 与 `r.Group("/api/v1/...", v1Xxx.Router())`，并在 `models/register.go` 中追加对 `models/v1/<name>` 的空白导入。

## `zoox new` 生成的目录结构

```text
<project>/
├── .zoox/
│   └── config.yaml         # 元数据 + app/build/dev 默认
├── go.mod
├── cmd/
│   ├── migrate/main.go     # 仅 config + migrate.Run；`zoox database migrate` 使用
│   └── server/main.go      # config.Load、migrate.Run、middlewares.Setup、router.Register
├── config/
│   ├── config.go
│   └── load.go            # 读取 PORT 等环境变量
├── migrate/
│   └── migrate.go         # 调用 models.Register()
├── models/
│   └── register.go        # 含 // zoox:register-model-imports
├── router/
│   └── rest.go            # 含 // zoox:register-api-imports、// zoox:register-routes
├── middlewares/
│   └── middlewares.go     # Recovery、RequestID；Health
└── utils/
    └── utils.go           # 小工具占位
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

## 案例

### 案例 1：从零到可访问的 HTTP

```bash
zoox new ./demo-svc --module example.com/demo
cd demo-svc
zoox install
zoox build
PORT=9000 ./bin/server
# 另开终端
curl -sS http://127.0.0.1:9000/health
```

### 案例 2：开发模式（改代码即重编、重启）

```bash
cd demo-svc
zoox dev
# 修改 api 或 router 后保存，观察进程重启；默认仍监听 8080（或你在 config 中约定的端口）
```

### 案例 3：增加 `order` 模块并试 API

```bash
cd demo-svc
zoox gen module order
zoox build
./bin/server
curl -sS http://127.0.0.1:8080/api/v1/orders
curl -sS http://127.0.0.1:8080/api/v1/orders/123
```

### 案例 4：在仓库子目录中生成模块

```bash
cd demo-svc/some/nested/folder
zoox gen module payment
# 向上找到 go.mod 后，在 demo-svc 下生成 api/services/models
```

等价的显式方式：

```bash
zoox gen module -C /path/to/demo-svc payment
```

### 案例 5：用 `zoox new` 元数据与后续协作

```bash
zoox new ./billing \
  --module github.com/team/billing \
  --name "Billing" \
  --author "Platform Team" \
  --description "计量与账单服务"
# 将 .zoox/config.yaml 纳入版本库，供团队查看项目说明与当时 CLI 版本（cli 字段）
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
    └── …
```

## 模板变量（`gen module` 使用的 module 模板）

```text
{{.Module}}       — go.mod 的 module 路径
{{.APIVersion}}   — 当前固定 v1
{{.Name}}         — 命令行资源名（小写）
{{.Exported}}     — 导出前缀，如 user → User
{{.ResourcePath}} — URL 段，如 users
```

## 与文档示例的关系

[RESTful API 示例](/examples/rest-api) 以手写结构说明为主；本页的 CLI 提供**固定目录约定**与可重复生成的模块桩。将 `.zoox/config.yaml` 与 `go.mod` 一并纳入版本管理，便于在团队协作中保持「项目从哪一版 `zoox` 脚手架而来」的可追溯性。
