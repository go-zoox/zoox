package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/go-zoox/fs"
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

				funcName := "unknown"
				// get panic error occurred file and line
				pc, filepath, line, ok := runtime.Caller(2)
				if ok {
					filepath = filepath[len(fs.CurrentDir())+1:]
					funcName = runtime.FuncForPC(pc).Name()
					funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
					if len(funcNameParts) > 0 {
						funcName = funcNameParts[len(funcNameParts)-1]
					}
				}

				switch v := err.(type) {
				case error:
					ctx.Logger.Errorf("[recovery][%s:%d,%s][%s %s] %s", filepath, line, funcName, ctx.Method, ctx.Path, (fmt.Sprintf("%s", v)))

					ctx.Error(http.StatusInternalServerError, "Internal Server Error")
				case string:
					ctx.Logger.Errorf("[recovery][%s:%d,%s][%s %s] %s", filepath, line, funcName, ctx.Method, ctx.Path, v)
					ctx.Error(http.StatusInternalServerError, "Internal Server Error")
				default:
					ctx.Logger.Errorf("[recovery][%s:%d,%s][%s %s] unknown error: %#v (stack: %s)", filepath, line, funcName, ctx.Method, ctx.Path, v, string(debug.Stack()))
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
