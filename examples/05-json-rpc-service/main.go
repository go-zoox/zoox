package main

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

// Math service for JSON-RPC
type MathService struct{}

// Add method
func (m *MathService) Add(ctx context.Context, args *AddArgs, reply *AddReply) error {
	reply.Result = args.A + args.B
	return nil
}

// Subtract method
func (m *MathService) Subtract(ctx context.Context, args *SubtractArgs, reply *SubtractReply) error {
	reply.Result = args.A - args.B
	return nil
}

// Multiply method
func (m *MathService) Multiply(ctx context.Context, args *MultiplyArgs, reply *MultiplyReply) error {
	reply.Result = args.A * args.B
	return nil
}

// Divide method
func (m *MathService) Divide(ctx context.Context, args *DivideArgs, reply *DivideReply) error {
	if args.B == 0 {
		return &JSONRPCError{
			Code:    -32000,
			Message: "Division by zero",
		}
	}
	reply.Result = args.A / args.B
	return nil
}

// Power method
func (m *MathService) Power(ctx context.Context, args *PowerArgs, reply *PowerReply) error {
	reply.Result = math.Pow(args.Base, args.Exponent)
	return nil
}

// Sqrt method
func (m *MathService) Sqrt(ctx context.Context, args *SqrtArgs, reply *SqrtReply) error {
	if args.Number < 0 {
		return &JSONRPCError{
			Code:    -32000,
			Message: "Cannot calculate square root of negative number",
		}
	}
	reply.Result = math.Sqrt(args.Number)
	return nil
}

// User service for JSON-RPC
type UserService struct {
	users map[int]*User
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[int]*User),
	}
}

// GetUser method
func (u *UserService) GetUser(ctx context.Context, args *GetUserArgs, reply *GetUserReply) error {
	user, exists := u.users[args.ID]
	if !exists {
		return &JSONRPCError{
			Code:    -32000,
			Message: "User not found",
		}
	}
	reply.User = user
	return nil
}

// CreateUser method
func (u *UserService) CreateUser(ctx context.Context, args *CreateUserArgs, reply *CreateUserReply) error {
	if args.Name == "" {
		return &JSONRPCError{
			Code:    -32000,
			Message: "Name is required",
		}
	}
	if args.Email == "" {
		return &JSONRPCError{
			Code:    -32000,
			Message: "Email is required",
		}
	}

	// Generate new ID
	id := len(u.users) + 1
	user := &User{
		ID:        id,
		Name:      args.Name,
		Email:     args.Email,
		CreatedAt: time.Now(),
	}
	
	u.users[id] = user
	reply.User = user
	return nil
}

// ListUsers method
func (u *UserService) ListUsers(ctx context.Context, args *ListUsersArgs, reply *ListUsersReply) error {
	var users []*User
	for _, user := range u.users {
		users = append(users, user)
	}
	reply.Users = users
	reply.Total = len(users)
	return nil
}

// UpdateUser method
func (u *UserService) UpdateUser(ctx context.Context, args *UpdateUserArgs, reply *UpdateUserReply) error {
	user, exists := u.users[args.ID]
	if !exists {
		return &JSONRPCError{
			Code:    -32000,
			Message: "User not found",
		}
	}

	if args.Name != "" {
		user.Name = args.Name
	}
	if args.Email != "" {
		user.Email = args.Email
	}
	user.UpdatedAt = time.Now()

	reply.User = user
	return nil
}

// DeleteUser method
func (u *UserService) DeleteUser(ctx context.Context, args *DeleteUserArgs, reply *DeleteUserReply) error {
	_, exists := u.users[args.ID]
	if !exists {
		return &JSONRPCError{
			Code:    -32000,
			Message: "User not found",
		}
	}

	delete(u.users, args.ID)
	reply.Success = true
	return nil
}

// Data structures
type AddArgs struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

type AddReply struct {
	Result float64 `json:"result"`
}

type SubtractArgs struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

type SubtractReply struct {
	Result float64 `json:"result"`
}

type MultiplyArgs struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

type MultiplyReply struct {
	Result float64 `json:"result"`
}

type DivideArgs struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

type DivideReply struct {
	Result float64 `json:"result"`
}

type PowerArgs struct {
	Base     float64 `json:"base"`
	Exponent float64 `json:"exponent"`
}

type PowerReply struct {
	Result float64 `json:"result"`
}

type SqrtArgs struct {
	Number float64 `json:"number"`
}

type SqrtReply struct {
	Result float64 `json:"result"`
}

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetUserArgs struct {
	ID int `json:"id"`
}

type GetUserReply struct {
	User *User `json:"user"`
}

type CreateUserArgs struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserReply struct {
	User *User `json:"user"`
}

type ListUsersArgs struct{}

type ListUsersReply struct {
	Users []*User `json:"users"`
	Total int     `json:"total"`
}

