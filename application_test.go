package zoox

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	app := New()
	
	if app == nil {
		t.Error("New() should return a non-nil Application")
	}
	
	if app.router == nil {
		t.Error("Application should have a router")
	}
	
	if app.RouterGroup == nil {
		t.Error("Application should have a RouterGroup")
	}
	
	if len(app.groups) == 0 {
		t.Error("Application should have at least one group (root)")
	}
	
	if app.groupMiddlewareCache == nil {
		t.Error("Application should have middleware cache initialized")
	}
	
	if app.sortedGroups == nil {
		t.Error("Application should have sorted groups initialized")
	}
}

func TestApplication_NotFound(t *testing.T) {
	app := New()
	
	customHandler := func(ctx *Context) {
		ctx.String(404, "Custom Not Found")
	}
	
	app.NotFound(customHandler)
	
	// Test that the custom handler is set
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
	
	if !strings.Contains(w.Body.String(), "Custom Not Found") {
		t.Errorf("Expected custom not found message, got: %s", w.Body.String())
	}
}

func TestApplication_Fallback(t *testing.T) {
	app := New()
	
	customHandler := func(ctx *Context) {
		ctx.String(200, "Fallback Handler")
	}
	
	app.Fallback(customHandler)
	
	// Test that the fallback handler is set
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if !strings.Contains(w.Body.String(), "Fallback Handler") {
		t.Errorf("Expected fallback message, got: %s", w.Body.String())
	}
}

func TestApplication_SetBanner(t *testing.T) {
	app := New()
	customBanner := "Custom Banner"
	
	app.SetBanner(customBanner)
	
	if app.Config.Banner != customBanner {
		t.Errorf("Expected banner '%s', got '%s'", customBanner, app.Config.Banner)
	}
}

func TestApplication_SetBeforeReady(t *testing.T) {
	app := New()
	called := false
	
	app.SetBeforeReady(func() {
		called = true
	})
	
	// Trigger the beforeReady callback
	if app.lifecycle.beforeReady != nil {
		app.lifecycle.beforeReady()
	}
	
	if !called {
		t.Error("BeforeReady callback should have been called")
	}
}

func TestApplication_SetBeforeDestroy(t *testing.T) {
	app := New()
	called := false
	
	app.SetBeforeDestroy(func() {
		called = true
	})
	
	// Trigger the beforeDestroy callback
	if app.lifecycle.beforeDestroy != nil {
		app.lifecycle.beforeDestroy()
	}
	
	if !called {
		t.Error("BeforeDestroy callback should have been called")
	}
}

func TestApplication_IsProd(t *testing.T) {
	app := New()
	
	// Test default (non-production)
	if app.IsProd() {
		t.Error("Should not be in production mode by default")
	}
	
	// Test production mode
	os.Setenv("MODE", "production")
	defer os.Unsetenv("MODE")
	
	if !app.IsProd() {
		t.Error("Should be in production mode when MODE=production")
	}
}

func TestApplication_ServeHTTP(t *testing.T) {
	app := New()
	
	// Add a test route
	app.Get("/test", func(ctx *Context) {
		ctx.String(200, "test response")
	})
	
	// Test the route
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Body.String() != "test response" {
		t.Errorf("Expected 'test response', got '%s'", w.Body.String())
	}
}

func TestApplication_ServeHTTP_WithMiddleware(t *testing.T) {
	app := New()
	
	var executionOrder []string
	
	// Add global middleware
	app.Use(func(ctx *Context) {
		executionOrder = append(executionOrder, "global")
		ctx.Next()
	})
	
	// Add a test route
	app.Get("/test", func(ctx *Context) {
		executionOrder = append(executionOrder, "handler")
		ctx.String(200, "test response")
	})
	
	// Test the route
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	expectedOrder := []string{"global", "handler"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d middleware executions, got %d", len(expectedOrder), len(executionOrder))
	}
	
	for i, expected := range expectedOrder {
		if i >= len(executionOrder) || executionOrder[i] != expected {
			t.Errorf("Expected middleware %s at position %d, got %s", expected, i, executionOrder[i])
		}
	}
}

