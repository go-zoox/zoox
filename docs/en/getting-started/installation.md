# Installation Guide

This guide will help you install and configure the Zoox framework.

## Prerequisites

### Go Version Requirements

Zoox requires **Go 1.22.1 or higher**.

Check your Go version:

```bash
go version
```

If your version is lower than 1.22.1, please install or upgrade Go first.

### Installing Go

There are multiple ways to install and manage Go versions. We recommend using **GVM (Go Version Manager)** for version management.

#### Method 1: Using GVM (Recommended)

GVM is a powerful Go version management tool that makes it easy to install, switch, and manage multiple Go versions.

##### Installing GVM

```bash
# Install using curl
curl -o- https://raw.githubusercontent.com/zcorky/gvm/master/install | bash

# Or install using wget
wget -qO- https://raw.githubusercontent.com/zcorky/gvm/master/install | bash
```

After installation, reload your shell configuration:

```bash
# Reload environment variables
source ~/.bashrc  # Linux or macOS (bash)
# or
source ~/.zshrc   # macOS (zsh)
```

##### Installing Go with GVM

```bash
# View available Go versions
gvm ls-remote

# Install a specific version
gvm install go1.22.1

# Use the installed version
gvm use go1.22.1 --default
```

#### Method 2: Official Installation

Visit the [official Go website](https://golang.org/dl/) to download and install Go for your operating system.

## Installing Zoox

### Using go get

```bash
go get github.com/go-zoox/zoox
```

### Using go mod

If you're using Go modules (recommended for Go 1.11+):

```bash
# Initialize a new module (if not already done)
go mod init your-project-name

# Add Zoox as a dependency
go get github.com/go-zoox/zoox
```

## Verifying Installation

Create a simple test file to verify the installation:

```go
package main

import (
    "github.com/go-zoox/zoox"
)

func main() {
    app := zoox.New()
    app.Get("/", func(ctx *zoox.Context) {
        ctx.String(200, "Hello, Zoox!")
    })
    app.Run(":8080")
}
```

Run the test:

```bash
go run main.go
```

Visit `http://localhost:8080` in your browser. If you see "Hello, Zoox!", the installation is successful!

## Next Steps

- [Quick Start Guide](/en/getting-started/quick-start) - Get up and running in 5 minutes
- [First Application](/en/getting-started/first-app) - Build your first Zoox application
- [Examples](/en/getting-started/examples) - Explore common use cases
