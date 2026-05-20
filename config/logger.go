package config

// Logger configures the application logger and its transports.
type Logger struct {
	// Level is the minimum log level (debug, info, warn, error).
	Level string `config:"level"`

	// Middleware is the built-in HTTP request logger middleware.
	Middleware `config:"middleware"`

	// Transports lists extra sinks (e.g. file). Zoox always includes console;
	// configured transports are stacked on top. Empty = console only.
	Transports []Transport `config:"transports"`
}

// Transport is one logger sink (see components/application/logger).
type Transport struct {
	// Type is the transport name: "console" or "file".
	Type string `config:"type"`
	// Path is the default log file when Type is "file".
	Path string `config:"path"`
	// Levels maps log level names to file paths when Type is "file".
	Levels map[string]string `config:"levels"`
}

// Middleware ...
type Middleware struct {
	// Disabled is the logger middleware disabled.
	Disabled bool `config:"disabled"`
}
