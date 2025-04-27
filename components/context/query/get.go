package query

import "github.com/go-zoox/core-utils/strings"

// Get gets request query with the given name.
func (q *query) Get(key string, defaultValue ...string) strings.Value {
	value := q.request.URL.Query().Get(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return strings.Value(value)
}
