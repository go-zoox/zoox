# JSON-RPC 服务

Zoox 支持 JSON-RPC 2.0 协议，可以轻松创建 JSON-RPC 服务。

## 基本用法

### 创建 JSON-RPC 路由

```go
package main

import (
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()
	
	// 创建 JSON-RPC 路由
	app.JSONRPC("/rpc", func(registry zoox.JSONRPCRegistry) {
		// 注册方法
		registry.Register("add", func(params map[string]interface{}) (interface{}, error) {
			a := int(params["a"].(float64))
			b := int(params["b"].(float64))
			return a + b, nil
		})
		
		registry.Register("subtract", func(params map[string]interface{}) (interface{}, error) {
			a := int(params["a"].(float64))
			b := int(params["b"].(float64))
			return a - b, nil
		})
	})
	
	app.Run(":8080")
}
```

**说明**: JSON-RPC 实现参考 `group.go:194-217` 和 `application.go:392-399`。

## 注册方法

### 基本方法注册

```go
app.JSONRPC("/rpc", func(registry zoox.JSONRPCRegistry) {
	registry.Register("hello", func(params map[string]interface{}) (interface{}, error) {
		name := params["name"].(string)
		return map[string]interface{}{
			"message": "Hello, " + name,
		}, nil
	})
})
```

### 使用结构体参数

```go
type AddParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

app.JSONRPC("/rpc", func(registry zoox.JSONRPCRegistry) {
	registry.Register("add", func(params map[string]interface{}) (interface{}, error) {
		var p AddParams
		// 将 params 转换为结构体
		// ...
		return p.A + p.B, nil
	})
})
```

## 错误处理

```go
app.JSONRPC("/rpc", func(registry zoox.JSONRPCRegistry) {
	registry.Register("divide", func(params map[string]interface{}) (interface{}, error) {
		a := int(params["a"].(float64))
		b := int(params["b"].(float64))
		
		if b == 0 {
			return nil, errors.New("division by zero")
		}
		
		return a / b, nil
	})
})
```

## 批量请求

JSON-RPC 支持批量请求：

```go
// 客户端发送
[
	{"jsonrpc": "2.0", "method": "add", "params": {"a": 1, "b": 2}, "id": 1},
	{"jsonrpc": "2.0", "method": "subtract", "params": {"a": 5, "b": 3}, "id": 2}
]

// 服务器返回
[
	{"jsonrpc": "2.0", "result": 3, "id": 1},
	{"jsonrpc": "2.0", "result": 2, "id": 2}
]
```

## 通知（Notification）

通知是不需要响应的请求：

```go
// 客户端发送（没有 id 字段）
{"jsonrpc": "2.0", "method": "notify", "params": {"message": "Hello"}}

// 服务器不返回响应
```

## 完整示例

```go
package main

import (
	"errors"
	
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()
	
	app.JSONRPC("/rpc", func(registry zoox.JSONRPCRegistry) {
		// 加法
		registry.Register("add", func(params map[string]interface{}) (interface{}, error) {
			a := int(params["a"].(float64))
			b := int(params["b"].(float64))
			return a + b, nil
		})
		
		// 减法
		registry.Register("subtract", func(params map[string]interface{}) (interface{}, error) {
			a := int(params["a"].(float64))
			b := int(params["b"].(float64))
			return a - b, nil
		})
		
		// 乘法
		registry.Register("multiply", func(params map[string]interface{}) (interface{}, error) {
			a := int(params["a"].(float64))
			b := int(params["b"].(float64))
			return a * b, nil
		})
		
		// 除法（带错误处理）
		registry.Register("divide", func(params map[string]interface{}) (interface{}, error) {
			a := int(params["a"].(float64))
			b := int(params["b"].(float64))
			
			if b == 0 {
				return nil, errors.New("division by zero")
			}
			
			return a / b, nil
		})
	})
	
	app.Run(":8080")
}
```

## 客户端调用示例

### JavaScript

```javascript
// 单个请求
fetch('http://localhost:8080/rpc', {
	method: 'POST',
	headers: {
		'Content-Type': 'application/json',
	},
	body: JSON.stringify({
		jsonrpc: '2.0',
		method: 'add',
		params: { a: 1, b: 2 },
		id: 1
	})
})
.then(res => res.json())
.then(data => console.log(data));

// 批量请求
fetch('http://localhost:8080/rpc', {
	method: 'POST',
	headers: {
		'Content-Type': 'application/json',
	},
	body: JSON.stringify([
		{jsonrpc: '2.0', method: 'add', params: {a: 1, b: 2}, id: 1},
		{jsonrpc: '2.0', method: 'subtract', params: {a: 5, b: 3}, id: 2}
	])
})
.then(res => res.json())
.then(data => console.log(data));
```

### cURL

```bash
# 单个请求
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "add",
    "params": {"a": 1, "b": 2},
    "id": 1
  }'

# 批量请求
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '[
    {"jsonrpc": "2.0", "method": "add", "params": {"a": 1, "b": 2}, "id": 1},
    {"jsonrpc": "2.0", "method": "subtract", "params": {"a": 5, "b": 3}, "id": 2}
  ]'
```

## 中间件支持

JSON-RPC 路由支持中间件：

```go
app.Use(middleware.Logger())
app.Use(middleware.Recovery())

app.JSONRPC("/rpc", func(registry zoox.JSONRPCRegistry) {
	// 注册方法
})
```

## 认证

为 JSON-RPC 添加认证：

```go
rpc := app.Group("/rpc")
rpc.Use(middleware.JWT())  // 或 BasicAuth、BearerToken

rpc.JSONRPC("", func(registry zoox.JSONRPCRegistry) {
	// 注册方法
})
```

## 下一步

- 🔌 学习 [代理功能](proxy.md) - 反向代理和路径重写
- ⏰ 查看 [定时任务](cron-jobs.md) - Cron 任务调度
- 🚀 探索 [其他高级功能](websocket.md) - WebSocket 等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
