package query

import (
	"fmt"

	"github.com/go-zoox/core-utils/strings"
)

// ConstantsQueryIDKeys is the keys that are used to identify the id.
var ConstantsQueryIDKeys = []string{
	"id",
	"_id",
}

// ID returns the id.
func (q *query) ID() (id strings.Value, err error) {
	for _, key := range ConstantsQueryIDKeys {
		if v := q.Get(key); v != "" {
			return v, nil
		}
	}

	return "", fmt.Errorf("id not found")
}

// MustID returns the id.
func (q *query) MustID() (id strings.Value) {
	id, err := q.ID()
	if err != nil {
		panic(err)
	}
	return id
}
