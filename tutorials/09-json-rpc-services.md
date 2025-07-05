# Tutorial 09: JSON-RPC Services

## 📖 Overview

Learn to build JSON-RPC services with Zoox for structured API communication. This tutorial covers service architecture, method registration, error handling, and client integration.

## 🎯 Learning Objectives

- Understand JSON-RPC protocol
- Build RPC services and methods
- Handle RPC errors and validation
- Create RPC clients and documentation
- Implement service discovery

## 📋 Prerequisites

- Completed [Tutorial 01: Getting Started](./01-getting-started.md)
- Understanding of RPC concepts
- Basic knowledge of JSON and APIs

## 🚀 Getting Started

### Basic JSON-RPC Server

```go
package main

import (
    "encoding/json"
    "fmt"
    "reflect"
    
    "github.com/go-zoox/zoox"
)

type RPCRequest struct {
    Jsonrpc string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
    ID      interface{} `json:"id,omitempty"`
}

type RPCResponse struct {
    Jsonrpc string      `json:"jsonrpc"`
    Result  interface{} `json:"result,omitempty"`
    Error   *RPCError   `json:"error,omitempty"`
    ID      interface{} `json:"id,omitempty"`
}

type RPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type RPCServer struct {
    methods map[string]reflect.Value
}

func NewRPCServer() *RPCServer {
    return &RPCServer{
        methods: make(map[string]reflect.Value),
    }
}

func (s *RPCServer) Register(name string, method interface{}) {
    s.methods[name] = reflect.ValueOf(method)
}

func (s *RPCServer) Handle(ctx *zoox.Context) {
    var req RPCRequest
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(400, RPCResponse{
            Jsonrpc: "2.0",
            Error: &RPCError{
                Code:    -32700,
                Message: "Parse error",
            },
            ID: nil,
        })
        return
    }
    
    method, exists := s.methods[req.Method]
    if !exists {
        ctx.JSON(200, RPCResponse{
            Jsonrpc: "2.0",
            Error: &RPCError{
                Code:    -32601,
                Message: "Method not found",
            },
            ID: req.ID,
        })
        return
    }
    
    // Call method
    result, err := s.callMethod(method, req.Params)
    if err != nil {
        ctx.JSON(200, RPCResponse{
            Jsonrpc: "2.0",
            Error: &RPCError{
                Code:    -32603,
                Message: err.Error(),
            },
            ID: req.ID,
        })
        return
    }
    
    ctx.JSON(200, RPCResponse{
        Jsonrpc: "2.0",
        Result:  result,
        ID:      req.ID,
    })
}

func (s *RPCServer) callMethod(method reflect.Value, params interface{}) (interface{}, error) {
    methodType := method.Type()
    
    // Handle different parameter types
    var args []reflect.Value
    
    if params != nil {
        paramsValue := reflect.ValueOf(params)
        
        if methodType.NumIn() == 1 {
            // Single parameter
            args = []reflect.Value{paramsValue}
        } else if methodType.NumIn() > 1 {
            // Multiple parameters - expect array
            if paramsValue.Kind() == reflect.Slice {
                for i := 0; i < paramsValue.Len() && i < methodType.NumIn(); i++ {
                    args = append(args, paramsValue.Index(i))
                }
            }
        }
    }
    
    // Call the method
    results := method.Call(args)
    
    if len(results) == 0 {
        return nil, nil
    }
    
    // Check for error return
    if len(results) == 2 && results[1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
        if !results[1].IsNil() {
            return nil, results[1].Interface().(error)
        }
        return results[0].Interface(), nil
    }
    
    return results[0].Interface(), nil
}

// Example service methods
func Add(a, b float64) float64 {
    return a + b
}

func Subtract(a, b float64) float64 {
    return a - b
}

func Multiply(a, b float64) float64 {
    return a * b
}

func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

func main() {
    app := zoox.New()
    
    // Create RPC server
    rpcServer := NewRPCServer()
    
    // Register methods
    rpcServer.Register("add", Add)
    rpcServer.Register("subtract", Subtract)
    rpcServer.Register("multiply", Multiply)
    rpcServer.Register("divide", Divide)
    
    // RPC endpoint
    app.Post("/rpc", rpcServer.Handle)
    
    // Test client
    app.Get("/", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>JSON-RPC Test Client</title>
            <style>
                body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
                .method { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
                input, button { margin: 5px; padding: 8px; }
                .result { margin-top: 10px; padding: 10px; background: #f0f0f0; border-radius: 3px; }
            </style>
        </head>
        <body>
            <h1>JSON-RPC Test Client</h1>
            
            <div class="method">
                <h3>Add</h3>
                <input type="number" id="add-a" placeholder="First number">
                <input type="number" id="add-b" placeholder="Second number">
                <button onclick="callRPC('add', [parseFloat(document.getElementById('add-a').value), parseFloat(document.getElementById('add-b').value)], 'add-result')">Add</button>
                <div id="add-result" class="result"></div>
            </div>
            
            <div class="method">
                <h3>Divide</h3>
                <input type="number" id="div-a" placeholder="Dividend">
                <input type="number" id="div-b" placeholder="Divisor">
                <button onclick="callRPC('divide', [parseFloat(document.getElementById('div-a').value), parseFloat(document.getElementById('div-b').value)], 'div-result')">Divide</button>
                <div id="div-result" class="result"></div>
            </div>
            
            <script>
                async function callRPC(method, params, resultId) {
                    const request = {
                        jsonrpc: "2.0",
                        method: method,
                        params: params,
                        id: Date.now()
                    };
                    
                    try {
                        const response = await fetch('/rpc', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify(request)
                        });
                        
                        const result = await response.json();
                        
                        const resultDiv = document.getElementById(resultId);
                        if (result.error) {
                            resultDiv.innerHTML = '<strong>Error:</strong> ' + result.error.message;
                            resultDiv.style.color = 'red';
                        } else {
                            resultDiv.innerHTML = '<strong>Result:</strong> ' + result.result;
                            resultDiv.style.color = 'green';
                        }
                    } catch (error) {
                        console.error('RPC call failed:', error);
                    }
                }
            </script>
        </body>
        </html>
        `
        ctx.HTML(200, html, nil)
    })
    
    app.Listen(":8080")
}
```

## 📚 Key Takeaways

1. **RPC Protocol**: Implement JSON-RPC 2.0 specification correctly
2. **Method Registration**: Register and manage RPC methods dynamically
3. **Error Handling**: Proper error codes and messages
4. **Type Safety**: Handle parameter types and validation
5. **Documentation**: Provide clear API documentation

## 🎯 Next Steps

- Learn [Tutorial 10: Authentication & Authorization](./10-authentication-authorization.md)
- Explore [Tutorial 11: Database Integration](./11-database-integration.md)
- Study [Tutorial 12: Caching Strategies](./12-caching-strategies.md)

---

**Congratulations!** You've mastered JSON-RPC services in Zoox! 