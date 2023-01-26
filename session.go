package zoox

import (
	"github.com/go-zoox/random"
	"github.com/go-zoox/session"
)

var defaultSessionSecretKey = "go-zoox_" + random.String(24)

func newSession(ctx *Context) session.Session {
	secretKey := defaultSessionSecretKey
	if ctx.App.SecretKey != "" {
		secretKey = ctx.App.SecretKey
	}

	return session.New(ctx.Cookie(), secretKey)
}
