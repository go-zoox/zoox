package zoox

import (
	"github.com/go-zoox/cookie"
)

func newCookie(ctx *Context) cookie.Cookie {
	return cookie.New(
		ctx.Writer,
		ctx.Request,
	)
}
