package middleware

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/go-errors/errors"
	"github.com/go-zoox/zoox"
)

// Recovery is the recovery middleware
func Recovery() zoox.Middleware {
	return func(ctx *zoox.Context) {
		defer func() {
			if err := recover(); err != nil {
				// stackoverflow: https://stackoverflow.com/questions/52103182/how-to-get-the-stacktrace-of-a-panic-and-store-as-a-variable
				if ctx.Debug().IsDebugMode() {
					fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
				}

				httprequest, _ := httputil.DumpRequest(ctx.Request, false)
				goErr := errors.Wrap(err, 3)
				reset := string([]byte{27, 91, 48, 109})
				ctx.Logger.Errorf("[Nice Recovery] panic recovered:\n\n%s%s\n\n%s%s", httprequest, goErr.Error(), goErr.Stack(), reset)

				switch err.(type) {
				case error:
					ctx.Error(http.StatusInternalServerError, "Internal Server Error")
				case string:
					ctx.Error(http.StatusInternalServerError, "Internal Server Error")
				default:
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
