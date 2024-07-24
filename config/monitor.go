package config

import "time"

// Monitor defines the config of monitor.
type Monitor struct {
	Prometheus `config:"prometheus"`

	Sentry `config:"sentry"`
}

// Prometheus ...
type Prometheus struct {
	Enabled bool   `config:"enabled"`
	Path    string `config:"path"`
}

// Sentry ...
type Sentry struct {
	Enabled bool `config:"enabled"`
	//
	DSN   string `config:"dsn"`
	Debug bool   `config:"debug"`
	//
	WaitForDelivery bool          `config:"wait_for_delivery"`
	Timeout         time.Duration `config:"timeout"`
}
