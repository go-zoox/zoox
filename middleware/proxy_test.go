package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-zoox/zoox"
)

func TestProxySingleTargetInheritsAppTrustProxy(t *testing.T) {
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

	app := zoox.New()
	app.Config.TrustProxy = true
	app.Use(ProxySingleTarget(func(ctx *zoox.Context, cfg *ProxySingleTargetConfig) (next bool, err error) {
		cfg.Target = backend.URL
		return false, nil
	}))

	frontend := httptest.NewServer(app)
	defer frontend.Close()

	req, _ := http.NewRequest(http.MethodGet, frontend.URL+"/api", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Port", "443")

	res, err := frontend.Client().Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	res.Body.Close()
}
