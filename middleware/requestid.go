package middleware

import (
	"github.com/go-zoox/zoox"
)

// RequestID is a middleware that adds a request ID to the context.
func RequestID() zoox.Middleware {
	return func(ctx *zoox.Context) {
		if ctx.Get("X-Request-Id") == "" {
			ctx.Set("X-Request-Id", ctx.RequestID())
		}

		ctx.Next()
	}
}
