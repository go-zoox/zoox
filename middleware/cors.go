package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-zoox/zoox"
)

type CorsConfig struct {
	IgnoreFunc       func(ctx *zoox.Context) bool
	AllowOrigins     []string
	AllowOriginFunc  func(origin string) bool
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
}

func DefaultCorsConfig() *CorsConfig {
	return &CorsConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
	}
}

func CORS(cfg ...*CorsConfig) zoox.Middleware {
	cfgX := DefaultCorsConfig()
	if len(cfg) > 0 {
		cfgX = cfg[0]
	}

	return func(ctx *zoox.Context) {
		origin := ctx.Origin()

		isIgnored := false
		if cfgX.IgnoreFunc != nil {
			isIgnored = cfgX.IgnoreFunc(ctx)
		}
		isPreflight := ctx.Method == http.MethodOptions
		if !isPreflight || isIgnored {
			if len(cfgX.ExposeHeaders) > 0 {
				ctx.Set("Access-Control-Expose-Headers", strings.Join(cfgX.ExposeHeaders, ","))
			}

			ctx.Next()
			return
		}

		if cfgX.AllowOriginFunc != nil {
			if !cfgX.AllowOriginFunc(origin) {
				ctx.Status(http.StatusNoContent)
				return // skip
			}
		} else if len(cfgX.AllowOrigins) > 0 {
			var matched bool
			var err error
			for _, allowOrigin := range cfgX.AllowOrigins {
				if matched, err = regexp.MatchString(allowOrigin, origin); err == nil && matched {
					break
				}
			}

			if !matched {
				ctx.Status(http.StatusNoContent)
				return // skip
			}
		}

		ctx.Set("Access-Control-Allow-Origin", origin)

		if len(cfgX.AllowMethods) > 0 {
			ctx.Set("Access-Control-Allow-Methods", strings.Join(cfgX.AllowMethods, ","))
		}

		if len(cfgX.AllowHeaders) > 0 {
			ctx.Set("Access-Control-Allow-Headers", strings.Join(cfgX.AllowHeaders, ","))
		}

		if cfgX.AllowCredentials {
			ctx.Set("Access-Control-Allow-Credentials", "true")
		}

		ctx.Next()
	}
}