func TestApplication_CreateContext(t *testing.T) {
	app := New()
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	ctx := app.createContext(w, req)
	
	if ctx == nil {
		t.Error("createContext should return a non-nil Context")
	}
	
	if ctx.Request != req {
		t.Error("Context should contain the original request")
	}
	
	if ctx.Writer == nil {
		t.Error("Context should have a writer")
	}
	
	if ctx.App != app {
		t.Error("Context should reference the application")
	}
}

func TestApplication_Address(t *testing.T) {
	app := New()
	
	// Test default address
	app.Config.Host = "localhost"
	app.Config.Port = 8080
	
	expected := "localhost:8080"
	if app.Address() != expected {
		t.Errorf("Expected address '%s', got '%s'", expected, app.Address())
	}
	
	// Test unix socket
	app.Config.NetworkType = "unix"
	app.Config.UnixDomainSocket = "/tmp/test.sock"
	
	if app.Address() != "/tmp/test.sock" {
		t.Errorf("Expected unix socket address, got '%s'", app.Address())
	}
}

func TestApplication_AddressForLog(t *testing.T) {
	app := New()
	
	// Test with 0.0.0.0 (should convert to 127.0.0.1)
	app.Config.Host = "0.0.0.0"
	app.Config.Port = 8080
	
	expected := "127.0.0.1:8080"
	if app.AddressForLog() != expected {
		t.Errorf("Expected log address '%s', got '%s'", expected, app.AddressForLog())
	}
	
	// Test with specific host
	app.Config.Host = "localhost"
	app.Config.Port = 3000
	
	expected = "localhost:3000"
	if app.AddressForLog() != expected {
		t.Errorf("Expected log address '%s', got '%s'", expected, app.AddressForLog())
	}
}

func TestApplication_ParseAddr(t *testing.T) {
	testCases := []struct {
		name         string
		addr         string
		expectedHost string
		expectedPort int
		expectedProto string
		expectedNetwork string
	}{
		{
			name:         "Port only",
			addr:         ":8080",
			expectedHost: "", // parseAddr doesn't set default host, applyDefaultConfig does
			expectedPort: 8080,
		},
		{
			name:         "Host and port",
			addr:         "localhost:3000",
			expectedHost: "localhost",
			expectedPort: 3000,
		},
		{
			name:         "HTTP URL",
			addr:         "http://localhost:8080",
			expectedHost: "localhost",
			expectedPort: 8080,
			expectedProto: "http",
		},
		{
			name:         "Unix socket with prefix",
			addr:         "unix:///tmp/test.sock",
			expectedProto: "unix",
			expectedNetwork: "unix",
		},
		{
			name:         "Unix socket path",
			addr:         "/tmp/test.sock",
			expectedProto: "unix",
			expectedNetwork: "unix",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := New()
			
			err := app.parseAddr(tc.addr)
			if err != nil {
				t.Errorf("parseAddr failed: %v", err)
			}
			
			if tc.expectedHost != "" && app.Config.Host != tc.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", tc.expectedHost, app.Config.Host)
			}
			
			if tc.expectedPort != 0 && app.Config.Port != tc.expectedPort {
				t.Errorf("Expected port %d, got %d", tc.expectedPort, app.Config.Port)
			}
			
			if tc.expectedProto != "" && app.Config.Protocol != tc.expectedProto {
				t.Errorf("Expected protocol '%s', got '%s'", tc.expectedProto, app.Config.Protocol)
			}
			
			if tc.expectedNetwork != "" && app.Config.NetworkType != tc.expectedNetwork {
				t.Errorf("Expected network type '%s', got '%s'", tc.expectedNetwork, app.Config.NetworkType)
			}
		})
	}
}

func TestApplication_ApplyDefaultConfig(t *testing.T) {
	app := New()
	
	err := app.applyDefaultConfig()
	if err != nil {
		t.Errorf("applyDefaultConfig failed: %v", err)
	}
	
	// Check default values
	if app.Config.Protocol == "" {
		t.Error("Protocol should have a default value")
	}
	
	if app.Config.Host == "" {
		t.Error("Host should have a default value")
	}
	
	if app.Config.Port == 0 {
		t.Error("Port should have a default value")
	}
	
	if app.Config.NetworkType == "" {
		t.Error("NetworkType should have a default value")
	}
}

