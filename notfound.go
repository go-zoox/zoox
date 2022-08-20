package zoox

import (
	"net/http"
	"time"
)

// NotFound returns a HandlerFunc that replies with a 404 not found
func NotFound() HandlerFunc {
	return func(ctx *Context) {
		// api 405
		if ctx.AcceptJSON() {
			ctx.JSON(http.StatusMethodNotAllowed, H{
				"code":      405,
				"message":   "Method not allowed",
				"method":    ctx.Method,
				"path":      ctx.Path,
				"timestamp": time.Now(),
			})
			return
		}

		// page 404
		ctx.Error(http.StatusNotFound, "Not Found")
	}
}
