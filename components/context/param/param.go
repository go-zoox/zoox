package param

import "github.com/go-zoox/core-utils/strings"

// Param ...
type Param interface {
	Get(key string, defaultValue ...string) strings.Value
	Iterator() map[string]string
	//
	ID() (id strings.Value, err error)
}

type param struct {
	params map[string]string
}

// New creates a param.
func New(value map[string]string) Param {
	return &param{
		params: value,
	}
}
