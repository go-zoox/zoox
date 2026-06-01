package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-zoox/proxy"
	"github.com/go-zoox/proxy/utils/rewriter"
	"github.com/go-zoox/zoox"
)

// ProxyConfig defines the proxy config
type ProxyConfig struct {
	proxy.Config

	ErrorPages ProxyErrorPages
}

// ProxyErrorPages defines the error pages
type ProxyErrorPages struct {
	NotFound             string
	InternalServiceError string
	BadGateway           string
	ServiceUnavailable   string
	GatewayTimeout       string
	// ContentType overrides the default text/html response type (e.g. application/json).
	ContentType string
}

func writeProxyErrorResponse(ctx *zoox.Context, status int, body, contentType string) {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "application/json") {
		if ct := strings.TrimSpace(contentType); ct != "" {
			ctx.Data(status, ct, []byte(body))
			return
		}
		ctx.Data(status, "application/json; charset=utf-8", []byte(body))
		return
	}
	ctx.HTML(status, body)
}

// Proxy is a middleware that proxies the request.
func Proxy(fn func(ctx *zoox.Context, cfg *ProxyConfig) (next, stop bool, err error)) zoox.Middleware {
	return func(ctx *zoox.Context) {
		cfg := &ProxyConfig{}
		if ctx.App.Config.TrustProxy {
			cfg.Config.TrustProxy = true
		}
		next, stop, err := fn(ctx, cfg)
		if err != nil {
			ctx.Logger.Errorf("[middleware.proxy] proxy error: %#v", err)
			if v, ok := err.(*proxy.HTTPError); ok {
				html := v.Error()
				switch v.Status() {
				case http.StatusNotFound:
					if cfg.ErrorPages.NotFound != "" {
						html = cfg.ErrorPages.NotFound
					}
				case http.StatusBadGateway:
					if cfg.ErrorPages.BadGateway != "" {
						html = cfg.ErrorPages.BadGateway
					} else if cfg.ErrorPages.InternalServiceError != "" {
						html = cfg.ErrorPages.InternalServiceError
					}
				case http.StatusServiceUnavailable:
					if cfg.ErrorPages.ServiceUnavailable != "" {
						html = cfg.ErrorPages.ServiceUnavailable
					} else if cfg.ErrorPages.InternalServiceError != "" {
						html = cfg.ErrorPages.InternalServiceError
					}
				case http.StatusGatewayTimeout:
					if cfg.ErrorPages.GatewayTimeout != "" {
						html = cfg.ErrorPages.GatewayTimeout
					} else if cfg.ErrorPages.InternalServiceError != "" {
						html = cfg.ErrorPages.InternalServiceError
					}
				}

				writeProxyErrorResponse(ctx, v.Status(), html, cfg.ErrorPages.ContentType)
			} else {
				html := v.Error()
				if cfg.ErrorPages.InternalServiceError != "" {
					html = cfg.ErrorPages.InternalServiceError
				}

				writeProxyErrorResponse(ctx, http.StatusInternalServerError, html, cfg.ErrorPages.ContentType)
			}
			return
		}

		if stop {
			return
		}

		if next {
			ctx.Next()
			return
		}

		zoox.WrapH(proxy.New(&cfg.Config))(ctx)
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
		if ctx.App.Config.TrustProxy {
			proxyCfg.SingleHostConfig.TrustProxy = true
		}
		next, err := fn(ctx, proxyCfg)
		if err != nil {
			if v, ok := err.(*proxy.HTTPError); ok {
				ctx.Fail(err, v.Status(), v.Error(), v.Status())
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
