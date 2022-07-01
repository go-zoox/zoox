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
func Recovery() zoox.HandlerFunc {
	return func(ctx *zoox.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch v := err.(type) {
				case error:
					logger.Error("%s", trace(fmt.Sprintf("%s", v)))

					ctx.Error(http.StatusInternalServerError, v.Error())
				case string:
					logger.Error("%s", v)
					ctx.Error(http.StatusInternalServerError, v)
				default:
					logger.Error("unknown error: %v", v)
					ctx.Error(http.StatusInternalServerError, "unknown error")
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
