package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox"
)

// Recovery is the recovery middleware
func Recovery() zoox.Middleware {
	return func(ctx *zoox.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch v := err.(type) {
				case error:
					logger.Errorf("[recovery][%s %s] %s", ctx.Method, ctx.Path, (fmt.Sprintf("%s", v)))

					ctx.Error(http.StatusInternalServerError, "Internal Server Error")
				case string:
					logger.Errorf("[recovery][%s %s] %s", ctx.Method, ctx.Path, v)
					ctx.Error(http.StatusInternalServerError, "Internal Server Error")
				default:
					logger.Errorf("[recovery][%s %s] unknown error: %#v", ctx.Method, ctx.Path, v)
					ctx.Error(http.StatusInternalServerError, "Internal Server Error")
				}
			}
		}()

		ctx.Next()
	}
}

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n  %s:%d", file, line))
	}

	return str.String()
}
