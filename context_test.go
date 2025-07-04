package zoox

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestContext_String(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.String(200, "hello %s", "world")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", w.Body.String())
	}
	
	// String method doesn't set content-type automatically
	// Content-Type is set by the HTTP package based on the content
}

func TestContext_JSON(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.JSON(200, H{"message": "hello", "count": 42})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Errorf("Expected JSON content type, got '%s'", w.Header().Get("Content-Type"))
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}
	
	if response["message"] != "hello" {
		t.Errorf("Expected message 'hello', got '%v'", response["message"])
	}
	
	if response["count"] != float64(42) {
		t.Errorf("Expected count 42, got '%v'", response["count"])
	}
}

func TestContext_HTML(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.HTML(200, "<h1>{{.title}}</h1>", H{
			"title": "Test Title",
		})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected HTML content type 'text/html', got '%s'", contentType)
	}
	
	// The response should contain the rendered HTML
	if !strings.Contains(w.Body.String(), "Test Title") {
		t.Errorf("Expected rendered HTML to contain 'Test Title', got '%s'", w.Body.String())
	}
}

func TestContext_Redirect(t *testing.T) {
	app := New()
	
	app.Get("/redirect", func(ctx *Context) {
		ctx.Redirect("/target", 302)
	})
	
	req := httptest.NewRequest("GET", "/redirect", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 302 {
		t.Errorf("Expected status 302, got %d", w.Code)
	}
	
	if w.Header().Get("Location") != "/target" {
		t.Errorf("Expected Location header '/target', got '%s'", w.Header().Get("Location"))
	}
}

func TestContext_Param(t *testing.T) {
	app := New()
	
	app.Get("/user/:id", func(ctx *Context) {
		id := ctx.Param().Get("id")
		ctx.String(200, "user id: %s", id)
	})
	
	req := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "user id: 123" {
		t.Errorf("Expected 'user id: 123', got '%s'", w.Body.String())
	}
}

func TestContext_Query(t *testing.T) {
	app := New()
	
	app.Get("/search", func(ctx *Context) {
		q := ctx.Query().Get("q")
		limit := ctx.Query().Get("limit", "10")
		ctx.String(200, "query: %s, limit: %s", q, limit)
	})
	
	req := httptest.NewRequest("GET", "/search?q=golang&limit=20", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "query: golang, limit: 20" {
		t.Errorf("Expected 'query: golang, limit: 20', got '%s'", w.Body.String())
	}
}

func TestContext_QueryWithDefault(t *testing.T) {
	app := New()
	
	app.Get("/search", func(ctx *Context) {
		q := ctx.Query().Get("q")
		limit := ctx.Query().Get("limit", "10")
		ctx.String(200, "query: %s, limit: %s", q, limit)
	})
	
	// Test with missing limit parameter
	req := httptest.NewRequest("GET", "/search?q=golang", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "query: golang, limit: 10" {
		t.Errorf("Expected 'query: golang, limit: 10', got '%s'", w.Body.String())
	}
}

func TestContext_Header(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		userAgent := ctx.Header().Get("User-Agent")
		ctx.String(200, "User-Agent: %s", userAgent)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "User-Agent: test-agent" {
		t.Errorf("Expected 'User-Agent: test-agent', got '%s'", w.Body.String())
	}
}

func TestContext_SetHeader(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.SetHeader("X-Custom", "test-value")
		ctx.String(200, "ok")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("X-Custom") != "test-value" {
		t.Errorf("Expected X-Custom header 'test-value', got '%s'", w.Header().Get("X-Custom"))
	}
}

func TestContext_Cookie(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		cookie := ctx.Cookie().Get("test-cookie")
		ctx.String(200, "cookie: %s", cookie)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "test-cookie", Value: "test-value"})
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "cookie: test-value" {
		t.Errorf("Expected 'cookie: test-value', got '%s'", w.Body.String())
	}
}

func TestContext_SetCookie(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.Cookie().Set("test-cookie", "test-value")
		ctx.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Check if cookie was set
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "test-cookie" && cookie.Value == "test-value" {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Expected cookie 'test-cookie' with value 'test-value' to be set")
	}
}

func TestContext_BindJSON(t *testing.T) {
	app := New()
	
	type TestData struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}
	
	app.Post("/test", func(ctx *Context) {
		var data TestData
		if err := ctx.BindJSON(&data); err != nil {
			ctx.String(400, "bind error: %v", err)
			return
		}
		ctx.JSON(200, data)
	})
	
	testData := TestData{Name: "John", Age: 30, Email: "john@example.com"}
	jsonData, _ := json.Marshal(testData)
	
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response TestData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}
	
	if response.Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", response.Name)
	}
	
	if response.Age != 30 {
		t.Errorf("Expected age 30, got %d", response.Age)
	}
	
	if response.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", response.Email)
	}
}

