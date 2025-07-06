package zoox

import (
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestGroupMatchPath(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		path     string
		expected bool
	}{
		// Basic matching tests
		{"root path matches root prefix", "/", "/", true},
		{"empty prefix matches all", "", "/users", true},
		{"exact match", "/api", "/api", true},
		{"prefix match", "/api", "/api/users", true},
		{"prefix match with trailing slash in prefix", "/api/", "/api/users", true},
		{"should not match different prefix", "/api", "/users", false},
		{"should not match partial prefix", "/api", "/ap", false},

		// Original test cases
		{"original test case 1", "/v1", "/v1/users", true},
		{"original test case 2", "/v1", "/v1", true},
		{"original test case 3", "/v1", "/v2", false},
		{"original test case 4", "/v1", "/v1users", true}, // This will match with simple prefix matching
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &RouterGroup{prefix: tt.prefix}
			result := group.matchPath(tt.path)
			if result != tt.expected {
				t.Errorf("matchPath(%q, %q) = %v, want %v", tt.prefix, tt.path, result, tt.expected)
			}
		})
	}
}

func TestGroupMiddlewareInheritance(t *testing.T) {
	app := New()

	// Record middleware execution order
	var executionOrder []string

	// Root level middleware
	app.Use(func(ctx *Context) {
		executionOrder = append(executionOrder, "root")
		ctx.Next()
	})

	// First level group
	v1 := app.Group("/v1")
	v1.Use(func(ctx *Context) {
		executionOrder = append(executionOrder, "v1")
		ctx.Next()
	})

	// Second level group
	users := v1.Group("/users")
	users.Use(func(ctx *Context) {
		executionOrder = append(executionOrder, "users")
		ctx.Next()
	})

	users.Get("/:id", func(ctx *Context) {
		executionOrder = append(executionOrder, "handler")
		ctx.String(200, "user")
	})

	// Test request
	req := httptest.NewRequest("GET", "/v1/users/123", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// Verify middleware execution order
	expected := []string{"root", "v1", "users", "handler"}
	if len(executionOrder) != len(expected) {
		t.Errorf("Expected %d middleware executions, got %d", len(expected), len(executionOrder))
	}

	for i, middleware := range expected {
		if i >= len(executionOrder) || executionOrder[i] != middleware {
			t.Errorf("Expected middleware %s at position %d, got %s", middleware, i, executionOrder[i])
		}
	}

	// Verify response
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "user" {
		t.Errorf("Expected body 'user', got '%s'", w.Body.String())
	}
}

func TestGroupPathJoining(t *testing.T) {
	testcases := []struct {
		name     string
		prefix   string
		path     string
		expected string
	}{
		{
			name:     "simple join",
			prefix:   "/api",
			path:     "/users",
			expected: "/api/users",
		},
		{
			name:     "join with trailing slash in prefix",
			prefix:   "/api/",
			path:     "/users",
			expected: "/api/users",
		},
		{
			name:     "join with leading slash in path",
			prefix:   "/api",
			path:     "/users",
			expected: "/api/users",
		},
		{
			name:     "join with both slashes",
			prefix:   "/api/",
			path:     "/users",
			expected: "/api/users",
		},
		{
			name:     "empty prefix",
			prefix:   "",
			path:     "/users",
			expected: "/users",
		},
		{
			name:     "empty path",
			prefix:   "/api",
			path:     "",
			expected: "/api",
		},
		{
			name:     "root paths",
			prefix:   "/",
			path:     "/",
			expected: "/",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := New()
			group := &RouterGroup{
				app:    app,
				prefix: tc.prefix,
			}

			result := group.joinPath(tc.path)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestGroupNestedRouting(t *testing.T) {
	app := New()

	// Create nested routing groups
	api := app.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")

	users.Get("/:id", func(ctx *Context) {
		ctx.String(200, "user-"+ctx.Param().Get("id").String())
	})

	// Test request
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response content
	expected := "user-123"
	if w.Body.String() != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, w.Body.String())
	}
}

func TestGroupMiddlewareOrder(t *testing.T) {
	app := New()
	var order []string

	// Root level middleware
	app.Use(func(ctx *Context) {
		order = append(order, "global")
		ctx.Next()
	})

	// Group middleware
	api := app.Group("/api")
	api.Use(func(ctx *Context) {
		order = append(order, "api")
		ctx.Next()
	})

	v1 := api.Group("/v1")
	v1.Use(func(ctx *Context) {
		order = append(order, "v1")
		ctx.Next()
	})

	// Sub-group middleware
	v1.Get("/test", func(ctx *Context) {
		order = append(order, "handler")
		ctx.String(200, "ok")
	})

	// Test request
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// Verify middleware execution order
	expected := []string{"global", "api", "v1", "handler"}
	if len(order) != len(expected) {
		t.Errorf("Expected %d middleware executions, got %d", len(expected), len(order))
	}

	for i, middleware := range expected {
		if i >= len(order) || order[i] != middleware {
			t.Errorf("Expected middleware %s at position %d, got %s", middleware, i, order[i])
		}
	}
}

func TestGroupConflictResolution(t *testing.T) {
	app := New()

	// Create two potentially conflicting groups
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// More specific group
	users := v1.Group("/users")
	users.Get("/list", func(ctx *Context) {
		ctx.String(200, "users-list")
	})

	// Less specific group with different path
	v1.Get("/info", func(ctx *Context) {
		ctx.String(200, "v1-info")
	})

	// Test request - should match more specific path
	req := httptest.NewRequest("GET", "/api/v1/users/list", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify only the most specific match is executed
	expected := "users-list"
	if w.Body.String() != expected {
		t.Errorf("Expected body '%s', got '%s'", expected, w.Body.String())
	}
}

func BenchmarkGroupMatchPath(b *testing.B) {
	group := &RouterGroup{prefix: "/api/v1/users/:id"}
	path := "/api/v1/users/123/profile"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group.matchPath(path)
	}
}

func BenchmarkGroupMiddlewareCollection(b *testing.B) {
	app := New()

	// Create deeply nested groups
	api := app.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")
	profile := users.Group("/profile")

	// Add middlewares at each level
	api.Use(func(ctx *Context) { ctx.Next() })
	v1.Use(func(ctx *Context) { ctx.Next() })
	users.Use(func(ctx *Context) { ctx.Next() })
	profile.Use(func(ctx *Context) { ctx.Next() })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profile.getAllMiddlewares()
	}
}

func TestRegexpQuoteMeta(t *testing.T) {
	testcases := []string{
		"/users/:id",
		"/users/{id}",
		"/files/*path",
	}

	for _, tc := range testcases {
		quoted := regexp.QuoteMeta(tc)
		t.Logf("Input: %s, QuoteMeta: %s", tc, quoted)
	}
}
