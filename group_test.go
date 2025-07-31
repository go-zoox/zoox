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

func TestGroupDynamicRouting(t *testing.T) {
	app := New()

	// 测试动态路由组
	userGroup := app.Group("/:id/profile")
	userGroup.Get("/settings", func(ctx *Context) {
		ctx.String(200, "user-settings-"+ctx.Param().Get("id").String())
	})

	categoryGroup := app.Group("/:id/xxx/:cat")
	categoryGroup.Get("/details", func(ctx *Context) {
		id := ctx.Param().Get("id").String()
		cat := ctx.Param().Get("cat").String()
		ctx.String(200, "category-"+id+"-"+cat)
	})

	// 测试请求
	tests := []struct {
		path     string
		expected string
		status   int
	}{
		{"/123/profile/settings", "user-settings-123", 200},
		{"/456/xxx/books/details", "category-456-books", 200},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}

			if w.Body.String() != tt.expected {
				t.Errorf("Expected body '%s', got '%s'", tt.expected, w.Body.String())
			}
		})
	}
}

func TestGroupDynamicRoutingDebug(t *testing.T) {
	app := New()

	// 测试动态路由组
	userGroup := app.Group("/:id/profile")
	userGroup.Get("/settings", func(ctx *Context) {
		ctx.String(200, "user-settings-"+ctx.Param().Get("id").String())
	})

	categoryGroup := app.Group("/:id/xxx/:cat")
	categoryGroup.Get("/details", func(ctx *Context) {
		id := ctx.Param().Get("id").String()
		cat := ctx.Param().Get("cat").String()
		ctx.String(200, "category-"+id+"-"+cat)
	})

	// 打印所有路由组
	t.Logf("Total groups: %d", len(app.groups))
	for i, group := range app.groups {
		t.Logf("Group %d: prefix='%s'", i, group.prefix)
	}

	// 测试路径匹配
	testPaths := []string{"/123/profile/settings", "/456/xxx/books/details"}
	for _, path := range testPaths {
		t.Logf("\nTesting path: %s", path)
		for i, group := range app.groups {
			matches := group.matchPath(path)
			t.Logf("  Group %d ('%s') matches: %v", i, group.prefix, matches)
		}
	}
}

