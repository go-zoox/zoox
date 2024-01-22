package middleware

import (
	"github.com/go-zoox/zoox"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DefaultPrometheus ...
const DefaultPrometheus = "/metrics"

// PrometheusOption ...
type PrometheusOption struct {
	Path string
}

// Prometheus ...
func Prometheus(opts ...func(opt *PrometheusOption)) zoox.Middleware {
	opt := &PrometheusOption{
		Path: DefaultPrometheus,
	}
	for _, o := range opts {
		o(opt)
	}

	return func(ctx *zoox.Context) {
		if ctx.Path == opt.Path {
			promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		ctx.Next()
	}
}
