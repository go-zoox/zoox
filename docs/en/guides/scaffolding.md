# Scaffolding CLI

The official `zoox` tool creates a **new project** (`new`), scaffolds **domain modules** in an existing repo (`gen module`), and ships companion commands — **`install`**, **`dev`**, **`build`** — for day-to-day work. Human-readable metadata and overridable defaults live in **`.zoox/config.yaml`** at the project root (see below). Templates are embedded in the binary; the **Go module path** stays in **`go.mod` only** (not duplicated in `.zoox/config.yaml`).

## Install

```bash
go install github.com/go-zoox/zoox/cmd/zoox@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`, then run `zoox --help`.

## Command overview

| Command | Purpose |
|--------|---------|
| `zoox new` | Create a full skeleton, including `.zoox/config.yaml`. |
| `zoox gen module <name>` | Add `api/v1/<name>`, `services/v1/<name>`, `models/v1/<name>` and patch `router/rest.go` + `models/register.go`. |
| `zoox install` | Run `go mod tidy` in the project root (default: cwd, use `-C` to point at the root). |
| `zoox dev` | Run `go mod tidy`, then **watch, rebuild, and run** the main package (default `./cmd/server`) on the **host** platform. |
| `zoox build` | Run `go mod tidy`, then `go build` the main package (default output `./bin/server`). |

`gen module` currently fixes the API segment to **`v1`** (`scaffold.DefaultAPIVersion`).

## `.zoox/config.yaml`

`zoox new` writes this file at the project root. Keys include:

| Key | Description |
|-----|-------------|
| `version` | File schema version; currently `1`. |
| `name` | Human-readable project name. |
| `author` | Optional author string. |
| `description` | Optional project description. |
| `created_at` | UTC RFC3339 timestamp, set by `zoox new`. |
| `cli` | The **zoox CLI version** that ran `zoox new`. |
| `app.entry` | Main package path, default `./cmd/server`. |
| `build.output` | Binary output path, default `./bin/server`. |
| `dev.ignores` | Optional extra ignore paths for the watcher (list). |

Example:

```yaml
version: 1
name: my-api
author: team@example.com
description: Order and inventory demo
created_at: "2026-01-15T10:00:00Z"
cli: "1.0.0"
app:
  entry: ./cmd/server
build:
  output: ./bin/server
dev:
  ignores: []
```

## `zoox new`

```text
zoox new [--module path] [--go version] [--name] [--author] [--description] <dir>
```

- `--module` / `-m`: the `go.mod` module path.  
- `--go`: `go` version line in `go.mod` (default `1.22`).  
- `--name`, `--author`, `--description`: written to `.zoox/config.yaml` (`name` defaults to the directory name).  
- The template pins a `github.com/go-zoox/zoox` `require` line; upgrade with `go get` as needed.

**Minimal example**

```bash
mkdir -p /tmp/zoox-demo && zoox new /tmp/zoox-demo --module github.com/acme/zoox-demo
```

**With metadata**

```bash
zoox new ./my-api \
  --module github.com/acme/my-api \
  --name "Acme API" \
  --author "you@example.com" \
  --description "Internal REST edge"
cd my-api
zoox install
zoox dev
```

By default, without `PORT`, the app listens on `:8080` and `GET /health` returns JSON.

## `zoox install` / `dev` / `build`

- Use **`-C` / `--context`** to set the project root (the directory that contains `go.mod`). Default: current directory.  
- **`dev`** is for local edit → rebuild → run on the **current machine**.  
- **`build`**: use `-e` / `--entry`, `-o` / `--output` to override paths (see `zoox build --help`); relative output paths are resolved against the project root.  

```bash
cd my-api
zoox install
zoox build
./bin/server
```

From another path:

```bash
zoox dev -C /path/to/my-api
zoox build -C /path/to/my-api -o dist/server
```

## `zoox gen module`

```text
zoox gen module [--dir path] <name>
```

- `<name>` must match `[a-z][a-z0-9]*` (e.g. `user`).  
- `--dir` / `-C`: walk **up** from that path to find `go.mod`.  
- **Marker lines** (leading tab) must remain in:  
  - `router/rest.go`: `// zoox:register-api-imports`, `// zoox:register-routes`  
  - `models/register.go`: `// zoox:register-model-imports`  

## Layout after `zoox new`

```text
<project>/
├── .zoox/config.yaml
├── go.mod
├── cmd/server/main.go
├── config/
├── migrate/
├── models/register.go
├── router/rest.go
├── middlewares/
└── utils/
```

## After `zoox gen module user`

```text
api/v1/user/
services/v1/user/
models/v1/user/
```

Example routes: `GET /api/v1/users`, `GET /api/v1/users/:id`.

## Examples (copy-paste)

### 1) From zero to a running HTTP server

```bash
zoox new ./demo-svc --module example.com/demo
cd demo-svc
zoox install
zoox build
PORT=9000 ./bin/server
curl -sS http://127.0.0.1:9000/health
```

### 2) Hot reload (watch + build + run)

```bash
cd demo-svc
zoox dev
```

### 3) Add an `order` module and call the sample API

```bash
cd demo-svc
zoox gen module order
zoox build
./bin/server
curl -sS http://127.0.0.1:8080/api/v1/orders
curl -sS http://127.0.0.1:8080/api/v1/orders/123
```

### 4) `gen module` from a nested directory

```bash
cd demo-svc/some/nested/folder
zoox gen module payment
# or
zoox gen module -C /path/to/demo-svc payment
```

### 5) Commit metadata for your team

```bash
zoox new ./billing \
  --module github.com/team/billing \
  --name "Billing" \
  --author "Platform Team" \
  --description "Usage and billing"
# Commit .zoox/config.yaml (with go.mod) for the `name` / `author` / `description` / `cli` record.
```

## Template sources in the repository

`cmd/zoox/scaffold/templates/` — `new/` and `module/` in the [Zoox](https://github.com/go-zoox/zoox) repo.

## Module template variables

```text
{{.Module}}       — go.mod module path
{{.APIVersion}}   — fixed to v1 today
{{.Name}}         — resource name (lowercase)
{{.Exported}}     — e.g. user → User
{{.ResourcePath}} — URL segment, e.g. users
```

## Related docs

- Chinese: [Scaffolding CLI](/guides/scaffolding) (zh-CN)  
- Hand-written walkthrough: [RESTful API example](/en/examples/rest-api) — the CLI is the **opinionated layout** + repeatable stubs.
