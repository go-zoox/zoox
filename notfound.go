package zoox

import "net/http"

// NotFound returns a HandlerFunc that replies with a 404 not found
func NotFound() HandlerFunc {
	return func(ctx *Context) {
		ctx.String(http.StatusNotFound, "404 NOT Found: %s\n", ctx.Path)
	}
}
