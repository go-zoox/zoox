package middleware

import (
	"time"

	"github.com/go-zoox/core-utils/regexp"
	"github.com/go-zoox/zoox"
)

// DefaultStaticCacheMaxAge ...
const DefaultStaticCacheMaxAge = 365 * 24 * time.Hour

// DefaultStaticCacheRegExp ...
const DefaultStaticCacheRegExp = "\\.(css|js|ico|jpg|png|jpeg|webp|gif|socket|ws|map|webmanifest)$"

// StaticCacheConfig ...
type StaticCacheConfig struct {
	// MaxAge is the duration that client caches the static file.
	// Default is 365 days.
	MaxAge time.Duration

	// RegExp is the regular expression that matches the static file.
	// Default is "\\.(css|js|ico|jpg|png|jpeg|webp|gif|socket|ws|map|webmanifest)$"
	RegExp string
}

// StaticCache is a middleware that adds a "Cache-Control" header to the request.
func StaticCache(cfg ...*StaticCacheConfig) zoox.Middleware {
	cfgX := &StaticCacheConfig{}
	if len(cfg) > 0 && cfg[0] != nil {
		cfgX = cfg[0]
	}

	staticFileMaxAge := DefaultStaticCacheMaxAge
	if cfgX.MaxAge != 0 {
		staticFileMaxAge = cfgX.MaxAge
	}

	staticFileRegExp := DefaultStaticCacheRegExp
	if cfgX.RegExp != "" {
		staticFileRegExp = cfgX.RegExp
	}

	staticFileRe, err := regexp.New(staticFileRegExp)
	if err != nil {
		panic(err)
	}

	isStaticFile := func(path string) bool {
		return staticFileRe.Match(path)
	}

	return func(ctx *zoox.Context) {
		if isStaticFile(ctx.Path) {
			ctx.SetCacheControlWithMaxAge(staticFileMaxAge)
		}

		ctx.Next()

		// @TODO zoox middleware not work here
		// if isStaticFile(ctx.Path) {
		// 	ctx.SetCacheControlWithMaxAge(staticFileMaxAge)
		// }
	}
}
