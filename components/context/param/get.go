package param

import "github.com/go-zoox/core-utils/strings"

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
