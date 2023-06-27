package zoox

import (
	"net/http"
)

// WrapH wraps a http.Handler to a HandlerFunc
func WrapH(handler http.Handler) HandlerFunc {
	return func(ctx *Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

// WrapF wraps a http.HandlerFunc to a HandlerFunc
func WrapF(handler http.HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		handler(ctx.Writer, ctx.Request)
	}
}