func TestGroupDynamicPathMatching(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		path     string
		expected bool
	}{
		// 静态路径匹配
		{"static exact match", "/api/v1", "/api/v1", true},
		{"static prefix match", "/api/v1", "/api/v1/users", true},
		{"static no match", "/api/v1", "/api/v2", false},

		// 动态路径匹配
		{"single param", "/:id", "/123", true},
		{"single param with static", "/:id/profile", "/123/profile", true},
		{"single param with static and more", "/:id/profile", "/123/profile/settings", true},
		{"single param no match", "/:id/profile", "/123/settings", false},

		// 多个参数
		{"multiple params", "/:id/xxx/:cat", "/123/xxx/books", true},
		{"multiple params with more", "/:id/xxx/:cat", "/123/xxx/books/details", true},
		{"multiple params no match", "/:id/xxx/:cat", "/123/yyy/books", false},

		// 混合格式
		{"bracket format", "/{id}/profile", "/123/profile", true},
		{"bracket format no match", "/{id}/profile", "/123/settings", false},

		// 通配符
		{"wildcard", "/files/*path", "/files/docs/readme.txt", true},
		{"wildcard root", "/*path", "/anything/goes/here", true},

		// 边界情况
		{"empty parts", "/:id//profile", "/123//profile", true},
		{"trailing slash in prefix", "/:id/profile/", "/123/profile/settings", true},
		{"insufficient path parts", "/:id/profile/settings", "/123/profile", false},
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

func TestGroupDynamicRoutingWithMiddleware(t *testing.T) {
	app := New()

	var middlewareOrder []string

	// 全局中间件
	app.Use(func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "global")
		ctx.Next()
	})

	// 动态路由组 - 用户相关
	userGroup := app.Group("/:id/profile")
	userGroup.Use(func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "user-profile")
		ctx.Next()
	})
	userGroup.Get("/settings", func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "handler")
		ctx.String(200, "user-settings-"+ctx.Param().Get("id").String())
	})

	// 动态路由组 - 分类相关
	categoryGroup := app.Group("/:id/xxx/:cat")
	categoryGroup.Use(func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "category")
		ctx.Next()
	})
	categoryGroup.Get("/details", func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "handler")
		id := ctx.Param().Get("id").String()
		cat := ctx.Param().Get("cat").String()
		ctx.String(200, "category-"+id+"-"+cat)
	})

	// 测试用户路由组中间件
	t.Run("user profile middleware", func(t *testing.T) {
		middlewareOrder = nil // 重置
		req := httptest.NewRequest("GET", "/123/profile/settings", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		// 验证中间件执行顺序
		expected := []string{"global", "user-profile", "handler"}
		if len(middlewareOrder) != len(expected) {
			t.Errorf("Expected %d middleware executions, got %d", len(expected), len(middlewareOrder))
		}

		for i, middleware := range expected {
			if i >= len(middlewareOrder) || middlewareOrder[i] != middleware {
				t.Errorf("Expected middleware %s at position %d, got %s", middleware, i, middlewareOrder[i])
			}
		}

		// 验证响应
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "user-settings-123" {
			t.Errorf("Expected body 'user-settings-123', got '%s'", w.Body.String())
		}
	})

	// 测试分类路由组中间件
	t.Run("category middleware", func(t *testing.T) {
		middlewareOrder = nil // 重置
		req := httptest.NewRequest("GET", "/456/xxx/books/details", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		// 验证中间件执行顺序
		expected := []string{"global", "category", "handler"}
		if len(middlewareOrder) != len(expected) {
			t.Errorf("Expected %d middleware executions, got %d", len(expected), len(middlewareOrder))
		}

		for i, middleware := range expected {
			if i >= len(middlewareOrder) || middlewareOrder[i] != middleware {
				t.Errorf("Expected middleware %s at position %d, got %s", middleware, i, middlewareOrder[i])
			}
		}

		// 验证响应
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "category-456-books" {
			t.Errorf("Expected body 'category-456-books', got '%s'", w.Body.String())
		}
	})
}

func TestNestedDynamicGroupMiddleware(t *testing.T) {
	app := New()

	var middlewareOrder []string

	// 全局中间件
	app.Use(func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "global")
		ctx.Next()
	})

	// 一级动态路由组
	userGroup := app.Group("/:userId")
	userGroup.Use(func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "user-"+ctx.Param().Get("userId").String())
		ctx.Next()
	})

	// 二级动态路由组
	profileGroup := userGroup.Group("/profile/:section")
	profileGroup.Use(func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "profile-"+ctx.Param().Get("section").String())
		ctx.Next()
	})

	// 三级路由
	profileGroup.Get("/details", func(ctx *Context) {
		middlewareOrder = append(middlewareOrder, "handler")
		userId := ctx.Param().Get("userId").String()
		section := ctx.Param().Get("section").String()
		ctx.String(200, "user-"+userId+"-profile-"+section)
	})

	// 测试嵌套动态路由组中间件
	middlewareOrder = nil // 重置
	req := httptest.NewRequest("GET", "/123/profile/settings/details", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// 验证中间件执行顺序
	expected := []string{"global", "user-123", "profile-settings", "handler"}
	if len(middlewareOrder) != len(expected) {
		t.Errorf("Expected %d middleware executions, got %d", len(expected), len(middlewareOrder))
		t.Logf("Actual order: %v", middlewareOrder)
	}

	for i, middleware := range expected {
		if i >= len(middlewareOrder) || middlewareOrder[i] != middleware {
			t.Errorf("Expected middleware %s at position %d, got %s", middleware, i, middlewareOrder[i])
		}
	}

	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "user-123-profile-settings" {
		t.Errorf("Expected body 'user-123-profile-settings', got '%s'", w.Body.String())
	}
}
