package session

import (
	"time"

	"github.com/go-zoox/cookie"
	"github.com/go-zoox/random"
	gosession "github.com/go-zoox/session"
)

var defaultSessionSecretKey = "go-zoox_" + random.String(24)

func New(cookie cookie.Cookie, secretKey string, maxAge time.Duration) gosession.Session {
	if secretKey == "" {
		secretKey = defaultSessionSecretKey
	}

	return gosession.New(cookie, secretKey, &gosession.Config{
		MaxAge: maxAge,
	})
}
