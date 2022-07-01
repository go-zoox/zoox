package middleware

import "github.com/go-zoox/zoox"

func HealthCheck() zoox.HandlerFunc {
	return func(ctx *zoox.Context) {
		ctx.String(200, "OK")
	}
}
