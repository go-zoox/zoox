package query

import (
	"net/http"

	"github.com/go-zoox/core-utils/strings"
)

// Query ...
type Query interface {
	Get(key string, defaultValue ...string) strings.Value
	//
	Page(defaultValue ...uint) uint
	PageSize(defaultValue ...uint) uint
	Where() map[string]strings.Value
	OrderBy() map[string]strings.Value
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

// Page returns the page.
func (q *query) Page(defaultValue ...uint) uint {
	if v := q.Get("page").UInt(); v != 0 {
		return v
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 1
}

// PageSize returns the page size.
// If the page size is not set, it returns 10.
func (q *query) PageSize(defaultValue ...uint) uint {
	if v := q.Get("page_size").UInt(); v != 0 {
		return v
	}

	if v := q.Get("pageSize").UInt(); v != 0 {
		return v
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 10
}

// Where returns the where.
func (q *query) Where() map[string]strings.Value {
	// where := make(map[string]strings.Value)
	// ignoreKeys := map[string]bool{
	// 	"page":      true,
	// 	"page_size": true,
	// 	"pageSize":  true,
	// 	"order_by":  true,
	// 	"orderBy":   true,
	// }

	// values := q.request.URL.Query()
	// for key, value := range values {
	// 	if ok := ignoreKeys[key]; ok {
	// 		continue
	// 	}

	// 	where[key] = strings.Value(value[0])
	// }

	// return where

	panic("not implemented")
}

// OrderBy returns the order by.
func (q *query) OrderBy() map[string]strings.Value {
	panic("not implemented")
}