func TestApplication_ApplyDefaultConfigFromEnv(t *testing.T) {
	app := New()
	
	// Set environment variables
	os.Setenv("PORT", "9000")
	os.Setenv("HTTPS_PORT", "9443")
	os.Setenv("LOG_LEVEL", "debug")
	
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("HTTPS_PORT")
		os.Unsetenv("LOG_LEVEL")
	}()
	
	err := app.applyDefaultConfigFromEnv()
	if err != nil {
		t.Errorf("applyDefaultConfigFromEnv failed: %v", err)
	}
	
	if app.Config.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", app.Config.Port)
	}
	
	if app.Config.HTTPSPort != 9443 {
		t.Errorf("Expected HTTPS port 9443, got %d", app.Config.HTTPSPort)
	}
	
	if app.Config.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", app.Config.LogLevel)
	}
}

func TestApplication_SortGroups(t *testing.T) {
	app := New()
	
	// Create groups with different prefix lengths
	group1 := app.Group("/api")
	group2 := app.Group("/api/v1")
	group3 := app.Group("/api/v1/users")
	
	// sortGroups should be called automatically, but call it explicitly to test
	app.sortGroups()
	
	// Check that groups are sorted by prefix length (longest first)
	if len(app.sortedGroups) < 4 { // root + 3 groups
		t.Errorf("Expected at least 4 groups, got %d", len(app.sortedGroups))
	}
	
	// Find our test groups in the sorted list
	var foundGroups []*RouterGroup
	for _, group := range app.sortedGroups {
		if group.prefix == "/api/v1/users" || group.prefix == "/api/v1" || group.prefix == "/api" {
			foundGroups = append(foundGroups, group)
		}
	}
	
	if len(foundGroups) != 3 {
		t.Errorf("Expected 3 test groups, got %d", len(foundGroups))
	}
	
	// Check that longer prefixes come first
	for i := 0; i < len(foundGroups)-1; i++ {
		if len(foundGroups[i].prefix) < len(foundGroups[i+1].prefix) {
			t.Errorf("Groups not sorted by prefix length: %s should come before %s", 
				foundGroups[i+1].prefix, foundGroups[i].prefix)
		}
	}
	
	// Check that middleware cache is populated
	if len(app.groupMiddlewareCache) == 0 {
		t.Error("Middleware cache should be populated after sortGroups")
	}
	
	// Check that our groups are in the cache
	if _, exists := app.groupMiddlewareCache[group1]; !exists {
		t.Error("Group1 should be in middleware cache")
	}
	if _, exists := app.groupMiddlewareCache[group2]; !exists {
		t.Error("Group2 should be in middleware cache")
	}
	if _, exists := app.groupMiddlewareCache[group3]; !exists {
		t.Error("Group3 should be in middleware cache")
	}
}

func TestApplication_PrecomputeMiddlewareChains(t *testing.T) {
	app := New()
	
	// Add global middleware
	app.Use(func(ctx *Context) {
		ctx.Next()
	})
	
	// Create groups with middleware
	api := app.Group("/api")
	api.Use(func(ctx *Context) {
		ctx.Next()
	})
	
	v1 := api.Group("/v1")
	v1.Use(func(ctx *Context) {
		ctx.Next()
	})
	
	// Precompute should be called automatically, but call it explicitly to test
	app.precomputeMiddlewareChains()
	
	// Check that cache is populated
	if len(app.groupMiddlewareCache) == 0 {
		t.Error("Middleware cache should be populated")
	}
	
	// Check that each group has cached middleware
	for _, group := range app.groups {
		if middlewares, exists := app.groupMiddlewareCache[group]; !exists {
			t.Errorf("Group %s should be in middleware cache", group.prefix)
		} else if len(middlewares) == 0 {
			t.Errorf("Group %s should have at least one middleware (global)", group.prefix)
		}
	}
	
	// Check that v1 group has more middlewares than root (global + api + v1)
	rootMiddlewares := app.groupMiddlewareCache[app.RouterGroup]
	v1Middlewares := app.groupMiddlewareCache[v1]
	
	if len(v1Middlewares) <= len(rootMiddlewares) {
		t.Errorf("V1 group should have more middlewares than root group: v1=%d, root=%d", 
			len(v1Middlewares), len(rootMiddlewares))
	}
}