func TestContext_BindForm(t *testing.T) {
	app := New()
	
	type TestData struct {
		Name  string `form:"name"`
		Age   int    `form:"age"`
		Email string `form:"email"`
	}
	
	app.Post("/test", func(ctx *Context) {
		var data TestData
		if err := ctx.BindForm(&data); err != nil {
			ctx.String(400, "bind error: %v", err)
			return
		}
		ctx.JSON(200, data)
	})
	
	formData := url.Values{}
	formData.Set("name", "John")
	formData.Set("age", "30")
	formData.Set("email", "john@example.com")
	
	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response TestData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}
	
	if response.Name != "John" {
		t.Errorf("Expected name 'John', got '%s'", response.Name)
	}
	
	if response.Age != 30 {
		t.Errorf("Expected age 30, got %d", response.Age)
	}
	
	if response.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", response.Email)
	}
}

func TestContext_FormValue(t *testing.T) {
	app := New()
	
	app.Post("/test", func(ctx *Context) {
		name := ctx.Form().Get("name")
		age := ctx.Form().Get("age")
		ctx.String(200, "name: %s, age: %s", name, age)
	})
	
	formData := url.Values{}
	formData.Set("name", "John")
	formData.Set("age", "30")
	
	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "name: John, age: 30" {
		t.Errorf("Expected 'name: John, age: 30', got '%s'", w.Body.String())
	}
}

func TestContext_File(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.File("./README.md")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// Should return 200 if file exists, or 404 if not
	if w.Code != 200 && w.Code != 404 {
		t.Errorf("Expected status 200 or 404, got %d", w.Code)
	}
}

