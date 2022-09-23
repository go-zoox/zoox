package middleware

import (
	"github.com/go-zoox/proxy/utils/rewriter"
	"github.com/go-zoox/zoox"
)

// RewriteConfig is the configuration for the Rewrite middleware.
type RewriteConfig struct {
	Rewrites rewriter.Rewriters
}

// Rewrite is a middleware that rewrites the request path.
func Rewrite(cfg ...*RewriteConfig) zoox.Middleware {
	rewrites := rewriter.Rewriters{}
	if len(cfg) > 0 && cfg[0] != nil {
		if cfg[0].Rewrites != nil {
			rewrites = cfg[0].Rewrites
		}
	}

	return func(ctx *zoox.Context) {
		newPath := rewrites.Rewrite(ctx.Path)
		if newPath != ctx.Path {
			ctx.Request.URL.Path = newPath
			ctx.Path = newPath
		}

		ctx.Next()
	}
}
