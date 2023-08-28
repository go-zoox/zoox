# Zoox - A Lightweight Web Framework

[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-zoox/zoox)](https://pkg.go.dev/github.com/go-zoox/zoox)
[![Build Status](https://github.com/go-zoox/zoox/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/go-zoox/zoox/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-zoox/zoox)](https://goreportcard.com/report/github.com/go-zoox/zoox)
[![Coverage Status](https://coveralls.io/repos/github/go-zoox/zoox/badge.svg?branch=master)](https://coveralls.io/github/go-zoox/zoox?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/go-zoox/zoox.svg)](https://github.com/go-zoox/zoox/issues)
[![Release](https://img.shields.io/github/tag/go-zoox/zoox.svg?label=Release)](https://github.com/go-zoox/zoox/tags)

## Installation
To install the package, run:

```bash
go get github.com/go-zoox/zoox
```

## Getting Started

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	app := zoox.Default()

	app.Get("/", func(ctx *zoox.Context) {
		ctx.Write([]byte("helloworld"))
	})

	app.Run(":8080")
}
```

## DevTools

```bash
# install
go install github.com/go-zoox/zoox/cmd/zoox@latest
```

```bash
# dev
zoox dev
```

```bash
# build
zoox build
```

```bash

## License
GoZoox is released under the [MIT License](./LICENSE).
