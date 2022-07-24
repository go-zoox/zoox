package middleware

import (
	"github.com/go-zoox/proxy/utils/rewriter"
	"github.com/go-zoox/zoox"
)

// RewriteConfig is the configuration for the Rewrite middleware.
type RewriteConfig struct {
	Rewrites map[string]string
}

// Rewrite is a middleware that rewrites the request path.
func Rewrite(cfg ...*RewriteConfig) zoox.Middleware {
	cfgX := &RewriteConfig{}
	if len(cfg) > 0 {
		cfgX = cfg[0]
	}

	rewriters := rewriter.Rewriters{}
	for k, v := range cfgX.Rewrites {
		rewriters = append(rewriters, &rewriter.Rewriter{
			From: k,
			To:   v,
		})
	}

	return func(ctx *zoox.Context) {
		newPath := rewriters.Rewrite(ctx.Path)
		if newPath != ctx.Path {
			ctx.Request.URL.Path = newPath
			ctx.Path = newPath
		}

		ctx.Next()
	}
}
