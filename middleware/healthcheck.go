package middleware

import "github.com/go-zoox/zoox"

func HealthCheck() zoox.HandlerFunc {
	return func(ctx *zoox.Context) {
		if ctx.Path == "/health" {
			ctx.String(200, "OK")
			return
		}

		ctx.Next()
	}
}
