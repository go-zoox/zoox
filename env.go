package zoox

import "os"

// Env ...
type Env interface {
	Get(key string) string
}

type env struct {
}

func newEnv() *env {
	return &env{}
}

// Get ...
func (e *env) Get(key string) string {
	return os.Getenv(key)
}
