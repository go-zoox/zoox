# 模板引擎

Zoox 提供了内置的模板引擎支持，可以轻松渲染 HTML 模板。

## 基本用法

### 设置模板目录

```go
package main

import "github.com/go-zoox/zoox"

func main() {
	app := zoox.New()
	
	// 设置模板目录
	app.SetTemplates("./templates/*")
	
	app.Run(":8080")
}
```

**说明**: 模板设置参考 `application.go:343-350`。

### 渲染模板

```go
app.Get("/", func(ctx *zoox.Context) {
	// 渲染模板
	ctx.Render(200, "index.html", zoox.H{
		"title": "Home Page",
		"name":  "Zoox",
	})
})
```

**说明**: Render 方法参考 `context.go:434-440`。

## 模板文件结构

```
project/
├── main.go
└── templates/
    ├── index.html
    ├── about.html
    └── layout.html
```

### 示例模板文件

`templates/index.html`:

```html
<!DOCTYPE html>
<html>
<head>
	<title>{{.title}}</title>
</head>
<body>
	<h1>Welcome to {{.name}}</h1>
	<p>This is the home page.</p>
</body>
</html>
```

## 自定义模板函数

### 注册模板函数

```go
app.SetTemplates("./templates/*", template.FuncMap{
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
})
```

### 在模板中使用

```html
<h1>{{upper .title}}</h1>
<p>{{formatDate .date}}</p>
```

## 直接渲染 HTML

### 渲染 HTML 字符串

```go
app.Get("/", func(ctx *zoox.Context) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Home</title>
	</head>
	<body>
		<h1>Hello, Zoox!</h1>
	</body>
	</html>
	`
	ctx.HTML(200, html)
})
```

**说明**: HTML 方法参考 `context.go:423-432`。

### 带数据的 HTML

```go
app.Get("/", func(ctx *zoox.Context) {
	html := `<h1>Hello, {{.name}}!</h1>`
	ctx.HTML(200, html, zoox.H{
		"name": "Zoox",
	})
})
```

## 模板继承

### 定义基础模板

`templates/layout.html`:

```html
<!DOCTYPE html>
<html>
<head>
	<title>{{.title}}</title>
</head>
<body>
	<header>
		<nav>Navigation</nav>
	</header>
	<main>
		{{block "content" .}}{{end}}
	</main>
	<footer>
		<p>Footer</p>
	</footer>
</body>
</html>
```

### 继承模板

`templates/index.html`:

```html
{{template "layout.html" .}}

{{define "content"}}
	<h1>Home Page</h1>
	<p>Welcome to Zoox!</p>
{{end}}
```

## 完整示例

```go
package main

import (
	"html/template"
	"strings"
	"time"
	
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()
	
	// 设置模板目录和自定义函数
	app.SetTemplates("./templates/*", template.FuncMap{
		"upper": strings.ToUpper,
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"add": func(a, b int) int {
			return a + b
		},
	})
	
	// 渲染模板
	app.Get("/", func(ctx *zoox.Context) {
		ctx.Render(200, "index.html", zoox.H{
			"title": "Home Page",
			"name":  "Zoox",
			"date":  time.Now(),
		})
	})
	
	// 直接渲染 HTML
	app.Get("/about", func(ctx *zoox.Context) {
		html := `
		<!DOCTYPE html>
		<html>
		<head><title>About</title></head>
		<body>
			<h1>About Us</h1>
			<p>This is the about page.</p>
		</body>
		</html>
		`
		ctx.HTML(200, html)
	})
	
	app.Run(":8080")
}
```

## 模板缓存

Zoox 会自动缓存编译后的模板，提高性能。

### 禁用缓存（开发环境）

在开发环境中，可以禁用缓存以便实时看到模板更改：

```go
// 每次请求重新加载模板（仅开发环境）
if !app.IsProd() {
	// 禁用缓存逻辑
}
```

## 常用模板函数

### 字符串函数

```go
app.SetTemplates("./templates/*", template.FuncMap{
	"upper":   strings.ToUpper,
	"lower":   strings.ToLower,
	"title":   strings.Title,
	"trim":    strings.TrimSpace,
	"replace": strings.ReplaceAll,
})
```

### 日期时间函数

```go
app.SetTemplates("./templates/*", template.FuncMap{
	"formatDate": func(t time.Time, layout string) string {
		return t.Format(layout)
	},
	"now": func() time.Time {
		return time.Now()
	},
})
```

### 数学函数

```go
app.SetTemplates("./templates/*", template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"multiply": func(a, b int) int {
		return a * b
	},
})
```

## 最佳实践

### 1. 组织模板文件

```
templates/
├── layouts/
│   └── base.html
├── pages/
│   ├── index.html
│   └── about.html
└── partials/
    ├── header.html
    └── footer.html
```

### 2. 使用模板继承

```html
{{template "layouts/base.html" .}}

{{define "content"}}
	<!-- 页面内容 -->
{{end}}
```

### 3. 转义用户输入

模板引擎会自动转义 HTML，防止 XSS 攻击：

```html
<!-- 自动转义 -->
<p>{{.userInput}}</p>

<!-- 如果需要原始 HTML（谨慎使用） -->
<p>{{.htmlContent | safeHTML}}</p>
```

### 4. 缓存静态 HTML

对于不经常变化的页面，可以缓存渲染结果：

```go
app.Get("/", func(ctx *zoox.Context) {
	cacheKey := "page:index"
	var html string
	
	if ctx.Cache().Get(cacheKey, &html) != nil {
		// 缓存未命中，渲染模板
		// ... 渲染逻辑
		ctx.Cache().Set(cacheKey, html, time.Hour)
	}
	
	ctx.HTML(200, html)
})
```

## 下一步

- 📝 查看 [Context API](context.md) - 了解响应方法
- 🛣️ 学习 [路由系统](routing.md) - 路由配置
- 🚀 探索 [其他功能](../advanced/websocket.md) - WebSocket 等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
