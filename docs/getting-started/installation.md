# 安装指南

本文档将指导你如何安装和配置 Zoox 框架。

## 前置要求

### Go 版本要求

Zoox 需要 **Go 1.22.1 或更高版本**。

检查你的 Go 版本：

```bash
go version
```

如果版本低于 1.22.1，请先安装或升级 Go。

### 安装 Go

有多种方式可以安装和管理 Go 版本，推荐使用 **GVM (Go Version Manager)** 进行版本管理。

#### 方法 1: 使用 GVM（推荐）

GVM 是一个强大的 Go 版本管理工具，可以轻松安装、切换和管理多个 Go 版本。

##### 安装 GVM

```bash
# 使用 curl 安装
curl -o- https://raw.githubusercontent.com/zcorky/gvm/master/install | bash

# 或使用 wget 安装
wget -qO- https://raw.githubusercontent.com/zcorky/gvm/master/install | bash
```

安装完成后，重新加载 shell 配置：

```bash
# 重新加载环境变量
source ~/.bashrc  # Linux 或 macOS (bash)
# 或
source ~/.zshrc   # macOS (zsh)
```

##### 使用 GVM 安装 Go

```bash
# 查看可用的 Go 版本
gvm ls-remote

# 安装指定版本（推荐安装 1.22.1 或更高版本）
gvm install 1.22.1

# 使用指定的 Go 版本
gvm use 1.22.1

# 查看当前使用的 Go 版本
gvm current

# 查看已安装的所有版本
gvm ls
```

##### 项目级版本管理

在项目目录中创建 `.gvmrc` 文件，指定该项目使用的 Go 版本：

```bash
# 在项目根目录创建 .gvmrc 文件
echo "1.22.1" > .gvmrc

# 进入项目时，GVM 会自动切换到指定版本
cd my-zoox-app
gvm use  # 自动使用 .gvmrc 中指定的版本
```

##### 其他 GVM 常用命令

```bash
# 移除指定版本
gvm remove 1.21.0

# 使用指定版本执行命令
gvm exec 1.22.1 go version

# 在指定版本的新 shell 中执行
gvm shell 1.22.1
```

##### GVM 故障排除

如果遇到 `gvm: command not found` 错误：

1. **重新加载环境变量**：
   ```bash
   source ~/.bashrc  # 或 source ~/.zshrc
   ```

2. **重启终端**

3. **重新注册 GVM**：
   ```bash
   # 如果使用 zmicro 工具集
   zmicro register
   # 然后重新加载环境变量
   ```

更多信息请参考 [GVM 官方仓库](https://github.com/zcorky/gvm)。

#### 方法 2: 官方安装包

访问 [Go 官网](https://golang.org/dl/) 下载对应平台的安装包，按照官方指南进行安装。

#### 方法 3: 包管理器

```bash
# macOS (Homebrew)
brew install go

# Linux (apt)
sudo apt update
sudo apt install golang-go

# Linux (yum)
sudo yum install golang
```

### 环境变量

确保 `GOPATH` 和 `GOROOT` 已正确配置：

```bash
go env GOPATH
go env GOROOT
```

使用 GVM 时，这些环境变量会自动设置，无需手动配置。

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

**解决方案**: 
1. 如果使用 GVM，安装并使用新版本：
   ```bash
   gvm install 1.22.1
   gvm use 1.22.1
   ```
2. 如果使用官方安装包，从 [Go 官网](https://golang.org/dl/) 下载并安装新版本
3. 如果使用包管理器，升级到最新版本：
   ```bash
   # macOS
   brew upgrade go
   
   # Linux (apt)
   sudo apt upgrade golang-go
   ```

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
