package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-zoox/zoox"
)

// User represents a user in our system
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserStore manages users in memory
type UserStore struct {
	users  map[int]*User
	nextID int
	mutex  sync.RWMutex
}

// NewUserStore creates a new user store
func NewUserStore() *UserStore {
	store := &UserStore{
		users:  make(map[int]*User),
		nextID: 1,
	}
	
	// Add some sample data
	store.CreateUser("John Doe", "john@example.com")
	store.CreateUser("Jane Smith", "jane@example.com")
	
	return store
}

// CreateUser creates a new user
func (s *UserStore) CreateUser(name, email string) *User {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	user := &User{
		ID:        s.nextID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	s.users[s.nextID] = user
	s.nextID++
	
	return user
}

// GetUser retrieves a user by ID
func (s *UserStore) GetUser(id int) (*User, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	user, exists := s.users[id]
	return user, exists
}

// GetAllUsers retrieves all users
func (s *UserStore) GetAllUsers() []*User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	
	return users
}

// UpdateUser updates an existing user
func (s *UserStore) UpdateUser(id int, name, email string) (*User, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	user, exists := s.users[id]
	if !exists {
		return nil, false
	}
	
	user.Name = name
	user.Email = email
	user.UpdatedAt = time.Now()
	
	return user, true
}

// DeleteUser deletes a user by ID
func (s *UserStore) DeleteUser(id int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	_, exists := s.users[id]
	if exists {
		delete(s.users, id)
	}
	
	return exists
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func main() {
	// Create Zoox app
	app := zoox.Default()
	
	// Initialize user store
	userStore := NewUserStore()
	
	// Custom middleware for basic authentication (demo purposes)
	authMiddleware := func(ctx *zoox.Context) {
		// Simple hardcoded authentication for demo
		auth := ctx.Header("Authorization")
		if auth != "Bearer demo-token" {
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid or missing authorization token",
				Code:    http.StatusUnauthorized,
			})
			return
		}
		ctx.Next()
	}
	
	// Global middleware
	app.Use(func(ctx *zoox.Context) {
		// Log all requests
		fmt.Printf("[%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), ctx.Method, ctx.Path)
		ctx.Next()
	})
	
	// Health check endpoint
	app.Get("/health", func(ctx *zoox.Context) {
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})
	
	// Public API routes
	public := app.Group("/api/v1")
	
	// Get all users (public)
	public.Get("/users", func(ctx *zoox.Context) {
		users := userStore.GetAllUsers()
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"users": users,
			"count": len(users),
		})
	})
	
	// Get user by ID (public)
	public.Get("/users/:id", func(ctx *zoox.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid user ID format",
				Code:    http.StatusBadRequest,
			})
			return
		}
		
		user, exists := userStore.GetUser(id)
		if !exists {
			ctx.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"user": user,
		})
	})
	
	// Protected API routes (require authentication)
	protected := app.Group("/api/v1", authMiddleware)
	
	// Create user (protected)
	protected.Post("/users", func(ctx *zoox.Context) {
		var req CreateUserRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid JSON payload",
				Code:    http.StatusBadRequest,
			})
			return
		}
		
		// Basic validation
		if req.Name == "" || req.Email == "" {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "Name and email are required",
				Code:    http.StatusBadRequest,
			})
			return
		}
		
		user := userStore.CreateUser(req.Name, req.Email)
		ctx.JSON(http.StatusCreated, map[string]interface{}{
			"message": "User created successfully",
			"user":    user,
		})
	})
	
	// Update user (protected)
	protected.Put("/users/:id", func(ctx *zoox.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid user ID format",
				Code:    http.StatusBadRequest,
			})
			return
		}
		
		var req UpdateUserRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid JSON payload",
				Code:    http.StatusBadRequest,
			})
			return
		}
		
		// Basic validation
		if req.Name == "" || req.Email == "" {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "Name and email are required",
				Code:    http.StatusBadRequest,
			})
			return
		}
		
		user, exists := userStore.UpdateUser(id, req.Name, req.Email)
		if !exists {
			ctx.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "User updated successfully",
			"user":    user,
		})
	})
	
	// Delete user (protected)
	protected.Delete("/users/:id", func(ctx *zoox.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid user ID format",
				Code:    http.StatusBadRequest,
			})
			return
		}
		
		exists := userStore.DeleteUser(id)
		if !exists {
			ctx.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "User deleted successfully",
		})
	})
	
	// API documentation endpoint
	app.Get("/api/docs", func(ctx *zoox.Context) {
		docs := map[string]interface{}{
			"title":       "Basic Server API",
			"version":     "1.0.0",
			"description": "A simple REST API demonstrating Zoox framework features",
			"endpoints": map[string]interface{}{
				"GET /health":          "Health check",
				"GET /api/v1/users":    "Get all users (public)",
				"GET /api/v1/users/:id": "Get user by ID (public)",
				"POST /api/v1/users":   "Create user (requires auth)",
				"PUT /api/v1/users/:id": "Update user (requires auth)",
				"DELETE /api/v1/users/:id": "Delete user (requires auth)",
			},
			"authentication": map[string]interface{}{
				"type":   "Bearer Token",
				"header": "Authorization: Bearer demo-token",
				"note":   "Use 'Bearer demo-token' for protected endpoints",
			},
		}
		
		ctx.JSON(http.StatusOK, docs)
	})
	
	// Start server
	fmt.Println("🚀 Basic Server starting...")
	fmt.Println("📍 Server running on http://localhost:8080")
	fmt.Println("📚 API Documentation: http://localhost:8080/api/docs")
	fmt.Println("🔍 Health Check: http://localhost:8080/health")
	fmt.Println("👥 Users API: http://localhost:8080/api/v1/users")
	fmt.Println("🔐 Use 'Authorization: Bearer demo-token' for protected endpoints")
	
	app.Run(":8080")
} 