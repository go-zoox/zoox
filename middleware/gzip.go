// Modified from: https://github.com/gin-contrib/gzip

package middleware

import (
	"github.com/go-zoox/gzip"

	"github.com/go-zoox/zoox"
)

// GzipConfig is the configuration for gzip middleware.
type GzipConfig struct {
	Level int
	// Options  gzip.Options
	OptionFn gzip.Option
}

// Gzip is a gzip moddleware for zoox.
func Gzip(cfg ...*GzipConfig) zoox.Middleware {
	level := gzip.DefaultCompression
	var optionFn gzip.Option
	if len(cfg) > 0 {
		level = cfg[0].Level
		optionFn = cfg[0].OptionFn
	}

	return gzip.Gzip(level, optionFn)
}
