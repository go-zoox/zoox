package middleware

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/utils"
)

// RequestID is a middleware that adds a request ID to the context.
func RequestID() zoox.Middleware {
	return func(ctx *zoox.Context) {
		if ctx.Header().Get(utils.RequestIDHeader) == "" {
			requestID := ctx.RequestID()

			// update request to next
			ctx.Request.Header.Set(utils.RequestIDHeader, requestID)

			// set to response
			ctx.Set(utils.RequestIDHeader, requestID)
		}

		ctx.Next()
	}
}