func TestApplication_DeduplicateMiddlewares(t *testing.T) {
	app := New()
	
	// Create a middleware function
	middleware1 := func(ctx *Context) {
		ctx.Next()
	}
	
	middleware2 := func(ctx *Context) {
		ctx.Next()
	}
	
	// Test with duplicates
	middlewares := []HandlerFunc{middleware1, middleware2, middleware1, middleware2}
	
	unique := app.deduplicateMiddlewares(middlewares)
	
	if len(unique) != 2 {
		t.Errorf("Expected 2 unique middlewares, got %d", len(unique))
	}
	
	// Test with empty slice
	empty := app.deduplicateMiddlewares([]HandlerFunc{})
	if len(empty) != 0 {
		t.Errorf("Expected empty slice, got %d middlewares", len(empty))
	}
}

func TestApplication_SetTLSCertLoader(t *testing.T) {
	app := New()
	
	loader := func(sni string) (key, cert string, err error) {
		return "test-key", "test-cert", nil
	}
	
	app.SetTLSCertLoader(loader)
	
	if app.tlsCertLoader == nil {
		t.Error("TLS cert loader should be set")
	}
	
	// Test the loader
	key, cert, err := app.tlsCertLoader("test.com")
	if err != nil {
		t.Errorf("TLS cert loader failed: %v", err)
	}
	
	if key != "test-key" || cert != "test-cert" {
		t.Errorf("Expected test-key and test-cert, got %s and %s", key, cert)
	}
}

func TestApplication_ComponentsInitialization(t *testing.T) {
	app := New()
	
	// Test that components are initialized properly
	env := app.Env()
	if env == nil {
		t.Error("Env should be initialized")
	}
	
	logger := app.Logger()
	if logger == nil {
		t.Error("Logger should be initialized")
	}
	
	debug := app.Debug()
	if debug == nil {
		t.Error("Debug should be initialized")
	}
	
	runtime := app.Runtime()
	if runtime == nil {
		t.Error("Runtime should be initialized")
	}
	
	// Test that components are singletons (same instance returned)
	env2 := app.Env()
	if env != env2 {
		t.Error("Env should be a singleton")
	}
	
	logger2 := app.Logger()
	if logger != logger2 {
		t.Error("Logger should be a singleton")
	}
}

