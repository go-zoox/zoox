package config

// Logger ...
type Logger struct {
	Middleware `config:"middleware"`
}

// Middleware ...
type Middleware struct {
	Disabled bool `config:"disabled"`
}
