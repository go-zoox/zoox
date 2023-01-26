package zoox

import "github.com/go-zoox/core-utils/strings"

// Param ...
type Param interface {
	Get(key string, defaultValue ...string) strings.Value
	Iterator() map[string]string
}

type param struct {
	ctx *Context
	//
	params map[string]string
}

func newParams(ctx *Context, value map[string]string) Param {
	return &param{
		ctx:    ctx,
		params: value,
	}
}

// Get gets request param with the given name.
func (q *param) Get(key string, defaultValue ...string) strings.Value {
	value, ok := q.params[key]
	if ok {
		return strings.Value(value)
	}

	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return strings.Value(value)
}

// Iterator ...
func (q *param) Iterator() map[string]string {
	return q.params
}
