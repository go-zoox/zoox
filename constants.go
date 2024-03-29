package zoox

import (
	"fmt"
	"time"

	"github.com/go-zoox/random"
)

// DefaultMiddlewares is the default global middleware
var DefaultMiddlewares = map[string]func(app *Application){
	// Logger,
}

// DefaultMiddleware ...
func DefaultMiddleware(name string, fn func(app *Application)) {
	DefaultMiddlewares[name] = fn
}

// DefaultGroupsFns ...
var DefaultGroupsFns = map[string]func(r *RouterGroup){}

// DefaultGroup ...
func DefaultGroup(prefix string, fn func(r *RouterGroup)) {
	if _, ok := DefaultGroupsFns[prefix]; ok {
		panic(fmt.Errorf("zoox: default group (%s) already registered", prefix))
	}

	DefaultGroupsFns[prefix] = fn
}

// DefaultSecretKey uses for session encryption and decryption.
var DefaultSecretKey = random.String(16)

// DefaultSessionMaxAge is the default session max age.
var DefaultSessionMaxAge = 1 * 24 * time.Hour

// BuiltInEnv is the built-in environment variable.
var (
	BuiltInEnvPort      = "PORT"
	BuiltInEnvHTTPSPort = "HTTPS_PORT"
	BuiltInEnvMode      = "MODE"

	BuiltInEnvLogLevel = "LOG_LEVEL"

	BuiltInEnvSecretKey = "SECRET_KEY"

	BuiltInEnvSessionMaxAge = "SESSION_MAX_AGE"

	BuiltInEnvRedisHost = "REDIS_HOST"
	BuiltInEnvRedisPort = "REDIS_PORT"
	BuiltInEnvRedisUser = "REDIS_USER"
	BuiltInEnvRedisPass = "REDIS_PASS"
	BuiltInEnvRedisDB   = "REDIS_DB"

	BuiltInEnvMonitorPrometheusEnabled = "MONITOR_PROMETHEUS_ENABLED"
	BuiltInEnvMonitorPrometheusPath    = "MONITOR_PROMETHEUS_PATH"

	BuiltInEnvMonitorSentryEnabled         = "MONITOR_SENTRY_ENABLED"
	BuiltInEnvMonitorSentryDSN             = "MONITOR_SENTRY_DSN"
	BuiltInEnvMonitorSentryDebug           = "MONITOR_SENTRY_DEBUG"
	BuiltInEnvMonitorSentryRepanic         = "MONITOR_SENTRY_REPANIC"
	BuiltInEnvMonitorSentryWaitForDelivery = "MONITOR_SENTRY_WAIT_FOR_DELIVERY"
	BuiltInEnvMonitorSentryTimeout         = "MONITOR_SENTRY_TIMEOUT"
)
