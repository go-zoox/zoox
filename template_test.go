package zoox

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApplication_Template(t *testing.T) {
	app := New()
	
	// Test template rendering with inline content
	app.Get("/template", func(ctx *Context) {
		ctx.Template(200, func(tc *TemplateConfig) {
			tc.ContentType = "text/html"
			tc.Content = "<h1>{{.title}}</h1><p>Hello {{.name}}</p>"
			tc.Data = H{
				"title": "Test Page",
				"name":  "John",
			}
		})
	})
	
	req := httptest.NewRequest("GET", "/template", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Expected HTML content type, got '%s'", w.Header().Get("Content-Type"))
	}
	
	if !strings.Contains(w.Body.String(), "Test Page") {
		t.Errorf("Expected rendered template to contain 'Test Page', got '%s'", w.Body.String())
	}
	
	if !strings.Contains(w.Body.String(), "Hello John") {
		t.Errorf("Expected rendered template to contain 'Hello John', got '%s'", w.Body.String())
	}
}

func TestApplication_SetTemplates(t *testing.T) {
	app := New()
	
	// Test setting templates directory that doesn't exist should panic
	defer func() {
		if r := recover(); r != nil {
			// Expected panic when templates directory doesn't exist
			if !strings.Contains(r.(error).Error(), "pattern matches no files") {
				t.Errorf("Expected panic about no files matching pattern, got: %v", r)
			}
		} else {
			t.Error("Expected panic when templates directory doesn't exist")
		}
	}()
	
	app.SetTemplates("./templates")
}

func TestContext_Render(t *testing.T) {
	app := New()
	
	// Test render method - this will fail without actual template files
	// but we can test that the method exists and behaves correctly
	app.Get("/render", func(ctx *Context) {
		ctx.Render(200, "test.html", H{
			"message": "Hello World",
		})
	})
	
	req := httptest.NewRequest("GET", "/render", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// Since we don't have actual template files, this will return 500
	// but we can test that the method exists and doesn't panic
	if w.Code != 500 {
		t.Errorf("Expected status 500 (template not found), got %d", w.Code)
	}
	
	if !strings.Contains(w.Body.String(), "templates is not initialized") {
		t.Errorf("Expected template error message, got '%s'", w.Body.String())
	}
}

func TestContext_RenderHTML(t *testing.T) {
	app := New()
	
	// Test render HTML method - this uses a file path
	app.Get("/render-html", func(ctx *Context) {
		ctx.RenderHTML("./test.html")
	})
	
	req := httptest.NewRequest("GET", "/render-html", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// Since we don't have actual HTML files, this will return 500
	// but we can test that the method exists and doesn't panic
	if w.Code != 500 {
		t.Errorf("Expected status 500 (file not found), got %d", w.Code)
	}
} 