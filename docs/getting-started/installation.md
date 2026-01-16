# 安装指南

本文档将指导你如何安装和配置 Zoox 框架。

## 前置要求

### Go 版本要求

Zoox 需要 **Go 1.22.1 或更高版本**。

检查你的 Go 版本：

```bash
go version
```

如果版本低于 1.22.1，请先升级 Go：
- 访问 [Go 官网](https://golang.org/dl/) 下载最新版本
- 或使用包管理器升级（如 `brew upgrade go` on macOS）

### 环境变量

确保 `GOPATH` 和 `GOROOT` 已正确配置：

```bash
go env GOPATH
go env GOROOT
```

## 安装步骤

### 1. 创建项目（如果还没有）

```bash
mkdir my-zoox-app
cd my-zoox-app
go mod init my-zoox-app
```

### 2. 安装 Zoox

使用 `go get` 命令安装 Zoox：

```bash
go get github.com/go-zoox/zoox
```

这将自动下载 Zoox 及其所有依赖项。

### 3. 验证安装

创建一个简单的测试文件 `main.go`：

```go
package main

import (
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()
	
	app.Get("/", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "Zoox installed successfully!",
		})
	})
	
	app.Run(":8080")
}
```

运行应用：

```bash
go run main.go
```

如果看到类似以下的输出，说明安装成功：

```
  ____               
 /_  / ___  ___ __ __
  / /_/ _ \/ _ \\ \ /
 /___/\___/\___/_\_\  v1.16.6

Lightweight, high performance Go web framework

https://github.com/go-zoox/zoox
____________________________________O/_______
                                    O\

[router] register:      GET /
Server started at http://127.0.0.1:8080
```

在另一个终端测试：

```bash
curl http://localhost:8080
```

应该返回：

```json
{"message":"Zoox installed successfully!"}
```

## 常见问题

### 问题 1: `go: cannot find module providing package`

**原因**: Go 模块未正确初始化或网络问题。

**解决方案**:
1. 确保已运行 `go mod init`
2. 检查网络连接
3. 设置 Go 代理（如果需要）：
   ```bash
   go env -w GOPROXY=https://goproxy.cn,direct  # 中国用户
   ```

### 问题 2: `go: requires go >= 1.22.1`

**原因**: Go 版本过低。

**解决方案**: 升级 Go 到 1.22.1 或更高版本。

### 问题 3: 依赖下载失败

**原因**: 网络问题或代理配置。

**解决方案**:
1. 检查网络连接
2. 配置 Go 代理：
   ```bash
   go env -w GOPROXY=https://goproxy.cn,direct
   ```
3. 清理模块缓存：
   ```bash
   go clean -modcache
   go mod download
   ```

### 问题 4: 端口已被占用

**原因**: 8080 端口已被其他程序使用。

**解决方案**:
1. 更改端口：
   ```go
   app.Run(":3000")  // 使用其他端口
   ```
2. 或停止占用端口的程序

## 下一步

安装成功后，你可以：

1. 📖 阅读 [5分钟快速开始](quick-start.md) 了解基本用法
2. 🎯 查看 [第一个应用教程](first-app.md) 学习完整开发流程
3. 💡 浏览 [常见场景示例](examples.md) 查看实际应用

---

**准备好了吗？** 👉 [5分钟快速开始](quick-start.md)
