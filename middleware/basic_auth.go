package middleware

import (
	"crypto/subtle"

	"github.com/go-zoox/zoox"
)

func BasicAuth(realm string, credentials map[string]string) zoox.Middleware {
	return func(ctx *zoox.Context) {
		user, pass, ok := ctx.Request.BasicAuth()
		if !ok {
			ctx.Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			ctx.Status(401)
			return
		}

		credPass, credUserOk := credentials[user]
		if !credUserOk || subtle.ConstantTimeCompare([]byte(pass), []byte(credPass)) != 1 {
			ctx.Status(401)
			return
		}

		ctx.Next()
	}
}