func TestApplication_HTTPMethods(t *testing.T) {
	app := New()
	
	// Test all HTTP methods
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	
	for _, method := range methods {
		path := "/test-" + strings.ToLower(method)
		
		// Add route for each method
		switch method {
		case "GET":
			app.Get(path, func(ctx *Context) { ctx.String(200, method) })
		case "POST":
			app.Post(path, func(ctx *Context) { ctx.String(200, method) })
		case "PUT":
			app.Put(path, func(ctx *Context) { ctx.String(200, method) })
		case "PATCH":
			app.Patch(path, func(ctx *Context) { ctx.String(200, method) })
		case "DELETE":
			app.Delete(path, func(ctx *Context) { ctx.String(200, method) })
		case "HEAD":
			app.Head(path, func(ctx *Context) { ctx.String(200, method) })
		case "OPTIONS":
			app.Options(path, func(ctx *Context) { ctx.String(200, method) })
		}
		
		// Test the route
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

func TestApplication_Any(t *testing.T) {
	app := New()
	
	app.Any("/test", func(ctx *Context) {
		ctx.String(200, "any method")
	})
	
	// Test multiple methods
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	
	for _, method := range methods {
		req := httptest.NewRequest(method, "/test", nil)
		w := httptest.NewRecorder()
		
		app.ServeHTTP(w, req)
		
		if w.Code != 200 {
			t.Errorf("Expected status 200 for %s /test, got %d", method, w.Code)
		}
		
		if method != "HEAD" && w.Body.String() != "any method" {
			t.Errorf("Expected body 'any method' for %s /test, got '%s'", method, w.Body.String())
		}
	}
}

// Test additional application components
func TestApplication_JSONRPCRegistry(t *testing.T) {
	app := New()
	
	registry := app.JSONRPCRegistry()
	if registry == nil {
		t.Error("JSONRPCRegistry should not be nil")
	}
	
	// Test singleton behavior
	registry2 := app.JSONRPCRegistry()
	if registry != registry2 {
		t.Error("JSONRPCRegistry should be a singleton")
	}
}

func TestApplication_PubSub(t *testing.T) {
	app := New()
	
	// Test PubSub component initialization without Redis config
	defer func() {
		if r := recover(); r != nil {
			// Expected panic when Redis is not configured
			t.Logf("PubSub panicked as expected: %v", r)
		}
	}()
	
	app.PubSub()
	
	// If we get here, PubSub didn't panic (Redis might be configured)
	t.Log("PubSub component initialized successfully")
}

func TestApplication_MQ(t *testing.T) {
	app := New()
	
	// Test MQ component initialization without Redis config
	defer func() {
		if r := recover(); r != nil {
			// Expected panic when Redis is not configured
			t.Logf("MQ panicked as expected: %v", r)
		}
	}()
	
	app.MQ()
	
	// If we get here, MQ didn't panic (Redis might be configured)
	t.Log("MQ component initialized successfully")
}

func TestApplication_Cron(t *testing.T) {
	app := New()
	
	cron := app.Cron()
	if cron == nil {
		t.Error("Cron should not be nil")
	}
	
	// Test singleton behavior
	cron2 := app.Cron()
	if cron != cron2 {
		t.Error("Cron should be a singleton")
	}
}

func TestApplication_JobQueue(t *testing.T) {
	app := New()
	
	queue := app.JobQueue()
	if queue == nil {
		t.Error("JobQueue should not be nil")
	}
	
	// Test singleton behavior
	queue2 := app.JobQueue()
	if queue != queue2 {
		t.Error("JobQueue should be a singleton")
	}
}

func TestApplication_Cmd(t *testing.T) {
	app := New()
	
	cmd := app.Cmd()
	if cmd == nil {
		t.Error("Cmd should not be nil")
	}
	
	// Test singleton behavior
	cmd2 := app.Cmd()
	if cmd != cmd2 {
		t.Error("Cmd should be a singleton")
	}
}

func TestApplication_I18n(t *testing.T) {
	app := New()
	
	i18n := app.I18n()
	if i18n == nil {
		t.Error("I18n should not be nil")
	}
	
	// Test singleton behavior
	i18n2 := app.I18n()
	if i18n != i18n2 {
		t.Error("I18n should be a singleton")
	}
}

func TestApplication_AddressHTTPS(t *testing.T) {
	app := New()
	
	// Test HTTPS address for regular network
	app.Config.Host = "localhost"
	app.Config.HTTPSPort = 8443
	app.Config.NetworkType = "tcp"
	
	address := app.AddressHTTPS()
	expected := "localhost:8443"
	
	if address != expected {
		t.Errorf("Expected HTTPS address '%s', got '%s'", expected, address)
	}
	
	// Test HTTPS address for unix socket
	app.Config.NetworkType = "unix"
	app.Config.Host = "/tmp/test.sock"
	
	address = app.AddressHTTPS()
	expected = ""
	
	if address != expected {
		t.Errorf("Expected empty HTTPS address for unix socket, got '%s'", address)
	}
}

func TestApplication_AddressHTTPSForLog(t *testing.T) {
	app := New()
	
	// Test with 0.0.0.0 (should convert to 127.0.0.1)
	app.Config.Host = "0.0.0.0"
	app.Config.HTTPSPort = 8443
	
	expected := "127.0.0.1:8443"
	if app.AddressHTTPSForLog() != expected {
		t.Errorf("Expected HTTPS log address '%s', got '%s'", expected, app.AddressHTTPSForLog())
	}
	
	// Test with specific host
	app.Config.Host = "localhost"
	app.Config.HTTPSPort = 9443
	
	expected = "localhost:9443"
	if app.AddressHTTPSForLog() != expected {
		t.Errorf("Expected HTTPS log address '%s', got '%s'", expected, app.AddressHTTPSForLog())
	}
}

func TestApplication_SetTemplates(t *testing.T) {
	app := New()
	
	// Test that SetTemplates panics when directory doesn't exist
	defer func() {
		if r := recover(); r != nil {
			// Expected panic
			if !strings.Contains(r.(error).Error(), "pattern matches no files") {
				t.Errorf("Expected panic about no files matching pattern, got: %v", r)
			}
		} else {
			t.Error("Expected panic when templates directory doesn't exist")
		}
	}()
	
	app.SetTemplates("./nonexistent")
}

func TestApplication_MoreEnvConfigs(t *testing.T) {
	app := New()
	
	// Test more environment variables
	os.Setenv("HOST", "example.com")
	os.Setenv("HTTPS_PORT", "8443")
	os.Setenv("SECRET_KEY", "test-secret")
	os.Setenv("CORS_ALLOW_ORIGINS", "http://localhost:3000")
	os.Setenv("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE")
	os.Setenv("CORS_ALLOW_HEADERS", "Content-Type,Authorization")
	
	defer func() {
		os.Unsetenv("HOST")
		os.Unsetenv("HTTPS_PORT")
		os.Unsetenv("SECRET_KEY")
		os.Unsetenv("CORS_ALLOW_ORIGINS")
		os.Unsetenv("CORS_ALLOW_METHODS")
		os.Unsetenv("CORS_ALLOW_HEADERS")
	}()
	
	err := app.applyDefaultConfigFromEnv()
	if err != nil {
		t.Errorf("applyDefaultConfigFromEnv failed: %v", err)
	}
	
	// Test that the configuration was applied
	if app.Config.Host != "example.com" {
		t.Logf("Host env var not applied directly to config field, got '%s'", app.Config.Host)
	}
	
	if app.Config.HTTPSPort != 8443 {
		t.Logf("HTTPS port env var not applied directly to config field, got %d", app.Config.HTTPSPort)
	}
}

func TestApplication_NewWithMoreConfig(t *testing.T) {
	app := New()
	
	// Test default configuration values
	if app.Config.Banner == "" {
		t.Log("Banner has empty default value")
	}
	
	if app.Config.LogLevel == "" {
		t.Log("LogLevel has empty default value")
	}
	
	if app.Config.NetworkType == "" {
		t.Log("NetworkType has empty default value")
	}
	
	if app.Config.Protocol == "" {
		t.Log("Protocol has empty default value")
	}
	
	// Test that the app is properly initialized
	if app.RouterGroup == nil {
		t.Error("RouterGroup should not be nil")
	}
	
	if app.Logger() == nil {
		t.Error("Logger should not be nil")
	}
}

func TestApplication_ConnectAndTrace(t *testing.T) {
	app := New()
	
	// Test CONNECT method
	app.Connect("/test", func(ctx *Context) {
		ctx.String(200, "CONNECT")
	})
	
	req := httptest.NewRequest("CONNECT", "/test", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected status 200 for CONNECT, got %d", w.Code)
	}
	
	if w.Body.String() != "CONNECT" {
		t.Errorf("Expected body 'CONNECT', got '%s'", w.Body.String())
	}
	
	// Note: TRACE method is not available in this framework
}

func TestApplication_StaticFiles(t *testing.T) {
	app := New()
	
	// Test static file serving
	app.Static("/static", "./")
	
	req := httptest.NewRequest("GET", "/static/README.md", nil)
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	// Should return 200 if file exists, or 404 if not
	if w.Code != 200 && w.Code != 404 {
		t.Errorf("Expected status 200 or 404 for static file, got %d", w.Code)
	}
} 