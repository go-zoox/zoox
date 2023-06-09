package env

import "os"

// Env ...
type Env interface {
	Get(key string) string
}

type env struct {
}

// New ...
func New() Env {
	return &env{}
}

// Get ...
func (e *env) Get(key string) string {
	return os.Getenv(key)
}
