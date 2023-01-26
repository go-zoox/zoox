package zoox

import (
	"github.com/go-zoox/jwt"
	"github.com/go-zoox/random"
)

var defaultJwtSecretKey = "go-zoox_" + random.String(24)

func newJwt(ctx *Context) jwt.Jwt {
	secretKey := defaultJwtSecretKey
	if ctx.App.SecretKey != "" {
		secretKey = ctx.App.SecretKey
	}

	return jwt.New(secretKey)
}
