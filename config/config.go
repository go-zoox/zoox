package config

import (
	"github.com/go-zoox/cache"
	"github.com/go-zoox/session"
)

// Config defines the config of zoox.Application.
type Config struct {
	Protocol  string
	Host      string
	Port      int
	HTTPSPort int

	// BodySizeLimit is the limit of the request body size.
	BodySizeLimit int64

	//
	NetworkType      string
	UnixDomainSocket string

	// TLS
	// TLS Certificate
	TLSCertFile string
	// TLS Private Key
	TLSKeyFile string
	// TLS Ca Certificate
	TLSCaCertFile string
	//
	TLSCert string
	TLSKey  string

	//
	LogLevel string `config:"log_level"`
	//
	SecretKey string `config:"secret_key"`
	//
	Session session.Config `config:"session"`
	//
	Cache cache.Config `config:"cache"`
	//
	Redis Redis `config:"redis"`
	//
	Banner string
	//
	Monitor Monitor `config:"monitor"`
}
