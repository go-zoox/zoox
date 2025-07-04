package zoox

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrapH(t *testing.T) {
	// Create a standard http.Handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("wrapped handler"))
	})
	
	// Wrap it with WrapH
	wrappedHandler := WrapH(handler)
	
	// Test the wrapped handler
	app := New()
	app.Get("/test", wrappedHandler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "wrapped handler" {
		t.Errorf("Expected 'wrapped handler', got '%s'", w.Body.String())
	}
}

func TestWrapF(t *testing.T) {
	// Create a standard http.HandlerFunc
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte("wrapped function"))
	}
	
	// Wrap it with WrapF
	wrappedHandler := WrapF(handlerFunc)
	
	// Test the wrapped handler
	app := New()
	app.Get("/test", wrappedHandler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 201 {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
	
	if w.Body.String() != "wrapped function" {
		t.Errorf("Expected 'wrapped function', got '%s'", w.Body.String())
	}
} 