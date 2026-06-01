package zoox

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppProxyInheritsTrustProxyFromConfig(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("X-Forwarded-Proto"), "https"; got != want {
			t.Fatalf("X-Forwarded-Proto: got %q, want %q", got, want)
		}
		if got, want := r.Header.Get("X-Forwarded-Port"), "443"; got != want {
			t.Fatalf("X-Forwarded-Port: got %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()

	app := New()
	app.Config.TrustProxy = true
	app.Proxy("/api", backend.URL)

	frontend := httptest.NewServer(app)
	defer frontend.Close()

	req, _ := http.NewRequest(http.MethodGet, frontend.URL+"/api/healthz", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Port", "443")

	res, err := frontend.Client().Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	res.Body.Close()
}

func TestContextProxyInheritsTrustProxyFromConfig(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("X-Forwarded-Proto"), "https"; got != want {
			t.Fatalf("X-Forwarded-Proto: got %q, want %q", got, want)
		}
		if got, want := r.Header.Get("X-Forwarded-Port"), "443"; got != want {
			t.Fatalf("X-Forwarded-Port: got %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()

	app := New()
	app.Config.TrustProxy = true
	app.Get("/proxy", func(ctx *Context) {
		ctx.Proxy(backend.URL)
	})

	frontend := httptest.NewServer(app)
	defer frontend.Close()

	req, _ := http.NewRequest(http.MethodGet, frontend.URL+"/proxy", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Port", "443")

	res, err := frontend.Client().Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	res.Body.Close()
}
