package middleware

import (
	"github.com/go-zoox/jwt"
	"github.com/go-zoox/zoox"
)

// Jwt is a middleware that authenticates via JWT.
func Jwt(secret string, algorithm ...string) zoox.Middleware {
	algorithmX := "HS256"
	if len(algorithm) > 0 {
		algorithmX = algorithm[0]
	}

	var j *jwt.Jwt
	switch algorithmX {
	case "HS256":
		j = jwt.NewHS256(secret)
	case "HS384":
		j = jwt.NewHS384(secret)
	case "HS512":
		j = jwt.NewHS512(secret)
	default:
		panic("unknown algorithm, allowed: HS256, HS384, HS512")
	}

	return func(ctx *zoox.Context) {
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			authHeader = ctx.Query().Get("access_token")
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

		if err := j.Verify(token); err != nil {
			ctx.Status(401)
			return
		}

		ctx.Next()
	}
}
