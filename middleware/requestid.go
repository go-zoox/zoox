package middleware

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/utils"
)

// RequestID is a middleware that adds a request ID to the context.
func RequestID() zoox.Middleware {
	return func(ctx *zoox.Context) {
		if ctx.Get(utils.RequestIDHeader) == "" {
			ctx.Set(utils.RequestIDHeader, ctx.RequestID())
		}

		ctx.Next()
	}
}
