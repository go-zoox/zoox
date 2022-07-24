package zoox

import (
	"github.com/go-zoox/debug"
)

// DebugEnv is the environment variable name for debug.
const DebugEnv = "GO_ZOOX_DEBUG"

// Debug ...
type Debug struct {
	core debug.Debugger
}

func newDebug(app *Application) *Debug {
	core := debug.New(DebugEnv, func(args ...interface{}) error {
		app.Logger.Debug(args[0].(string), args[1:]...)
		return nil
	})

	return &Debug{
		core: core,
	}
}

// Info logs debug info.
func (c *Debug) Info(args ...interface{}) {
	c.core.Info(args...)
}
