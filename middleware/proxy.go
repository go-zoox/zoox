package middleware

import (
	"regexp"

	"github.com/go-zoox/proxy"
	"github.com/go-zoox/zoox"
)

// ProxyConfig is the config of proxy middlewares
type ProxyConfig struct {
	Rewrites map[string]ProxyRewrite
}

// ProxyRewrite ...
type ProxyRewrite struct {
	Target   string
	Rewrites map[string]string
}

// Proxy is a middleware that authenticates via Basic Auth.
func Proxy(cfg *ProxyConfig) zoox.Middleware {
	return func(ctx *zoox.Context) {
		for key, value := range cfg.Rewrites {
			if matched, err := regexp.MatchString(key, ctx.Path); err == nil && matched {
				p := proxy.NewSingleTarget(value.Target, &proxy.SingleTargetConfig{
					Rewrites: value.Rewrites,
				})

				p.ServeHTTP(ctx.Writer, ctx.Request)
				return
			}
		}

		ctx.Next()
	}
}
