package middleware

import "github.com/go-zoox/zoox"

// HealthCheck is a middleware that checks the health of the application.
func HealthCheck(path ...string) zoox.Middleware {
	pathX := "/health"
	if len(path) > 0 && path[0] != "" {
		pathX = path[0]
	}

	return func(ctx *zoox.Context) {
		if ctx.Path == pathX {
			ctx.String(200, "OK")
			return
		}

		ctx.Next()
	}
}
