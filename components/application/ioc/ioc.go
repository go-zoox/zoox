package ioc

import (
	goioc "github.com/go-zoox/ioc"
)

// IoC ...
type IoC interface {
	Controller() goioc.Container
	Service() goioc.Container
	Model() goioc.Container
	//
	Create() goioc.Container
}

type container struct {
	controller goioc.Container
	service    goioc.Container
	model      goioc.Container
}

// IoC ...
func New() IoC {
	return &container{
		controller: goioc.New(),
		service:    goioc.New(),
		model:      goioc.New(),
	}
}

// Controller ...
func (c *container) Controller() goioc.Container {
	return c.controller
}

// Service ...
func (c *container) Service() goioc.Container {
	return c.service
}

// Model ...
func (c *container) Model() goioc.Container {
	return c.model
}

// Create ...
func (c *container) Create() goioc.Container {
	return goioc.New()
}
