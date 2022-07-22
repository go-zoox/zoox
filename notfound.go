package zoox

import (
	"net/http"
)

// NotFound returns a HandlerFunc that replies with a 404 not found
func NotFound() HandlerFunc {
	return func(ctx *Context) {
		ctx.Error(http.StatusNotFound, "Not Found")
	}
}
