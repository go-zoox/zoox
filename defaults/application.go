package defaults

import (
	"github.com/go-zoox/zoox"
)

// Application returns a new default zoox.
func Application() *zoox.Application {
	return Defaults()
}
