package cmd

import (
	"context"
	"fmt"

	"github.com/go-zoox/command"
	"github.com/go-zoox/command/config"
)

// Cmd ...
type Cmd interface {
	Create(cfg *config.Config) (command.Command, error)
}

type cmd struct {
	ctx context.Context
}

// New creates a command.
func New(ctx context.Context) Cmd {
	return &cmd{
		ctx: ctx,
	}
}

// Create creates a command.
func (c *cmd) Create(cfg *config.Config) (command.Command, error) {
	if cfg == nil {
		return nil, fmt.Errorf("command config is nil")
	}

	if cfg.Context == nil {
		cfg.Context = c.ctx
	}

	return command.New(cfg)
}
