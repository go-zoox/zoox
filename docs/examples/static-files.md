# 静态文件服务示例

这是一个静态文件服务示例，展示了如何使用 Zoox 提供静态文件服务。

## 项目结构

```
static-files/
├── main.go
├── go.mod
└── public/
    ├── index.html
    ├── css/
    │   └── style.css
    ├── js/
    │   └── app.js
    └── images/
        └── logo.png
```

## 完整代码

### main.go

```go
package main

import (
	"time"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

func main() {
	app := zoox.New()

	// 全局中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())
	
	// 静态文件缓存中间件
	app.Use(middleware.StaticCache(&middleware.StaticCacheConfig{
		MaxAge: 7 * 24 * time.Hour, // 7天缓存
	}))

	// 提供静态文件服务（基本用法）
	app.Static("/static", "./public")

	// 提供静态文件服务（高级配置）
	app.Static("/assets", "./public", &zoox.StaticOptions{
		Gzip:         true,              // 启用 Gzip 压缩
		CacheControl: "public, max-age=3600", // 缓存控制
		MaxAge:       1 * time.Hour,    // 最大缓存时间
		Index:        true,             // 支持 index.html
	})

	// HTML 页面
	app.Get("/", func(ctx *zoox.Context) {
		ctx.RenderHTML("./public/index.html")
	})

	// API 路由
	api := app.Group("/api")
	api.Get("/info", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"message": "Static file server is running",
			"static_path": "/static",
			"assets_path": "/assets",
		})
	})

	app.Run(":8080")
}
```

### public/index.html

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>静态文件服务示例</title>
	<link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
	<div class="container">
		<h1>静态文件服务示例</h1>
		<img src="/static/images/logo.png" alt="Logo">
		<p>这是一个使用 Zoox 提供的静态文件服务示例。</p>
		<button onclick="loadInfo()">获取 API 信息</button>
		<div id="info"></div>
	</div>
	<script src="/static/js/app.js"></script>
</body>
</html>
```

### public/css/style.css

```css
body {
	font-family: Arial, sans-serif;
	margin: 0;
	padding: 20px;
	background-color: #f5f5f5;
}

.container {
	max-width: 800px;
	margin: 0 auto;
	background: white;
	padding: 30px;
	border-radius: 8px;
	box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

h1 {
	color: #333;
}

img {
	max-width: 200px;
	height: auto;
	margin: 20px 0;
}

button {
	background-color: #007bff;
	color: white;
	border: none;
	padding: 10px 20px;
	border-radius: 4px;
	cursor: pointer;
	font-size: 16px;
}

button:hover {
	background-color: #0056b3;
}

#info {
	margin-top: 20px;
	padding: 15px;
	background-color: #f8f9fa;
	border-radius: 4px;
}
```

### public/js/app.js

```javascript
async function loadInfo() {
	try {
		const response = await fetch('/api/info');
		const data = await response.json();
		document.getElementById('info').innerHTML = `
			<h3>服务器信息</h3>
			<pre>${JSON.stringify(data, null, 2)}</pre>
		`;
	} catch (error) {
		document.getElementById('info').innerHTML = `
			<p style="color: red;">错误: ${error.message}</p>
		`;
	}
}
```

## 功能说明

### 1. 基本静态文件服务

```go
app.Static("/static", "./public")
```

- 访问 `/static/css/style.css` 会返回 `./public/css/style.css`
- 支持所有常见的文件类型（HTML, CSS, JS, 图片等）

### 2. 高级静态文件配置

```go
app.Static("/assets", "./public", &zoox.StaticOptions{
	Gzip:         true,
	CacheControl: "public, max-age=3600",
	MaxAge:       1 * time.Hour,
	Index:        true,
})
```

**选项说明**：
- `Gzip`: 启用 Gzip 压缩，减少传输大小
- `CacheControl`: 设置缓存控制头
- `MaxAge`: 设置最大缓存时间
- `Index`: 支持目录索引（访问目录时返回 index.html）

### 3. 静态文件缓存中间件

```go
app.Use(middleware.StaticCache(&middleware.StaticCacheConfig{
	MaxAge: 7 * 24 * time.Hour,
}))
```

自动为静态文件添加缓存控制头。

### 4. 自定义文件系统

```go
import "net/http"

app.StaticFS("/static", http.Dir("./public"))
```

使用 Go 标准库的 `http.FileSystem` 接口。

## 使用场景

1. **前端应用部署**: 部署 React、Vue、Angular 等前端应用
2. **CDN 替代**: 作为本地 CDN 提供静态资源服务
3. **文件托管**: 托管文档、图片、视频等静态资源
4. **SPA 应用**: 单页应用的静态文件服务

## 测试

### 1. 启动服务器

```bash
go run main.go
```

### 2. 访问静态文件

```bash
# HTML 页面
curl http://localhost:8080/

# CSS 文件
curl http://localhost:8080/static/css/style.css

# JavaScript 文件
curl http://localhost:8080/static/js/app.js

# 图片文件
curl http://localhost:8080/static/images/logo.png
```

### 3. 访问 API

```bash
curl http://localhost:8080/api/info
```

## 注意事项

1. **路径安全**: 确保静态文件目录不会被直接访问敏感文件
2. **文件大小**: 大文件建议使用 CDN 或专门的静态文件服务
3. **缓存策略**: 合理设置缓存时间，平衡性能和更新频率
4. **MIME 类型**: Zoox 会自动识别常见文件类型的 MIME 类型

## 下一步

- 📡 查看 [RESTful API 示例](rest-api.md) - REST API 开发
- 🔌 查看 [WebSocket 应用示例](real-time-app.md) - WebSocket 应用
- 🏗️ 学习 [微服务示例](microservice.md) - 微服务架构
- 📚 阅读 [最佳实践](../best-practices.md) - 开发建议
