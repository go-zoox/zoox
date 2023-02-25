package debug

import (
	godebug "github.com/go-zoox/debug"
	"github.com/go-zoox/logger"
)

// DebugEnv is the environment variable name for debug.
const DebugEnv = "GO_ZOOX_DEBUG"

// Debug ...
type Debug interface {
	Info(args ...interface{})
	IsDebugMode() bool
}

type debug struct {
	core godebug.Debugger
}

func New(logger *logger.Logger) Debug {
	core := godebug.New(DebugEnv, func(args ...interface{}) error {
		logger.Debug(args[0].(string), args[1:]...)
		return nil
	})

	return &debug{
		core: core,
	}
}

// Info logs debug info.
func (c *debug) Info(args ...interface{}) {
	c.core.Debug(args...)
}

// Info logs debug info.
func (c *debug) IsDebugMode() bool {
	return c.core.IsDebugMode()
}
