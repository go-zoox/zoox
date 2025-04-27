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
	Where() *Where
	OrderBy() *OrderBy
	//
	ID() (id strings.Value, err error)
	MustID() (id strings.Value)
	AccessToken() (accessToken string, err error)
	MustAccessToken() (accessToken string)
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

// // Where returns the where.
// func (q *query) Where() map[string]strings.Value {
// 	// where := make(map[string]strings.Value)
// 	// ignoreKeys := map[string]bool{
// 	// 	"page":      true,
// 	// 	"page_size": true,
// 	// 	"pageSize":  true,
// 	// 	"order_by":  true,
// 	// 	"orderBy":   true,
// 	// }

// 	// values := q.request.URL.Query()
// 	// for key, value := range values {
// 	// 	if ok := ignoreKeys[key]; ok {
// 	// 		continue
// 	// 	}

// 	// 	where[key] = strings.Value(value[0])
// 	// }

// 	// return where

// 	panic("not implemented")
// }

// // OrderBy returns the order by.
// func (q *query) OrderBy() map[string]strings.Value {
// 	panic("not implemented")
// }
