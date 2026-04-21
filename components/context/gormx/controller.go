package gormx

import (
	gormxlib "github.com/go-zoox/gormx"
	"github.com/go-zoox/zoox"
)

// Controller is the zoox-specific controller abstraction for gormx params.
type Controller interface {
	Name() string
	Params(ctx *zoox.Context) *gormxlib.Params
}

// ControllerImpl is the default implementation of Controller.
type ControllerImpl struct{}

// Params returns gormx params from zoox context.
func (c *ControllerImpl) Params(ctx *zoox.Context) *gormxlib.Params {
	return NewParams(ctx)
}
