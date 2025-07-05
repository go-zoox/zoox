# Tutorial 14: Testing Strategies

## Overview
Learn comprehensive testing strategies for Zoox applications, including unit tests, integration tests, and end-to-end testing approaches.

## Learning Objectives
- Write unit tests for handlers and middleware
- Create integration tests for API endpoints
- Mock external dependencies
- Test WebSocket connections
- Performance and load testing
- Test coverage analysis

## Prerequisites
- Complete Tutorial 13: Monitoring & Logging
- Basic understanding of Go testing
- Familiarity with testing frameworks

## Testing Fundamentals

### Basic Unit Testing

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/go-zoox/zoox"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// User represents a user in our system
type User struct {
    ID       int       `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Created  time.Time `json:"created"`
}

// UserService handles user operations
type UserService struct {
    users map[int]*User
    nextID int
}

func NewUserService() *UserService {
    return &UserService{
        users: make(map[int]*User),
        nextID: 1,
    }
}

func (s *UserService) CreateUser(name, email string) *User {
    user := &User{
        ID:      s.nextID,
        Name:    name,
        Email:   email,
        Created: time.Now(),
    }
    s.users[s.nextID] = user
    s.nextID++
    return user
}

func (s *UserService) GetUser(id int) *User {
    return s.users[id]
}

func (s *UserService) GetAllUsers() []*User {
    var users []*User
    for _, user := range s.users {
        users = append(users, user)
    }
    return users
}

// Handlers
func createUserHandler(service *UserService) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        var req struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }
        
        if err := ctx.BindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Invalid JSON",
            })
            return
        }
        
        if req.Name == "" || req.Email == "" {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Name and email are required",
            })
            return
        }
        
        user := service.CreateUser(req.Name, req.Email)
        ctx.JSON(http.StatusCreated, user)
    }
}

func getUserHandler(service *UserService) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        id := ctx.Param("id")
        userID := 0
        if _, err := fmt.Sscanf(id, "%d", &userID); err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Invalid user ID",
            })
            return
        }
        
        user := service.GetUser(userID)
        if user == nil {
            ctx.JSON(http.StatusNotFound, map[string]string{
                "error": "User not found",
            })
            return
        }
        
        ctx.JSON(http.StatusOK, user)
    }
}

// Unit Tests
func TestUserService_CreateUser(t *testing.T) {
    service := NewUserService()
    
    user := service.CreateUser("John Doe", "john@example.com")
    
    assert.NotNil(t, user)
    assert.Equal(t, 1, user.ID)
    assert.Equal(t, "John Doe", user.Name)
    assert.Equal(t, "john@example.com", user.Email)
    assert.False(t, user.Created.IsZero())
}

func TestUserService_GetUser(t *testing.T) {
    service := NewUserService()
    
    // Create a user
    created := service.CreateUser("Jane Doe", "jane@example.com")
    
    // Get the user
    retrieved := service.GetUser(created.ID)
    
    assert.NotNil(t, retrieved)
    assert.Equal(t, created.ID, retrieved.ID)
    assert.Equal(t, created.Name, retrieved.Name)
    assert.Equal(t, created.Email, retrieved.Email)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
    service := NewUserService()
    
    user := service.GetUser(999)
    
    assert.Nil(t, user)
}

// Integration Tests
func TestCreateUserHandler(t *testing.T) {
    service := NewUserService()
    app := zoox.New()
    app.Post("/users", createUserHandler(service))
    
    tests := []struct {
        name           string
        payload        interface{}
        expectedStatus int
        expectedError  string
    }{
        {
            name: "Valid user creation",
            payload: map[string]string{
                "name":  "John Doe",
                "email": "john@example.com",
            },
            expectedStatus: http.StatusCreated,
        },
        {
            name:           "Missing name",
            payload:        map[string]string{"email": "john@example.com"},
            expectedStatus: http.StatusBadRequest,
            expectedError:  "Name and email are required",
        },
        {
            name:           "Missing email",
            payload:        map[string]string{"name": "John Doe"},
            expectedStatus: http.StatusBadRequest,
            expectedError:  "Name and email are required",
        },
        {
            name:           "Invalid JSON",
            payload:        "invalid json",
            expectedStatus: http.StatusBadRequest,
            expectedError:  "Invalid JSON",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var body bytes.Buffer
            json.NewEncoder(&body).Encode(tt.payload)
            
            req := httptest.NewRequest(http.MethodPost, "/users", &body)
            req.Header.Set("Content-Type", "application/json")
            
            w := httptest.NewRecorder()
            app.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.expectedError != "" {
                var response map[string]string
                err := json.NewDecoder(w.Body).Decode(&response)
                require.NoError(t, err)
                assert.Equal(t, tt.expectedError, response["error"])
            } else {
                var user User
                err := json.NewDecoder(w.Body).Decode(&user)
                require.NoError(t, err)
                assert.NotZero(t, user.ID)
                assert.NotEmpty(t, user.Name)
                assert.NotEmpty(t, user.Email)
            }
        })
    }
}

func TestGetUserHandler(t *testing.T) {
    service := NewUserService()
    created := service.CreateUser("John Doe", "john@example.com")
    
    app := zoox.New()
    app.Get("/users/:id", getUserHandler(service))
    
    tests := []struct {
        name           string
        userID         string
        expectedStatus int
        expectedError  string
    }{
        {
            name:           "Valid user ID",
            userID:         fmt.Sprintf("%d", created.ID),
            expectedStatus: http.StatusOK,
        },
        {
            name:           "User not found",
            userID:         "999",
            expectedStatus: http.StatusNotFound,
            expectedError:  "User not found",
        },
        {
            name:           "Invalid user ID",
            userID:         "abc",
            expectedStatus: http.StatusBadRequest,
            expectedError:  "Invalid user ID",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
            w := httptest.NewRecorder()
            
            app.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.expectedError != "" {
                var response map[string]string
                err := json.NewDecoder(w.Body).Decode(&response)
                require.NoError(t, err)
                assert.Equal(t, tt.expectedError, response["error"])
            } else {
                var user User
                err := json.NewDecoder(w.Body).Decode(&user)
                require.NoError(t, err)
                assert.Equal(t, created.ID, user.ID)
                assert.Equal(t, created.Name, user.Name)
                assert.Equal(t, created.Email, user.Email)
            }
        })
    }
}
```

## Middleware Testing

```go
// Custom middleware for testing
func authMiddleware(validToken string) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        token := ctx.Header("Authorization")
        if token != "Bearer "+validToken {
            ctx.JSON(http.StatusUnauthorized, map[string]string{
                "error": "Unauthorized",
            })
            ctx.Abort()
            return
        }
        ctx.Next()
    }
}

func TestAuthMiddleware(t *testing.T) {
    app := zoox.New()
    validToken := "test-token"
    
    app.Use(authMiddleware(validToken))
    app.Get("/protected", func(ctx *zoox.Context) {
        ctx.JSON(http.StatusOK, map[string]string{
            "message": "Access granted",
        })
    })
    
    tests := []struct {
        name           string
        token          string
        expectedStatus int
        expectedError  string
    }{
        {
            name:           "Valid token",
            token:          "Bearer " + validToken,
            expectedStatus: http.StatusOK,
        },
        {
            name:           "Invalid token",
            token:          "Bearer invalid-token",
            expectedStatus: http.StatusUnauthorized,
            expectedError:  "Unauthorized",
        },
        {
            name:           "Missing token",
            token:          "",
            expectedStatus: http.StatusUnauthorized,
            expectedError:  "Unauthorized",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, "/protected", nil)
            if tt.token != "" {
                req.Header.Set("Authorization", tt.token)
            }
            
            w := httptest.NewRecorder()
            app.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.expectedError != "" {
                var response map[string]string
                err := json.NewDecoder(w.Body).Decode(&response)
                require.NoError(t, err)
                assert.Equal(t, tt.expectedError, response["error"])
            }
        })
    }
}
```

## WebSocket Testing

```go
func TestWebSocketConnection(t *testing.T) {
    app := zoox.New()
    
    app.Get("/ws", func(ctx *zoox.Context) {
        ws, err := ctx.WebSocket()
        if err != nil {
            t.Errorf("WebSocket upgrade failed: %v", err)
            return
        }
        defer ws.Close()
        
        for {
            messageType, message, err := ws.ReadMessage()
            if err != nil {
                break
            }
            
            // Echo the message back
            err = ws.WriteMessage(messageType, message)
            if err != nil {
                break
            }
        }
    })
    
    server := httptest.NewServer(app)
    defer server.Close()
    
    // Convert HTTP URL to WebSocket URL
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
    
    // Connect to WebSocket
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    require.NoError(t, err)
    defer conn.Close()
    
    // Send a message
    testMessage := "Hello WebSocket"
    err = conn.WriteMessage(websocket.TextMessage, []byte(testMessage))
    require.NoError(t, err)
    
    // Read the echoed message
    messageType, message, err := conn.ReadMessage()
    require.NoError(t, err)
    assert.Equal(t, websocket.TextMessage, messageType)
    assert.Equal(t, testMessage, string(message))
}
```

## Mock Testing

```go
// Mock external service
type EmailService interface {
    SendEmail(to, subject, body string) error
}

type MockEmailService struct {
    SentEmails []struct {
        To      string
        Subject string
        Body    string
    }
    ShouldFail bool
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
    if m.ShouldFail {
        return errors.New("email service unavailable")
    }
    
    m.SentEmails = append(m.SentEmails, struct {
        To      string
        Subject string
        Body    string
    }{
        To:      to,
        Subject: subject,
        Body:    body,
    })
    
    return nil
}

func notifyUserHandler(emailService EmailService) zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        var req struct {
            Email   string `json:"email"`
            Message string `json:"message"`
        }
        
        if err := ctx.BindJSON(&req); err != nil {
            ctx.JSON(http.StatusBadRequest, map[string]string{
                "error": "Invalid JSON",
            })
            return
        }
        
        err := emailService.SendEmail(req.Email, "Notification", req.Message)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Failed to send email",
            })
            return
        }
        
        ctx.JSON(http.StatusOK, map[string]string{
            "message": "Email sent successfully",
        })
    }
}

func TestNotifyUserHandler(t *testing.T) {
    mockEmailService := &MockEmailService{}
    app := zoox.New()
    app.Post("/notify", notifyUserHandler(mockEmailService))
    
    payload := map[string]string{
        "email":   "user@example.com",
        "message": "Test notification",
    }
    
    var body bytes.Buffer
    json.NewEncoder(&body).Encode(payload)
    
    req := httptest.NewRequest(http.MethodPost, "/notify", &body)
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    app.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Len(t, mockEmailService.SentEmails, 1)
    assert.Equal(t, "user@example.com", mockEmailService.SentEmails[0].To)
    assert.Equal(t, "Notification", mockEmailService.SentEmails[0].Subject)
    assert.Equal(t, "Test notification", mockEmailService.SentEmails[0].Body)
}

func TestNotifyUserHandler_EmailServiceFailure(t *testing.T) {
    mockEmailService := &MockEmailService{ShouldFail: true}
    app := zoox.New()
    app.Post("/notify", notifyUserHandler(mockEmailService))
    
    payload := map[string]string{
        "email":   "user@example.com",
        "message": "Test notification",
    }
    
    var body bytes.Buffer
    json.NewEncoder(&body).Encode(payload)
    
    req := httptest.NewRequest(http.MethodPost, "/notify", &body)
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    app.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusInternalServerError, w.Code)
    assert.Len(t, mockEmailService.SentEmails, 0)
}
```

## Performance Testing

```go
func BenchmarkCreateUser(b *testing.B) {
    service := NewUserService()
    app := zoox.New()
    app.Post("/users", createUserHandler(service))
    
    payload := map[string]string{
        "name":  "John Doe",
        "email": "john@example.com",
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        var body bytes.Buffer
        json.NewEncoder(&body).Encode(payload)
        
        req := httptest.NewRequest(http.MethodPost, "/users", &body)
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        app.ServeHTTP(w, req)
        
        if w.Code != http.StatusCreated {
            b.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
        }
    }
}

func BenchmarkGetUser(b *testing.B) {
    service := NewUserService()
    user := service.CreateUser("John Doe", "john@example.com")
    
    app := zoox.New()
    app.Get("/users/:id", getUserHandler(service))
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", user.ID), nil)
        w := httptest.NewRecorder()
        
        app.ServeHTTP(w, req)
        
        if w.Code != http.StatusOK {
            b.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
        }
    }
}
```

## Test Coverage

```bash
# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Set coverage threshold
go test -cover ./... | grep -E "coverage: [0-9]+\.[0-9]+% of statements"
```

## Hands-on Exercise: Complete Test Suite

Create a comprehensive test suite for a blog API with the following requirements:

1. **Unit Tests**: Test all business logic functions
2. **Integration Tests**: Test all API endpoints
3. **Middleware Tests**: Test authentication and authorization
4. **Mock Tests**: Mock external dependencies (database, email service)
5. **Performance Tests**: Benchmark critical operations
6. **Coverage**: Achieve >80% test coverage

## Key Testing Principles

1. **Test Pyramid**: More unit tests, fewer integration tests, minimal E2E tests
2. **Test Independence**: Each test should be independent and repeatable
3. **Clear Naming**: Test names should clearly describe what is being tested
4. **Arrange-Act-Assert**: Structure tests with clear setup, execution, and verification
5. **Mock External Dependencies**: Use mocks to isolate units under test
6. **Test Edge Cases**: Include tests for error conditions and boundary values

## Next Steps

- Tutorial 15: Performance Optimization - Learn optimization techniques
- Tutorial 16: Security Best Practices - Implement security measures
- Explore advanced testing patterns
- Learn about contract testing
- Practice with test-driven development (TDD)

## Additional Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Framework](https://github.com/stretchr/testify)
- [Go Test Coverage](https://golang.org/doc/tutorial/add-a-test)
- [Testing Best Practices](https://golang.org/doc/effective_go.html#testing) 