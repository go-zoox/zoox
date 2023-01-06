package defaults

import (
	"github.com/go-zoox/zoox"
)

// Default returns a new default zoox.
func Default() *zoox.Application {
	return Defaults()
}
