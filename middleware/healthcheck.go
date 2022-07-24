package middleware

import "github.com/go-zoox/zoox"

// HealthCheck is a middleware that checks the health of the application.
func HealthCheck() zoox.Middleware {
	return func(ctx *zoox.Context) {
		if ctx.Path == "/health" {
			ctx.String(200, "OK")
			return
		}

		ctx.Next()
	}
}
