package middleware

import (
	"github.com/go-zoox/crypto/jwt"
	"github.com/go-zoox/zoox"
)

// Jwt is a middleware that authenticates via JWT.
func Jwt(secret string, opts ...*jwt.Options) zoox.Middleware {
	signer := jwt.New(secret, opts...)

	return func(ctx *zoox.Context) {
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			authHeader = ctx.Query().Get("access_token").ToString()
		}

		if authHeader == "" {
			ctx.Status(401)
			return
		}

		token := authHeader[7:]
		if token == "" {
			ctx.Status(401)
			return
		}

		if _, err := signer.Verify(token); err != nil {
			ctx.Status(401)
			return
		}

		ctx.Next()
	}
}
