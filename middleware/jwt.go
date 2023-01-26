package middleware

import (
	"net/http"

	"github.com/go-zoox/zoox"
)

// Jwt is a middleware that authenticates via JWT.
func Jwt() zoox.Middleware {
	return func(ctx *zoox.Context) {
		isUnauthorized := false
		reason := ""

		token, ok := ctx.BearerToken()
		if !ok {
			token = ctx.Query().Get("access_token").ToString()
		}

		signer := ctx.Jwt()
		if token == "" {
			isUnauthorized = true
			reason = "token not found"
		} else if _, err := signer.Verify(token); err != nil {
			isUnauthorized = true
			reason = "token invalid"
		}

		if isUnauthorized {
			if ctx.AcceptJSON() {
				ctx.JSON(http.StatusUnauthorized, zoox.H{
					"code":    401,
					"message": reason,
				})
			} else {
				ctx.Status(401)
			}
			return
		}

		ctx.Next()
	}
}
