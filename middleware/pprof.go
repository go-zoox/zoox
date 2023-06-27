package middleware

import (
	"net/http/pprof"

	"github.com/go-zoox/core-utils/strings"
	"github.com/go-zoox/zoox"
)

// DefaultPProfPath ...
const DefaultPProfPath = "/_/pprof"

// PProf ...
func PProf() zoox.Middleware {
	return func(ctx *zoox.Context) {
		if strings.StartsWith(ctx.Path, DefaultPProfPath) {
			relativePath := ctx.Path[len(DefaultPProfPath):]
			switch relativePath {
			case "/", "":
				zoox.WrapF(pprof.Index)(ctx)
			case "/cmdline":
				zoox.WrapF(pprof.Cmdline)(ctx)
			case "/profile":
				zoox.WrapF(pprof.Profile)(ctx)
			case "/symbol":
				zoox.WrapF(pprof.Symbol)(ctx)
			case "/trace":
				zoox.WrapF(pprof.Trace)(ctx)
			case "/allocs":
				zoox.WrapF(pprof.Handler("allocs").ServeHTTP)(ctx)
			case "/block":
				zoox.WrapF(pprof.Handler("block").ServeHTTP)(ctx)
			case "/goroutine":
				zoox.WrapF(pprof.Handler("goroutine").ServeHTTP)(ctx)
			case "/heap":
				zoox.WrapF(pprof.Handler("heap").ServeHTTP)(ctx)
			case "/mutex":
				zoox.WrapF(pprof.Handler("mutex").ServeHTTP)(ctx)
			case "/threadcreate":
				zoox.WrapF(pprof.Handler("threadcreate").ServeHTTP)(ctx)
			default:
				ctx.Status(404)
			}
			return
		}

		ctx.Next()
	}
}
