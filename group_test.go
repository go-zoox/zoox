package zoox

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestGroupMatchPath(t *testing.T) {
	testcases := []struct {
		name     string
		prefix   string
		path     string
		expected bool
	}{
		// 基本匹配测试
		{
			name:     "root path matches root prefix",
			prefix:   "/",
			path:     "/",
			expected: true,
		},
		{
			name:     "empty prefix matches all",
			prefix:   "",
			path:     "/api/users",
			expected: true,
		},
		{
			name:     "exact match",
			prefix:   "/api",
			path:     "/api",
			expected: true,
		},
		{
			name:     "prefix match with trailing slash",
			prefix:   "/api",
			path:     "/api/users",
			expected: true,
		},
		{
			name:     "prefix match with trailing slash in prefix",
			prefix:   "/api/",
			path:     "/api/users",
			expected: true,
		},
		// 边界测试
		{
			name:     "should not match similar prefix",
			prefix:   "/api",
			path:     "/apiv2",
			expected: false,
		},
		{
			name:     "should not match partial prefix",
			prefix:   "/api",
			path:     "/ap",
			expected: false,
		},
		{
			name:     "should not match different path",
			prefix:   "/api",
			path:     "/users",
			expected: false,
		},
		// 动态参数测试
		{
			name:     "match dynamic parameter with colon",
			prefix:   "/users/:id",
			path:     "/users/123",
			expected: true,
		},
		{
			name:     "match dynamic parameter with braces",
			prefix:   "/users/{id}",
			path:     "/users/123",
			expected: true,
		},
		{
			name:     "match nested dynamic parameters",
			prefix:   "/users/:id/posts/:pid",
			path:     "/users/123/posts/456",
			expected: true,
		},
		{
			name:     "match wildcard",
			prefix:   "/files/*path",
			path:     "/files/docs/readme.txt",
			expected: true,
		},
		// 复杂情况测试
		{
			name:     "dynamic parameter with additional path",
			prefix:   "/api/v1/users/:id",
			path:     "/api/v1/users/123/profile",
			expected: true,
		},
		{
			name:     "should not match wrong dynamic path",
			prefix:   "/users/:id",
			path:     "/posts/123",
			expected: false,
		},
		// 原有测试用例
		{
			name:     "original test case 1",
			prefix:   "/",
			path:     "/",
			expected: true,
		},
		{
			name:     "original test case 2",
			prefix:   "/api",
			path:     "/",
			expected: false,
		},
		{
			name:     "original test case 3",
			prefix:   "/",
			path:     "/api",
			expected: true,
		},
		{
			name:     "original test case 4",
			prefix:   "/v1/containers/:id",
			path:     "/v1/containers/d0ac6213f33620362e59cc1b855658f9792377335087c2f3ba1d43639466dd8a/terminal",
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := New()
			group := &RouterGroup{
				app:    app,
				prefix: tc.prefix,
			}

			result := group.matchPath(tc.path)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for prefix '%s' and path '%s'", 
					tc.expected, result, tc.prefix, tc.path)
			}
		})
	}
}

