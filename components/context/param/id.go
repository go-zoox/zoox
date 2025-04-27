package param

import (
	"fmt"

	"github.com/go-zoox/core-utils/strings"
)

// ConstantsParamsIDKeys is the keys that are used to identify the id.
var ConstantsParamsIDKeys = []string{
	"id",
	"_id",
}

// ID returns the id.
func (p *param) ID() (id strings.Value, err error) {
	for _, key := range ConstantsParamsIDKeys {
		if v := p.Get(key); v != "" {
			return v, nil
		}
	}

	return "", fmt.Errorf("id not found")
}

// MustID returns the id.
func (p *param) MustID() (id strings.Value) {
	id, err := p.ID()
	if err != nil {
		panic(err)
	}

	return id
}
