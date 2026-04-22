# Scaffolding CLI

The official `zoox` binary scaffolds a layout similar to **go-idp / eunomia**: **api**, **services**, **models**, plus **migrate**, **config**, **middlewares**, **utils**, and **router**. Templates are embedded in the tool.

## Install

```bash
go install github.com/go-zoox/zoox/cmd/zoox@latest
```

## Commands

| Command | Purpose |
|--------|---------|
| `zoox new` | Create a full skeleton (see tree below). |
| `zoox gen module <name>` | Add `api/v1/<name>`, `services/v1/<name>`, `models/v1/<name>` and patch `router/rest.go` + `models/register.go`. |

The generator currently fixes the API segment to **`v1`** (`scaffold.DefaultAPIVersion`).

### `zoox new`

```text
zoox new [--module path] [--go version] <dir>
```

The generated `go.mod` includes `require github.com/go-zoox/zoox v1.18.2` (upgrade with `go get` as needed).

### `zoox gen module`

Requires marker lines (leading tab) in:

- `router/rest.go`: `// zoox:register-api-imports`, `// zoox:register-routes`
- `models/register.go`: `// zoox:register-model-imports`

## Layout after `zoox new`

```text
<project>/
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

Example route: `GET /api/v1/users`, `GET /api/v1/users/:id`.

## Template sources

Under `cmd/zoox/scaffold/templates/` — see `new/` and `module/` in the Zoox repository.

## Locale

The Chinese guide is at `/guides/scaffolding`.
