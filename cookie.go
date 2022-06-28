package zoox

import (
	"net/http"
	"time"
)

// Cookie is a middleware for handling cookie.
type Cookie struct {
	ctx *Context
}

func newCookie(ctx *Context) *Cookie {
	return &Cookie{
		ctx: ctx,
	}
}

// Set sets response cookie with the given name and value.
func (c *Cookie) Set(name string, value string, maxAge time.Duration) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(maxAge),
		HttpOnly: true,
	}
	http.SetCookie(c.ctx.Writer, cookie)
}

// Get gets request cookie with the given name.
func (c *Cookie) Get(name string) string {
	cookie, err := c.ctx.Request.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

// Del deletes response cookie with the given name.
func (c *Cookie) Del(name string) {
	c.Set(name, "", -1)
}
