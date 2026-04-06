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

	// EnableH2C enables cleartext HTTP/2 (h2c) on the plaintext TCP HTTP server. Use only on trusted networks.
	EnableH2C bool
	// EnableHTTP3 starts an HTTP/3 (QUIC) server on UDP when HTTPS is configured (HTTPSPort non-zero, TCP).
	EnableHTTP3 bool
	// HTTP3Port is the UDP listen port for HTTP/3. Zero means the same port as HTTPSPort.
	HTTP3Port int
	// HTTP3AltSvcMaxAge is the ma= parameter for the Alt-Svc header on HTTPS responses when HTTP/3 is enabled.
	// Zero uses 86400. Negative values disable the Alt-Svc header.
	HTTP3AltSvcMaxAge int

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

	// Logger
	Logger Logger `config:"logger"`
}