type UpdateUserArgs struct {
	ID    int    `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type UpdateUserReply struct {
	User *User `json:"user"`
}

type DeleteUserArgs struct {
	ID int `json:"id"`
}

type DeleteUserReply struct {
	Success bool `json:"success"`
}

// JSON-RPC Error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *JSONRPCError) Error() string {
	return e.Message
}

func main() {
	app := zoox.Default()

	// Enable CORS
	app.Use(middleware.CORS())

	// Serve static files for the test page
	app.Get("/", func(ctx *zoox.Context) {
		ctx.HTML(200, jsonRPCPageHTML)
	})

	// Create services
	mathService := &MathService{}
	userService := NewUserService()

	// Register JSON-RPC services
	app.JSONRPC("/rpc/math", mathService)
	app.JSONRPC("/rpc/user", userService)

	// REST API for comparison
	app.Get("/api/health", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"services": []string{
				"Math Service - /rpc/math",
				"User Service - /rpc/user",
			},
		})
	})

	// Documentation endpoint
	app.Get("/api/docs", func(ctx *zoox.Context) {
		ctx.JSON(200, zoox.H{
			"math_service": zoox.H{
				"endpoint": "/rpc/math",
				"methods": []zoox.H{
					{"name": "Add", "params": "a, b (numbers)", "returns": "result (number)"},
					{"name": "Subtract", "params": "a, b (numbers)", "returns": "result (number)"},
					{"name": "Multiply", "params": "a, b (numbers)", "returns": "result (number)"},
					{"name": "Divide", "params": "a, b (numbers)", "returns": "result (number)"},
					{"name": "Power", "params": "base, exponent (numbers)", "returns": "result (number)"},
					{"name": "Sqrt", "params": "number (number)", "returns": "result (number)"},
				},
			},
			"user_service": zoox.H{
				"endpoint": "/rpc/user",
				"methods": []zoox.H{
					{"name": "GetUser", "params": "id (number)", "returns": "user (object)"},
					{"name": "CreateUser", "params": "name, email (strings)", "returns": "user (object)"},
					{"name": "ListUsers", "params": "none", "returns": "users (array), total (number)"},
					{"name": "UpdateUser", "params": "id (number), name, email (strings, optional)", "returns": "user (object)"},
					{"name": "DeleteUser", "params": "id (number)", "returns": "success (boolean)"},
				},
			},
		})
	})

	log.Println("JSON-RPC Service starting on http://localhost:8080")
	log.Println("\nAvailable Services:")
	log.Println("  Math Service: /rpc/math")
	log.Println("  User Service: /rpc/user")
	log.Println("\nTest Interface: http://localhost:8080")
	log.Println("API Documentation: http://localhost:8080/api/docs")

	app.Run(":8080")
}

const jsonRPCPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>JSON-RPC Service Test</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        .service-section { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .service-section h3 { margin-top: 0; color: #007bff; }
        .method-group { background: white; padding: 15px; margin: 10px 0; border-radius: 3px; border: 1px solid #ddd; }
        .method-group h4 { margin-top: 0; color: #333; }
        .form-group { margin: 10px 0; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .form-group input { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 3px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; margin: 5px; }
        button:hover { background: #0056b3; }
        .result { background: #e9ecef; padding: 10px; border-radius: 3px; margin: 10px 0; font-family: monospace; white-space: pre-wrap; }
        .error { background: #f8d7da; color: #721c24; }
        .success { background: #d4edda; color: #155724; }
        .flex-container { display: flex; gap: 20px; }
        .flex-item { flex: 1; }
    </style>
</head>
<body>
    <h1>🚀 Zoox JSON-RPC Service Test</h1>
    <p>This page demonstrates JSON-RPC services with the Zoox framework.</p>

    <div class="flex-container">
        <div class="flex-item">
            <div class="service-section">
                <h3>Math Service</h3>
                <p>Endpoint: <code>/rpc/math</code></p>

                <div class="method-group">
                    <h4>Add</h4>
                    <div class="form-group">
                        <label>A:</label>
                        <input type="number" id="add-a" value="10" step="any">
                    </div>
                    <div class="form-group">
                        <label>B:</label>
                        <input type="number" id="add-b" value="5" step="any">
                    </div>
                    <button onclick="callMathMethod('Add', {a: parseFloat(document.getElementById('add-a').value), b: parseFloat(document.getElementById('add-b').value)})">Add</button>
                </div>

                <div class="method-group">
                    <h4>Subtract</h4>
                    <div class="form-group">
                        <label>A:</label>
                        <input type="number" id="sub-a" value="10" step="any">
                    </div>
                    <div class="form-group">
                        <label>B:</label>
                        <input type="number" id="sub-b" value="3" step="any">
                    </div>
                    <button onclick="callMathMethod('Subtract', {a: parseFloat(document.getElementById('sub-a').value), b: parseFloat(document.getElementById('sub-b').value)})">Subtract</button>
                </div>

                <div class="method-group">
                    <h4>Multiply</h4>
                    <div class="form-group">
                        <label>A:</label>
                        <input type="number" id="mul-a" value="4" step="any">
                    </div>
                    <div class="form-group">
                        <label>B:</label>
                        <input type="number" id="mul-b" value="7" step="any">
                    </div>
                    <button onclick="callMathMethod('Multiply', {a: parseFloat(document.getElementById('mul-a').value), b: parseFloat(document.getElementById('mul-b').value)})">Multiply</button>
                </div>

                <div class="method-group">
                    <h4>Divide</h4>
                    <div class="form-group">
                        <label>A:</label>
                        <input type="number" id="div-a" value="20" step="any">
                    </div>
                    <div class="form-group">
                        <label>B:</label>
                        <input type="number" id="div-b" value="4" step="any">
                    </div>
                    <button onclick="callMathMethod('Divide', {a: parseFloat(document.getElementById('div-a').value), b: parseFloat(document.getElementById('div-b').value)})">Divide</button>
                </div>

                <div class="method-group">
                    <h4>Power</h4>
                    <div class="form-group">
                        <label>Base:</label>
                        <input type="number" id="pow-base" value="2" step="any">
                    </div>
                    <div class="form-group">
                        <label>Exponent:</label>
                        <input type="number" id="pow-exp" value="3" step="any">
                    </div>
                    <button onclick="callMathMethod('Power', {base: parseFloat(document.getElementById('pow-base').value), exponent: parseFloat(document.getElementById('pow-exp').value)})">Power</button>
                </div>

                <div class="method-group">
                    <h4>Square Root</h4>
                    <div class="form-group">
                        <label>Number:</label>
                        <input type="number" id="sqrt-num" value="16" step="any">
                    </div>
                    <button onclick="callMathMethod('Sqrt', {number: parseFloat(document.getElementById('sqrt-num').value)})">Square Root</button>
                </div>
            </div>
        </div>

        <div class="flex-item">
            <div class="service-section">
                <h3>User Service</h3>
                <p>Endpoint: <code>/rpc/user</code></p>

                <div class="method-group">
                    <h4>Create User</h4>
                    <div class="form-group">
                        <label>Name:</label>
                        <input type="text" id="create-name" value="John Doe">
                    </div>
                    <div class="form-group">
                        <label>Email:</label>
                        <input type="email" id="create-email" value="john@example.com">
                    </div>
                    <button onclick="callUserMethod('CreateUser', {name: document.getElementById('create-name').value, email: document.getElementById('create-email').value})">Create User</button>
                </div>

                <div class="method-group">
                    <h4>Get User</h4>
                    <div class="form-group">
                        <label>User ID:</label>
                        <input type="number" id="get-id" value="1">
                    </div>
                    <button onclick="callUserMethod('GetUser', {id: parseInt(document.getElementById('get-id').value)})">Get User</button>
                </div>

                <div class="method-group">
                    <h4>List Users</h4>
                    <button onclick="callUserMethod('ListUsers', {})">List All Users</button>
                </div>

                <div class="method-group">
                    <h4>Update User</h4>
                    <div class="form-group">
                        <label>User ID:</label>
                        <input type="number" id="update-id" value="1">
                    </div>
                    <div class="form-group">
                        <label>Name:</label>
                        <input type="text" id="update-name" value="Jane Doe">
                    </div>
                    <div class="form-group">
                        <label>Email:</label>
                        <input type="email" id="update-email" value="jane@example.com">
                    </div>
                    <button onclick="callUserMethod('UpdateUser', {id: parseInt(document.getElementById('update-id').value), name: document.getElementById('update-name').value, email: document.getElementById('update-email').value})">Update User</button>
                </div>

                <div class="method-group">
                    <h4>Delete User</h4>
                    <div class="form-group">
                        <label>User ID:</label>
                        <input type="number" id="delete-id" value="1">
                    </div>
                    <button onclick="callUserMethod('DeleteUser', {id: parseInt(document.getElementById('delete-id').value)})">Delete User</button>
                </div>
            </div>
        </div>
    </div>

    <div class="service-section">
        <h3>Results</h3>
        <div id="results" class="result">Results will appear here...</div>
    </div>

    <script>
        function callMathMethod(method, params) {
            callJSONRPC('/rpc/math', method, params);
        }

        function callUserMethod(method, params) {
            callJSONRPC('/rpc/user', method, params);
        }

        function callJSONRPC(endpoint, method, params) {
            const request = {
                jsonrpc: '2.0',
                method: method,
                params: params,
                id: Date.now()
            };

            fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(request)
            })
            .then(response => response.json())
            .then(data => {
                const resultsDiv = document.getElementById('results');
                resultsDiv.className = 'result ' + (data.error ? 'error' : 'success');
                resultsDiv.textContent = JSON.stringify(data, null, 2);
            })
            .catch(error => {
                const resultsDiv = document.getElementById('results');
                resultsDiv.className = 'result error';
                resultsDiv.textContent = 'Error: ' + error.message;
            });
        }
    </script>
</body>
</html>` 