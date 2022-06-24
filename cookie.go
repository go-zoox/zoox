package zoox

import (
	"fmt"
	"time"
)

// Cookie is a middleware for handling cookie.
type Cookie struct {
	ctx *Context
}

// Set sets response cookie with the given name and value.
func (c *Cookie) Set(name string, value string, maxAge time.Duration) {
	expires := time.Now().Add(maxAge)

	c.ctx.SetHeader(
		"Set-Cookie",
		fmt.Sprintf("%s=%s; path=/; expires=%s; httponly", name, value, expires),
	)
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
