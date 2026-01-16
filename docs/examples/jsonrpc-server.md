# JSON-RPC 服务器示例

这是一个 JSON-RPC 服务器示例，展示了如何使用 Zoox 构建 JSON-RPC 2.0 服务。

## 项目结构

```
jsonrpc-server/
├── main.go
├── go.mod
└── handlers/
    └── calculator.go
```

## 完整代码

### main.go

```go
package main

import (
	"errors"
	"log"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
	"jsonrpc-server/handlers"
)

func main() {
	app := zoox.New()

	// 全局中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())

	// JSON-RPC 路由 - 计算器服务
	app.JSONRPC("/rpc", func(registry zoox.JSONRPCRegistry) {
		// 注册计算方法
		registry.Register("add", handlers.Add)
		registry.Register("subtract", handlers.Subtract)
		registry.Register("multiply", handlers.Multiply)
		registry.Register("divide", handlers.Divide)
		
		// 注册工具方法
		registry.Register("echo", handlers.Echo)
		registry.Register("ping", handlers.Ping)
	})

	// 健康检查
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"status": "ok",
			"service": "jsonrpc-server",
		})
	})

	// API 文档
	app.Get("/methods", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"methods": []string{
				"add",
				"subtract",
				"multiply",
				"divide",
				"echo",
				"ping",
			},
			"endpoint": "/rpc",
		})
	})

	log.Println("JSON-RPC server started on :8080")
	app.Run(":8080")
}
```

### handlers/calculator.go

```go
package handlers

import (
	"errors"
	"fmt"
	"time"
)

// Add 加法
func Add(params map[string]interface{}) (interface{}, error) {
	a, ok1 := params["a"].(float64)
	b, ok2 := params["b"].(float64)
	
	if !ok1 || !ok2 {
		return nil, errors.New("invalid parameters: a and b must be numbers")
	}
	
	return int(a) + int(b), nil
}

// Subtract 减法
func Subtract(params map[string]interface{}) (interface{}, error) {
	a, ok1 := params["a"].(float64)
	b, ok2 := params["b"].(float64)
	
	if !ok1 || !ok2 {
		return nil, errors.New("invalid parameters: a and b must be numbers")
	}
	
	return int(a) - int(b), nil
}

// Multiply 乘法
func Multiply(params map[string]interface{}) (interface{}, error) {
	a, ok1 := params["a"].(float64)
	b, ok2 := params["b"].(float64)
	
	if !ok1 || !ok2 {
		return nil, errors.New("invalid parameters: a and b must be numbers")
	}
	
	return int(a) * int(b), nil
}

// Divide 除法
func Divide(params map[string]interface{}) (interface{}, error) {
	a, ok1 := params["a"].(float64)
	b, ok2 := params["b"].(float64)
	
	if !ok1 || !ok2 {
		return nil, errors.New("invalid parameters: a and b must be numbers")
	}
	
	if b == 0 {
		return nil, errors.New("division by zero")
	}
	
	return float64(a) / float64(b), nil
}

// Echo 回显
func Echo(params map[string]interface{}) (interface{}, error) {
	message, ok := params["message"].(string)
	if !ok {
		return nil, errors.New("invalid parameter: message must be a string")
	}
	
	return map[string]interface{}{
		"echo": message,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}, nil
}

// Ping 心跳检测
func Ping(params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"pong": true,
		"message": "Server is alive",
	}, nil
}
```

## 使用示例

### 单个请求

#### 请求

```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "add",
    "params": {"a": 10, "b": 20},
    "id": 1
  }'
```

#### 响应

```json
{
  "jsonrpc": "2.0",
  "result": 30,
  "id": 1
}
```

### 批量请求

#### 请求

```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '[
    {"jsonrpc": "2.0", "method": "add", "params": {"a": 1, "b": 2}, "id": 1},
    {"jsonrpc": "2.0", "method": "multiply", "params": {"a": 3, "b": 4}, "id": 2},
    {"jsonrpc": "2.0", "method": "subtract", "params": {"a": 10, "b": 3}, "id": 3}
  ]'
```

#### 响应

```json
[
  {"jsonrpc": "2.0", "result": 3, "id": 1},
  {"jsonrpc": "2.0", "result": 12, "id": 2},
  {"jsonrpc": "2.0", "result": 7, "id": 3}
]
```

### 错误处理

#### 请求

```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "divide",
    "params": {"a": 10, "b": 0},
    "id": 1
  }'
```

#### 响应

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "division by zero"
  },
  "id": 1
}
```

### 通知（Notification）

通知是不需要响应的请求（没有 `id` 字段）：

```bash
curl -X POST http://localhost:8080/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "ping",
    "params": {}
  }'
```

服务器不会返回响应。

## JavaScript 客户端示例

### 使用 fetch API

```javascript
// 单个请求
async function jsonRpcCall(method, params, id = 1) {
	const response = await fetch('http://localhost:8080/rpc', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({
			jsonrpc: '2.0',
			method: method,
			params: params,
			id: id,
		}),
	});
	
	const data = await response.json();
	
	if (data.error) {
		throw new Error(data.error.message);
	}
	
	return data.result;
}

// 使用示例
const result = await jsonRpcCall('add', { a: 10, b: 20 });
console.log(result); // 30
```

### 批量请求

```javascript
async function jsonRpcBatch(requests) {
	const response = await fetch('http://localhost:8080/rpc', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(requests),
	});
	
	return await response.json();
}

// 使用示例
const results = await jsonRpcBatch([
	{ jsonrpc: '2.0', method: 'add', params: { a: 1, b: 2 }, id: 1 },
	{ jsonrpc: '2.0', method: 'multiply', params: { a: 3, b: 4 }, id: 2 },
]);
console.log(results);
```

## Python 客户端示例

```python
import requests

def jsonrpc_call(method, params, id=1):
    response = requests.post(
        'http://localhost:8080/rpc',
        json={
            'jsonrpc': '2.0',
            'method': method,
            'params': params,
            'id': id,
        }
    )
    data = response.json()
    
    if 'error' in data:
        raise Exception(data['error']['message'])
    
    return data['result']

# 使用示例
result = jsonrpc_call('add', {'a': 10, 'b': 20})
print(result)  # 30
```

## 功能特性

1. **JSON-RPC 2.0 标准**: 完全支持 JSON-RPC 2.0 协议
2. **批量请求**: 支持一次请求调用多个方法
3. **通知支持**: 支持不需要响应的通知请求
4. **错误处理**: 完善的错误处理和错误码支持
5. **类型安全**: 可以使用结构体进行参数验证

## 最佳实践

1. **参数验证**: 在方法处理函数中验证参数类型和有效性
2. **错误处理**: 使用标准的错误信息格式
3. **方法命名**: 使用清晰的命名空间和方法名
4. **文档化**: 为每个方法提供清晰的文档说明

## 下一步

- 📡 查看 [RESTful API 示例](rest-api.md) - REST API 开发
- 🔌 查看 [WebSocket 应用示例](real-time-app.md) - WebSocket 应用
- 🏗️ 学习 [API Gateway 示例](api-gateway.md) - API 网关
- 📚 阅读 [最佳实践](../best-practices.md) - 开发建议
