package middleware

import (
	"net/http"

	"github.com/go-zoox/zoox"
)

const DefaultRuntimePath = "/_/runtime"

// Runtime ...
func Runtime() zoox.Middleware {
	return func(ctx *zoox.Context) {
		if ctx.Path == DefaultRuntimePath {
			info := ctx.App.Runtime().Info()
			info.Version = zoox.Version

			ctx.JSON(http.StatusOK, info)
			return
		}

		ctx.Next()
	}
}
