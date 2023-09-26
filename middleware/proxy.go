package middleware

import (
	"regexp"

	"github.com/go-zoox/proxy"
	"github.com/go-zoox/proxy/utils/rewriter"
	"github.com/go-zoox/zoox"
)

type ProxyConfig = proxy.Config

func Proxy(fn func(ctx *zoox.Context, cfg *ProxyConfig) (next bool, err error)) zoox.Middleware {
	return func(ctx *zoox.Context) {
		cfg := &ProxyConfig{}
		next, err := fn(ctx, cfg)
		if err != nil {
			if v, ok := err.(*proxy.HTTPError); ok {
				ctx.Fail(err, v.Status(), v.Error())
			} else {
				ctx.Fail(err, 500, "proxy error")
			}
			return
		}

		if next {
			ctx.Next()
			return
		}

		zoox.WrapH(proxy.New(cfg))(ctx)
	}
}

// ProxySingleTargetConfig defines the proxy config
type ProxySingleTargetConfig struct {
	// internal proxy config
	proxy.SingleHostConfig

	// target url
	Target string
}

// ProxySingleTarget is a middleware that proxies the request.
func ProxySingleTarget(fn func(ctx *zoox.Context, cfg *ProxySingleTargetConfig) (next bool, err error)) zoox.Middleware {
	return func(ctx *zoox.Context) {
		proxyCfg := &ProxySingleTargetConfig{}
		next, err := fn(ctx, proxyCfg)
		if err != nil {
			if v, ok := err.(*proxy.HTTPError); ok {
				ctx.Fail(err, v.Status(), v.Error())
			} else {
				ctx.Fail(err, 500, "proxy error")
			}
			return
		}

		if next {
			ctx.Next()
			return
		}

		zoox.WrapH(proxy.NewSingleHost(proxyCfg.Target, &proxyCfg.SingleHostConfig))(ctx)
	}
}

// DEPRECIATED

// ProxyGroupsConfig is the config of proxy middlewares
type ProxyGroupsConfig struct {
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

// ProxyGroups is a middleware that proxies the request to the backend service.
func ProxyGroups(cfg *ProxyGroupsConfig) zoox.Middleware {
	// return func(ctx *zoox.Context) {
	// 	for _, group := range cfg.Rewrites {
	// 		if matched, err := regexp.MatchString(group.RegExp, ctx.Path); err == nil && matched {
	// 			// @BUG: this is not working
	// 			p := proxy.NewSingleTarget(group.Rewrite.Target, &proxy.SingleTargetConfig{
	// 				Rewrites: group.Rewrite.Rewrites,
	// 				// ChangeOrigin: true,
	// 			})

	// 			p.ServeHTTP(ctx.Writer, ctx.Request)
	// 			return
	// 		}
	// 	}

	// 	ctx.Next()
	// }

	return ProxySingleTarget(func(ctx *zoox.Context, proxyCfg *ProxySingleTargetConfig) (next bool, err error) {
		for _, group := range cfg.Rewrites {
			if matched, err := regexp.MatchString(group.RegExp, ctx.Path); err != nil {
				return false, err
			} else if matched {
				proxyCfg.Target = group.Rewrite.Target
				proxyCfg.Rewrites = group.Rewrite.Rewrites
				return false, nil
			}
		}

		return true, nil
	})
}
