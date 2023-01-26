package zoox

import "github.com/go-zoox/core-utils/object"

// Body ...
type Body interface {
	Get(key string, defaultValue ...interface{}) interface{}
}

// body ...
type body struct {
	ctx *Context
	//
	data map[string]interface{}
}

func newBody(ctx *Context) Body {
	return &body{
		ctx: ctx,
	}
}

// Get gets request form with the given name.
func (f *body) Get(key string, defaultValue ...interface{}) interface{} {
	if f.data == nil {
		f.data = f.ctx.Bodies()
	}

	value := object.Get(f.data, key)

	// @TODO generic cannot compare zero value
	// if value == "" && len(defaultValue) > 0 {
	// 	value = defaultValue[0]
	// }

	return value
}
