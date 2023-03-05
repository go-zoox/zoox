package middleware

import (
	"github.com/go-zoox/zoox"
)

// BearerToken is a middleware that authenticates via Bearer Token.
func BearerToken(tokens []string) zoox.Middleware {
	return func(ctx *zoox.Context) {
		tokenX, ok := ctx.BearerToken()
		if !ok {
			ctx.JSON(401, zoox.H{
				"code":    401001,
				"message": "unauthorized (no token found)",
			})
			return
		}

		for _, token := range tokens {
			if tokenX == token {
				ctx.Next()
				return
			}
		}

		ctx.JSON(401, zoox.H{
			"code":    401002,
			"message": "unauthorized (invalid token)",
		})
	}
}
