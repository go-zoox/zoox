package session

import (
	"github.com/go-zoox/cookie"
	"github.com/go-zoox/random"
	"github.com/go-zoox/session"
)

var defaultSessionSecretKey = "go-zoox_" + random.String(24)

func New(cookie cookie.Cookie, secretKey string) session.Session {
	if secretKey == "" {
		secretKey = defaultSessionSecretKey
	}

	return session.New(cookie, secretKey)
}
