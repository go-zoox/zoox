package middleware

import (
	"regexp"

	"github.com/go-zoox/proxy"
	"github.com/go-zoox/proxy/utils/rewriter"
	"github.com/go-zoox/zoox"
)

// ProxyConfig is the config of proxy middlewares
type ProxyConfig struct {
	// Rewrites map[string]ProxyRewrite
	Rewrites ProxyGroupRewrites `yaml:"rewrites" json:"rewrites"`
}

// ProxyGroupRewrites is a list of rewrite rules
type ProxyGroupRewrites []ProxyGroupRewrite

// ProxyGroupRewrite is a group of proxy rewrites
type ProxyGroupRewrite struct {
	Name    string       `yaml:"name" json:"name"`
	RegExp  string       `yaml:"regexp" json:"regexp"`
	Rewrite ProxyRewrite `yaml:"rewrite" json:"rewrite"`
}

// ProxyRewrite ...
type ProxyRewrite struct {
	Target   string            `yaml:"target" json:"target"`
	Rewrites ProxyRewriteRules `yaml:"rewrites" json:"rewrites"`
}

// ProxyRewriteRules ...
type ProxyRewriteRules = rewriter.Rewriters

// Proxy is a middleware that authenticates via Basic Auth.
func Proxy(cfg *ProxyConfig) zoox.Middleware {
	return func(ctx *zoox.Context) {
		for _, group := range cfg.Rewrites {
			if matched, err := regexp.MatchString(group.RegExp, ctx.Path); err == nil && matched {
				// @BUG: this is not working
				p := proxy.NewSingleTarget(group.Rewrite.Target, &proxy.SingleTargetConfig{
					Rewrites: group.Rewrite.Rewrites,
					// ChangeOrigin: true,
				})

				p.ServeHTTP(ctx.Writer, ctx.Request)
				return
			}
		}

		ctx.Next()
	}
}
