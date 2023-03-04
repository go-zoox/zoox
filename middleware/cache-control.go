package middleware

import (
	"fmt"
	"time"

	"github.com/go-zoox/core-utils/regexp"
	"github.com/go-zoox/headers"
	"github.com/go-zoox/zoox"
)

// CacheControlConfig ...
type CacheControlConfig struct {
	Paths  []string
	MaxAge time.Duration
	//
	Items []*CacheControlItem
}

type CacheControlItem struct {
	Path   regexp.RegExp
	MaxAge time.Duration
}

// CacheControl is a middleware that adds a "Cache-Control" header to the request.
func CacheControl(cfg *CacheControlConfig) zoox.Middleware {
	for _, path := range cfg.Paths {
		re, err := regexp.New(path)
		if err != nil {
			panic(err)
		}

		cfg.Items = append(cfg.Items, &CacheControlItem{
			Path:   re,
			MaxAge: cfg.MaxAge,
		})
	}

	return func(ctx *zoox.Context) {
		if ctx.Method != "GET" {
			ctx.Next()
			return
		}

		if cfg.Items != nil {
			for _, item := range cfg.Items {
				if item.Path.Match(ctx.Path) {
					maxAge := cfg.MaxAge / time.Second
					// ctx.Logger.Infof("[middleware][cache-control] hit path: %s, max-age: %d", ctx.Path, maxAge)
					ctx.Set(headers.CacheControl, fmt.Sprintf("public, max-age=%d", maxAge))
					break
				}
			}
		}

		ctx.Next()
	}
}