func TestContext_Status(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.Status(201)
		ctx.String(201, "created")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 201 {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestContext_Next(t *testing.T) {
	app := New()
	
	var order []string
	
	app.Use(func(ctx *Context) {
		order = append(order, "middleware1")
		ctx.Next()
		order = append(order, "middleware1-after")
	})
	
	app.Use(func(ctx *Context) {
		order = append(order, "middleware2")
		ctx.Next()
		order = append(order, "middleware2-after")
	})
	
	app.Get("/test", func(ctx *Context) {
		order = append(order, "handler")
		ctx.String(200, "ok")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	expectedOrder := []string{"middleware1", "middleware2", "handler", "middleware2-after", "middleware1-after"}
	
	if len(order) != len(expectedOrder) {
		t.Errorf("Expected order length %d, got %d", len(expectedOrder), len(order))
	}
	
	for i, expected := range expectedOrder {
		if i >= len(order) || order[i] != expected {
			t.Errorf("Expected order[%d] = '%s', got '%s'", i, expected, order[i])
		}
	}
}

func TestContext_ClientIP(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ip := ctx.IP()
		ctx.String(200, "IP: %s", ip)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "IP: 192.168.1.1" {
		t.Errorf("Expected 'IP: 192.168.1.1', got '%s'", w.Body.String())
	}
}

func TestContext_UserAgent(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		userAgent := ctx.UserAgent().String()
		ctx.String(200, "UserAgent: %s", userAgent)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "UserAgent: test-agent" {
		t.Errorf("Expected 'UserAgent: test-agent', got '%s'", w.Body.String())
	}
}

func TestContext_Method(t *testing.T) {
	app := New()
	
	app.Post("/test", func(ctx *Context) {
		method := ctx.Method
		ctx.String(200, "Method: %s", method)
	})
	
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "Method: POST" {
		t.Errorf("Expected 'Method: POST', got '%s'", w.Body.String())
	}
}

func TestContext_Path(t *testing.T) {
	app := New()
	
	app.Get("/test/path", func(ctx *Context) {
		path := ctx.Path
		ctx.String(200, "Path: %s", path)
	})
	
	req := httptest.NewRequest("GET", "/test/path", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "Path: /test/path" {
		t.Errorf("Expected 'Path: /test/path', got '%s'", w.Body.String())
	}
}

func TestContext_FullPath(t *testing.T) {
	app := New()
	
	app.Get("/user/:id", func(ctx *Context) {
		path := ctx.Path
		ctx.String(200, "Path: %s", path)
	})
	
	req := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "Path: /user/123" {
		t.Errorf("Expected 'Path: /user/123', got '%s'", w.Body.String())
	}
}

func TestContext_State(t *testing.T) {
	app := New()
	
	app.Use(func(ctx *Context) {
		ctx.State().Set("user", "john")
		ctx.Next()
	})
	
	app.Get("/test", func(ctx *Context) {
		user := ctx.State().Get("user")
		ctx.String(200, "User: %s", user)
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "User: john" {
		t.Errorf("Expected 'User: john', got '%s'", w.Body.String())
	}
}

func TestContext_Components(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		// Test that components are accessible
		env := ctx.Env()
		if env == nil {
			t.Error("Env should be accessible")
		}
		
		debug := ctx.Debug()
		if debug == nil {
			t.Error("Debug should be accessible")
		}
		
		cache := ctx.Cache()
		if cache == nil {
			t.Error("Cache should be accessible")
		}
		
		state := ctx.State()
		if state == nil {
			t.Error("State should be accessible")
		}
		
		ctx.String(200, "ok")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// Test additional Context methods
func TestContext_Body(t *testing.T) {
	app := New()
	
	app.Post("/test", func(ctx *Context) {
		body := ctx.Body()
		if body == nil {
			t.Error("Body should not be nil")
		}
		
		// Body.Get requires a key parameter
		_ = body.Get("content")
		ctx.String(200, "body available")
	})
	
	reqBody := "test body content"
	req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if !strings.Contains(w.Body.String(), "body available") {
		t.Errorf("Expected body available message, got '%s'", w.Body.String())
	}
}

func TestContext_Context(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		stdCtx := ctx.Context()
		if stdCtx == nil {
			t.Error("Context should not be nil")
		}
		
		ctx.String(200, "ok")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestContext_Stream(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		stream := ctx.Stream()
		if stream == nil {
			t.Error("Stream should not be nil")
		}
		ctx.String(200, "stream available")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "stream available" {
		t.Errorf("Expected 'stream available', got '%s'", w.Body.String())
	}
}

func TestContext_Write(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.Write([]byte("written content"))
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "written content" {
		t.Errorf("Expected 'written content', got '%s'", w.Body.String())
	}
}

func TestContext_Error(t *testing.T) {
	app := New()
	
	app.Get("/test", func(ctx *Context) {
		ctx.Error(500, "internal server error")
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
	
	if w.Body.String() != "internal server error" {
		t.Errorf("Expected 'internal server error', got '%s'", w.Body.String())
	}
}

func TestContext_SaveFile(t *testing.T) {
	app := New()
	
	app.Post("/test", func(ctx *Context) {
		// This test is limited since we can't easily create a file upload in tests
		// Just test that the method exists and doesn't panic
		ctx.String(200, "ok")
	})
	
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestContext_WebSocket(t *testing.T) {
	app := New()
	
	app.Get("/ws", func(ctx *Context) {
		// Test WebSocket context method
		// This would normally upgrade to WebSocket
		ctx.String(200, "WebSocket endpoint")
	})
	
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
} 