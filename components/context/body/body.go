package body

import "github.com/go-zoox/core-utils/object"

// Body ...
type Body interface {
	Get(key string, defaultValue ...interface{}) interface{}
}

// body ...
type body struct {
	getter func() map[string]any
	//
	data map[string]interface{}
}

// New creates a body.
func New(getter func() map[string]any) Body {
	return &body{
		getter: getter,
	}
}

// Get gets request form with the given name.
func (f *body) Get(key string, defaultValue ...interface{}) interface{} {
	if f.data == nil {
		f.data = f.getter()
	}

	value := object.Get(f.data, key)

	// @TODO generic cannot compare zero value
	// if value == "" && len(defaultValue) > 0 {
	// 	value = defaultValue[0]
	// }

	return value
}
