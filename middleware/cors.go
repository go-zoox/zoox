// reference:
//	MDN CORS Specification: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS

package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-zoox/zoox"
)

// CorsConfig is the configuration for the CORS middleware.
type CorsConfig struct {
	IgnoreFunc       func(ctx *zoox.Context) bool
	AllowOrigins     []string
	AllowOriginFunc  func(origin string) bool
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           int64
	ExposeHeaders    []string
}

// DefaultCorsConfig is the default CORS configuration.
func DefaultCorsConfig() *CorsConfig {
	return &CorsConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{},
	}
}

// CORS is a middleware for handling CORS (Cross-Origin Resource Sharing)
func CORS(cfg ...*CorsConfig) zoox.Middleware {
	cfgX := DefaultCorsConfig()
	if len(cfg) > 0 {
		cfgX = cfg[0]
	}

	return func(ctx *zoox.Context) {
		origin := ctx.Origin()

		if cfgX.IgnoreFunc != nil && cfgX.IgnoreFunc(ctx) {
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
				if allowOrigin == "*" {
					matched = true
					break
				}

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

		isPreflight := ctx.Method == http.MethodOptions
		// not preflight
		if !isPreflight {
			// Note that simple GET requests are not preflighted, and so if a request is made for a resource with credentials,
			//	if this header is not returned with the resource, the response is ignored by the browser and not returned to web content.
			if ctx.Method == http.MethodGet && cfgX.AllowCredentials {
				ctx.Set("Access-Control-Allow-Credentials", "true")
			}

			if len(cfgX.ExposeHeaders) > 0 {
				ctx.Set("Access-Control-Expose-Headers", strings.Join(cfgX.ExposeHeaders, ","))
			}

			ctx.Next()
			return
		}

		if len(cfgX.AllowMethods) > 0 {
			ctx.Set("Access-Control-Allow-Methods", strings.Join(cfgX.AllowMethods, ","))
		}

		if len(cfgX.AllowHeaders) > 0 {
			ctx.Set("Access-Control-Allow-Headers", strings.Join(cfgX.AllowHeaders, ","))
		}

		if cfgX.MaxAge != 0 {
			ctx.Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfgX.MaxAge))
		}

		if cfgX.AllowCredentials {
			ctx.Set("Access-Control-Allow-Credentials", "true")
		}

		ctx.Next()
	}
}
