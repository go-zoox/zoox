package config

// Logger ...
type Logger struct {
	// Level is the log level.
	Level string `config:"level"`

	// Middleware is the logger middleware.
	Middleware `config:"middleware"`
}

// Middleware ...
type Middleware struct {
	// Disabled is the logger middleware disabled.
	Disabled bool `config:"disabled"`
}