func TestGroupMiddlewareInheritance(t *testing.T) {
	app := New()
	
	// 记录中间件执行顺序
	var executionOrder []string
	
	// 根级中间件
	app.Use(func(ctx *Context) {
		executionOrder = append(executionOrder, "root")
		ctx.Next()
	})
	
	// 一级 group
	v1 := app.Group("/v1", func(g *RouterGroup) {
		g.Use(func(ctx *Context) {
			executionOrder = append(executionOrder, "v1")
			ctx.Next()
		})
	})
	
	// 二级 group
	v1.Group("/users", func(g *RouterGroup) {
		g.Use(func(ctx *Context) {
			executionOrder = append(executionOrder, "users")
			ctx.Next()
		})
		
		g.Get("/:id", func(ctx *Context) {
			executionOrder = append(executionOrder, "handler")
			ctx.JSON(200, map[string]string{"id": ctx.Param().Get("id").String()})
		})
	})
	
	// 测试请求
	req := httptest.NewRequest("GET", "/v1/users/123", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// 验证中间件执行顺序
	expectedOrder := []string{"root", "v1", "users", "handler"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d middleware calls, got %d", len(expectedOrder), len(executionOrder))
	}
	
	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected middleware %d to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
	
	// 验证响应
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
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
	
	// 创建嵌套的路由组
	api := app.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")
	
	users.Get("/:id", func(ctx *Context) {
		ctx.JSON(200, map[string]string{
			"id":   ctx.Param().Get("id").String(),
			"path": ctx.Path,
		})
	})
	
	// 测试请求
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
	
	// 验证响应内容
	body := w.Body.String()
	if !strings.Contains(body, `"id":"123"`) {
		t.Errorf("Expected response to contain id=123, got: %s", body)
	}
}

func TestGroupMiddlewareOrder(t *testing.T) {
	app := New()
	
	var order []string
	
	// 根级中间件
	app.Use(func(ctx *Context) {
		order = append(order, "global-1")
		ctx.Next()
	})
	
	app.Use(func(ctx *Context) {
		order = append(order, "global-2")
		ctx.Next()
	})
	
	// Group 中间件
	api := app.Group("/api")
	api.Use(func(ctx *Context) {
		order = append(order, "api-1")
		ctx.Next()
	})
	
	api.Use(func(ctx *Context) {
		order = append(order, "api-2")
		ctx.Next()
	})
	
	// 子 Group 中间件
	v1 := api.Group("/v1")
	v1.Use(func(ctx *Context) {
		order = append(order, "v1-1")
		ctx.Next()
	})
	
	v1.Get("/test", func(ctx *Context) {
		order = append(order, "handler")
		ctx.JSON(200, map[string]interface{}{"status": "ok"})
	})
	
	// 测试请求
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// 验证中间件执行顺序
	expectedOrder := []string{"global-1", "global-2", "api-1", "api-2", "v1-1", "handler"}
	
	if len(order) != len(expectedOrder) {
		t.Fatalf("Expected %d middleware calls, got %d: %v", len(expectedOrder), len(order), order)
	}
	
	for i, expected := range expectedOrder {
		if order[i] != expected {
			t.Errorf("Expected middleware %d to be '%s', got '%s'", i, expected, order[i])
		}
	}
}

func TestGroupConflictResolution(t *testing.T) {
	app := New()
	
	var executedHandlers []string
	
	// 创建两个可能冲突的组
	app.Group("/api", func(g *RouterGroup) {
		g.Use(func(ctx *Context) {
			executedHandlers = append(executedHandlers, "api-middleware")
			ctx.Next()
		})
	})
	
	app.Group("/api/v1", func(g *RouterGroup) {
		g.Use(func(ctx *Context) {
			executedHandlers = append(executedHandlers, "api-v1-middleware")
			ctx.Next()
		})
		
		g.Get("/users", func(ctx *Context) {
			executedHandlers = append(executedHandlers, "api-v1-handler")
			ctx.JSON(200, map[string]string{"message": "success"})
		})
	})
	
	// 测试请求 - 应该匹配更具体的路径
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
	
	// 验证只有最具体的匹配被执行
	expectedHandlers := []string{"api-v1-middleware", "api-v1-handler"}
	
	if len(executedHandlers) != len(expectedHandlers) {
		t.Fatalf("Expected %d handlers, got %d: %v", len(expectedHandlers), len(executedHandlers), executedHandlers)
	}
	
	for i, expected := range expectedHandlers {
		if executedHandlers[i] != expected {
			t.Errorf("Expected handler %d to be '%s', got '%s'", i, expected, executedHandlers[i])
		}
	}
}

func BenchmarkGroupMatchPath(b *testing.B) {
	app := New()
	group := &RouterGroup{
		app:    app,
		prefix: "/api/v1/users/:id",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group.matchPath("/api/v1/users/123")
	}
}

func BenchmarkGroupMiddlewareCollection(b *testing.B) {
	app := New()
	
	// 创建深层嵌套的组
	g1 := app.Group("/api")
	g1.Use(func(ctx *Context) { ctx.Next() })
	
	g2 := g1.Group("/v1")
	g2.Use(func(ctx *Context) { ctx.Next() })
	
	g3 := g2.Group("/users")
	g3.Use(func(ctx *Context) { ctx.Next() })
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g3.getAllMiddlewares()
	}
}

func TestGroupBuildRegexPattern(t *testing.T) {
	app := New()
	group := &RouterGroup{
		app: app,
	}

	testcases := []struct {
		prefix   string
		expected string
	}{
		{"/users/:id", `/users/([^/]+)`},
		{"/users/{id}", `/users/([^/]+)`},
		{"/files/*path", `/files/(.*)`},
		{"/users/:id/posts/:pid", `/users/([^/]+)/posts/([^/]+)`},
	}

	for _, tc := range testcases {
		t.Run(tc.prefix, func(t *testing.T) {
			result := group.buildRegexPattern(tc.prefix)
			t.Logf("Input: %s, Output: %s, Expected: %s", tc.prefix, result, tc.expected)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
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
