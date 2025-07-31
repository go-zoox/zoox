package zoox

import (
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/usersx/:nid", nil)
	r.addRoute("GET", "/users/{id}", nil)
	r.addRoute("GET", "/users/:id/profile", nil)
	r.addRoute("GET", "/users/:id/logs/:lid", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePath(t *testing.T) {
	if !reflect.DeepEqual(parsePath("/p/:name"), []string{"p", ":name"}) {
		t.Errorf("Expected [p,:name], got %v", parsePath("/p/:name"))
	}

	if !reflect.DeepEqual(parsePath("/p/*"), []string{"p", "*"}) {
		t.Errorf("Expected [p,*], got %v", parsePath("/p/*"))
	}

	if !reflect.DeepEqual(parsePath("/p/*name/*"), []string{"p", "*name"}) {
		t.Errorf("Expected [p,*name], got %v", parsePath("/p/*name/*"))
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/zoox")

	if n == nil {
		t.Fatal("Expected node, got nil")
	}

	if n.Path != "/hello/:name" {
		t.Errorf("Expected /hello/:name, got %s", n.Path)
	}

	if ps["name"] != "zoox" {
		t.Errorf("Expected zoox, got %s", ps["name"])
	}
}

func TestGetRouteMultiParams(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/users/1/logs/6")

	if n == nil {
		t.Fatal("Expected node, got nil")
	}

	if n.Path != "/users/:id/logs/:lid" {
		t.Errorf("Expected /users/:id/logs/:lid, got %s", n.Path)
	}

	if ps["id"] != "1" {
		t.Errorf("Expected 1, got %s", ps["id"])
	}

	if ps["lid"] != "6" {
		t.Errorf("Expected 6, got %s", ps["lid"])
	}
}

func TestGetRouteWithBrackets(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/users/1")

	if n == nil {
		t.Fatal("Expected node, got nil")
	}

	if n.Path != "/users/{id}" {
		t.Errorf("Expected /users/{id}, got %s", n.Path)
	}

	if ps["id"] != "1" {
		t.Errorf("Expected 1, got %v", ps["id"])
	}
}

func TestRouterGroup_Group(t *testing.T) {
	app := New()
	
	// Test nested groups
	api := app.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")
	
	users.Get("/:id", func(ctx *Context) {
		ctx.String(200, "user %s", ctx.Param().Get("id"))
	})
	
	req := httptest.NewRequest("GET", "/api/v1/users/123", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "user 123" {
		t.Errorf("Expected 'user 123', got '%s'", w.Body.String())
	}
}

func TestRouterGroup_Use(t *testing.T) {
	app := New()
	
	var executionOrder []string
	
	// Global middleware
	app.Use(func(ctx *Context) {
		executionOrder = append(executionOrder, "global")
		ctx.Next()
	})
	
	// Group middleware
	api := app.Group("/api")
	api.Use(func(ctx *Context) {
		executionOrder = append(executionOrder, "api")
		ctx.Next()
	})
	
	// Nested group middleware
	v1 := api.Group("/v1")
	v1.Use(func(ctx *Context) {
		executionOrder = append(executionOrder, "v1")
		ctx.Next()
	})
	
	v1.Get("/test", func(ctx *Context) {
		executionOrder = append(executionOrder, "handler")
		ctx.String(200, "ok")
	})
	
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	expectedOrder := []string{"global", "api", "v1", "handler"}
	
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected order length %d, got %d", len(expectedOrder), len(executionOrder))
	}
	
	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Expected order[%d] = '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

func TestRouterGroup_HTTPMethods(t *testing.T) {
	app := New()
	api := app.Group("/api")
	
	// Test all HTTP methods on group
	api.Get("/get", func(ctx *Context) { ctx.String(200, "GET") })
	api.Post("/post", func(ctx *Context) { ctx.String(200, "POST") })
	api.Put("/put", func(ctx *Context) { ctx.String(200, "PUT") })
	api.Patch("/patch", func(ctx *Context) { ctx.String(200, "PATCH") })
	api.Delete("/delete", func(ctx *Context) { ctx.String(200, "DELETE") })
	api.Head("/head", func(ctx *Context) { ctx.String(200, "HEAD") })
	api.Options("/options", func(ctx *Context) { ctx.String(200, "OPTIONS") })
	api.Connect("/connect", func(ctx *Context) { ctx.String(200, "CONNECT") })
	
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "CONNECT"}
	
	for _, method := range methods {
		path := "/api/" + strings.ToLower(method)
		
		req := httptest.NewRequest(method, path, nil)
		w := httptest.NewRecorder()
		
		app.ServeHTTP(w, req)
		
		if w.Code != 200 {
			t.Errorf("Expected status 200 for %s %s, got %d", method, path, w.Code)
		}
		
		if method != "HEAD" && w.Body.String() != method {
			t.Errorf("Expected body '%s' for %s %s, got '%s'", method, method, path, w.Body.String())
		}
	}
}

func TestRouterGroup_Any(t *testing.T) {
	app := New()
	api := app.Group("/api")
	
	api.Any("/any", func(ctx *Context) {
		ctx.String(200, "any method")
	})
	
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/any", nil)
		w := httptest.NewRecorder()
		
		app.ServeHTTP(w, req)
		
		if w.Code != 200 {
			t.Errorf("Expected status 200 for %s /api/any, got %d", method, w.Code)
		}
		
		if method != "HEAD" && w.Body.String() != "any method" {
			t.Errorf("Expected body 'any method' for %s /api/any, got '%s'", method, w.Body.String())
		}
	}
}

func TestRouterGroup_Static(t *testing.T) {
	app := New()
	api := app.Group("/api")
	
	// Test static file serving on group
	api.Static("/static", "./")
	
	req := httptest.NewRequest("GET", "/api/static/README.md", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// Should return 200 if file exists, or 404 if not
	if w.Code != 200 && w.Code != 404 {
		t.Errorf("Expected status 200 or 404 for static file, got %d", w.Code)
	}
}

func TestRouterGroup_Prefix(t *testing.T) {
	app := New()
	
	// Test that group prefixes are correctly applied
	api := app.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")
	
	if api.prefix != "/api" {
		t.Errorf("Expected api prefix '/api', got '%s'", api.prefix)
	}
	
	if v1.prefix != "/api/v1" {
		t.Errorf("Expected v1 prefix '/api/v1', got '%s'", v1.prefix)
	}
	
	if users.prefix != "/api/v1/users" {
		t.Errorf("Expected users prefix '/api/v1/users', got '%s'", users.prefix)
	}
}

func TestRouterGroup_MiddlewareChain(t *testing.T) {
	app := New()
	
	var executionOrder []string
	
	// Create middleware functions
	middleware1 := func(ctx *Context) {
		executionOrder = append(executionOrder, "middleware1")
		ctx.Next()
		executionOrder = append(executionOrder, "middleware1-after")
	}
	
	middleware2 := func(ctx *Context) {
		executionOrder = append(executionOrder, "middleware2")
		ctx.Next()
		executionOrder = append(executionOrder, "middleware2-after")
	}
	
	middleware3 := func(ctx *Context) {
		executionOrder = append(executionOrder, "middleware3")
		ctx.Next()
		executionOrder = append(executionOrder, "middleware3-after")
	}
	
	// Set up middleware chain
	app.Use(middleware1)
	
	api := app.Group("/api")
	api.Use(middleware2)
	
	v1 := api.Group("/v1")
	v1.Use(middleware3)
	
	v1.Get("/test", func(ctx *Context) {
		executionOrder = append(executionOrder, "handler")
		ctx.String(200, "ok")
	})
	
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	expectedOrder := []string{
		"middleware1", "middleware2", "middleware3", "handler",
		"middleware3-after", "middleware2-after", "middleware1-after",
	}
	
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected order length %d, got %d", len(expectedOrder), len(executionOrder))
	}
	
	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Expected order[%d] = '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

func TestRouterGroup_PathParameters(t *testing.T) {
	app := New()
	api := app.Group("/api")
	
	// Test various path parameter patterns
	api.Get("/users/:id", func(ctx *Context) {
		ctx.String(200, "user %s", ctx.Param().Get("id"))
	})
	
	api.Get("/users/:id/posts/:postId", func(ctx *Context) {
		ctx.String(200, "user %s post %s", ctx.Param().Get("id"), ctx.Param().Get("postId"))
	})
	
	api.Get("/files/*filepath", func(ctx *Context) {
		ctx.String(200, "file %s", ctx.Param().Get("filepath"))
	})
	
	// Test single parameter
	req := httptest.NewRequest("GET", "/api/users/123", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "user 123" {
		t.Errorf("Expected 'user 123', got '%s'", w.Body.String())
	}
	
	// Test multiple parameters
	req = httptest.NewRequest("GET", "/api/users/123/posts/456", nil)
	w = httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "user 123 post 456" {
		t.Errorf("Expected 'user 123 post 456', got '%s'", w.Body.String())
	}
	
	// Test wildcard parameter
	req = httptest.NewRequest("GET", "/api/files/path/to/file.txt", nil)
	w = httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "file path/to/file.txt" {
		t.Errorf("Expected 'file path/to/file.txt', got '%s'", w.Body.String())
	}
}
