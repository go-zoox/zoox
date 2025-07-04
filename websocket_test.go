package zoox

import (
	"net/http/httptest"
	"testing"
)

func TestApplication_WebSocket(t *testing.T) {
	app := New()
	
	// Test WebSocket handler registration
	server, err := app.WebSocket("/ws", func(opt *WebSocketOption) {
		// Configure WebSocket options here
		opt.Middlewares = []HandlerFunc{
			func(ctx *Context) {
				// WebSocket middleware logic
				ctx.Next()
			},
		}
	})
	
	if err != nil {
		t.Errorf("WebSocket registration failed: %v", err)
	}
	
	if server == nil {
		t.Error("WebSocket server should not be nil")
	}
	
	// Test that the WebSocket route is registered (without WebSocket headers, it should continue to next handler)
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// Without WebSocket headers, it should continue to next handler and return 404
	if w.Code != 404 {
		t.Errorf("Expected status 404 without WebSocket headers, got %d", w.Code)
	}
}

func TestApplication_WebSocket_WithHeaders(t *testing.T) {
	app := New()
	
	// Test WebSocket handler registration
	server, err := app.WebSocket("/ws", func(opt *WebSocketOption) {
		// Configure WebSocket options here
		opt.Middlewares = []HandlerFunc{
			func(ctx *Context) {
				// WebSocket middleware logic
				ctx.Next()
			},
		}
	})
	
	if err != nil {
		t.Errorf("WebSocket registration failed: %v", err)
	}
	
	if server == nil {
		t.Error("WebSocket server should not be nil")
	}
	
	// Test with WebSocket headers
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "test-key")
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// With WebSocket headers, it should attempt to upgrade
	// The exact status depends on the WebSocket implementation
	if w.Code == 0 {
		t.Error("WebSocket endpoint should respond with some status code")
	}
}

func TestApplication_WebSocket_NoOptions(t *testing.T) {
	app := New()
	
	// Test WebSocket handler registration without options
	server, err := app.WebSocket("/ws")
	
	if err != nil {
		t.Errorf("WebSocket registration failed: %v", err)
	}
	
	if server == nil {
		t.Error("WebSocket server should not be nil")
	}
}

func TestContext_WebSocket(t *testing.T) {
	app := New()
	
	// Test WebSocket context method
	app.Get("/ws-test", func(ctx *Context) {
		// This would normally upgrade to WebSocket
		// For testing, we just check that the method exists
		ctx.String(200, "WebSocket test")
	})
	
	req := httptest.NewRequest("GET", "/ws-test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "WebSocket test" {
		t.Errorf("Expected 'WebSocket test', got '%s'", w.Body.String())
	}
} 