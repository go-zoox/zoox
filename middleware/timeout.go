package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-zoox/zoox"
)

func Timeout(timeout time.Duration) zoox.Middleware {
	return func(ctx *zoox.Context) {
		c, cancel := context.WithTimeout(ctx.Request.Context(), timeout)
		defer func() {
			cancel()
			if c.Err() == context.DeadlineExceeded {
				ctx.Status(http.StatusGatewayTimeout)
			}
		}()

		ctx.Request = ctx.Request.WithContext(c)

		ctx.Next()
	}
}
