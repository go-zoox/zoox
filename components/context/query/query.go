package query

import (
	"net/http"

	"github.com/go-zoox/core-utils/strings"
)

// Query ...
type Query interface {
	Get(key string, defaultValue ...string) strings.Value
}

type query struct {
	request *http.Request
}

// New creates a query.
func New(request *http.Request) Query {
	return &query{
		request: request,
	}
}

// Get gets request query with the given name.
func (q *query) Get(key string, defaultValue ...string) strings.Value {
	value := q.request.URL.Query().Get(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return strings.Value(value)
}
