package zoox

import "github.com/go-zoox/core-utils/strings"

// Query ...
type Query interface {
	Get(key string, defaultValue ...string) strings.Value
}

type query struct {
	ctx *Context
}

func newQuery(ctx *Context) *query {
	return &query{
		ctx: ctx,
	}
}

// Get gets request query with the given name.
func (q *query) Get(key string, defaultValue ...string) strings.Value {
	value := q.ctx.Request.URL.Query().Get(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return strings.Value(value)
}
