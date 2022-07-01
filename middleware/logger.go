package middleware

import (
	"time"

	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox"
)

// Logger is a middleware that logs the request as it goes through the handler.
func Logger() zoox.HandlerFunc {
	return func(ctx *zoox.Context) {
		t := time.Now()

		ctx.Next()

		logger.Info("[%s] %s %s %d +%dms", ctx.Request.RemoteAddr, ctx.Method, ctx.Path, ctx.StatusCode, time.Since(t)/time.Millisecond)
	}
}
