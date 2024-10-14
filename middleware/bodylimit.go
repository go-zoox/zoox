package middleware

import (
	"net/http"

	"github.com/go-zoox/zoox"
)

type BodyLimitConfig struct {
	MaxSize int64
}

// BodyLimit is a middleware that sets a body size limit for the request.
func BodyLimit(opts ...func(cfg *BodyLimitConfig)) zoox.Middleware {
	opt := &BodyLimitConfig{
		// MaxSize: 1024 * 1024 * 10,
	}
	for _, o := range opts {
		o(opt)
	}

	return func(ctx *zoox.Context) {
		// @TODO handle client request body size limit
		if ctx.App.Config.BodySizeLimit > 0 {
			ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, opt.MaxSize)
		}

		ctx.Next()
	}
}
