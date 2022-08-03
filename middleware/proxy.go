package middleware

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/go-zoox/proxy/utils/rewriter"
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
				// @BUG: this is not working
				// p := proxy.NewSingleTarget(value.Target, &proxy.SingleTargetConfig{
				// 	Rewrites: value.Rewrites,
				// })

				// p.ServeHTTP(ctx.Writer, ctx.Request)
				// return

				rewriters := rewriter.Rewriters{}
				for k, v := range value.Rewrites {
					rewriters = append(rewriters, &rewriter.Rewriter{
						From: k,
						To:   v,
					})
				}

				ctx.Request.URL.Path = rewriters.Rewrite(ctx.Path)
				ctx.Path = ctx.Request.URL.Path

				u, err := url.Parse(value.Target)
				if err != nil {
					panic(fmt.Errorf("invalid proxy target: %s", value.Target))
				}

				p := httputil.NewSingleHostReverseProxy(u)

				p.ServeHTTP(ctx.Writer, ctx.Request)
				return
			}
		}

		ctx.Next()
	}
}
