package zoox

import "os"

// Env ...
type Env struct {
}

func newEnv() *Env {
	return &Env{}
}

// Get ...
func (e *Env) Get(key string) string {
	return os.Getenv(key)
}
