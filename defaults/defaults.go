package defaults

import (
	"os"

	"github.com/go-zoox/debug"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

// Defaults returns a new default zoox.
func Defaults() *zoox.Application {
	// @TODO use env for create sentry in top level
	if os.Getenv(zoox.BuiltInEnvMonitorSentryEnabled) == "true" && os.Getenv(zoox.BuiltInEnvMonitorSentryDSN) != "" {
		middleware.InitSentry(middleware.InitSentryOption{
			Dsn:   os.Getenv(zoox.BuiltInEnvMonitorSentryDSN),
			Debug: debug.IsDebugMode(),
		})
	}

	zoox.DefaultMiddleware("recovery", func(app *zoox.Application) {
		app.Use(middleware.Recovery())
	})

	zoox.DefaultMiddleware("request_id", func(app *zoox.Application) {
		app.Use(middleware.RequestID())
	})

	zoox.DefaultMiddleware("realip", func(app *zoox.Application) {
		app.Use(middleware.RealIP())
	})

	zoox.DefaultMiddleware("logger", func(app *zoox.Application) {
		app.Use(middleware.Logger())
	})

	zoox.DefaultMiddleware("healthcheck", func(app *zoox.Application) {
		app.Use(middleware.HealthCheck())
	})

	zoox.DefaultMiddleware("runtime", func(app *zoox.Application) {
		app.Use(middleware.Runtime())
	})

	if debug.IsDebugMode() {
		zoox.DefaultMiddleware("pprof", func(app *zoox.Application) {
			app.Use(middleware.PProf())
		})
	}

	// zoox.DefaultMiddleware("cors", func(app *zoox.Application) {
	// 	app.Use(middleware.CORS())
	// })

	app := zoox.New()

	app.SetBeforeReady(func() {
		if app.Config.Monitor.Prometheus.Enabled {
			app.Logger.Infof("[middleware] register: prometheus (by config) ...")

			app.Use(middleware.Prometheus(func(opt *middleware.PrometheusOption) {
				opt.Path = app.Config.Monitor.Prometheus.Path
			}))
		}

		if app.Config.Monitor.Sentry.Enabled {
			if app.Config.Monitor.Sentry.DSN == "" {
				panic("app.Config.Monitor.Sentry.DSN is required")
			}

			app.Logger.Infof("[middleware] register: sentry (by config) ...")

			// @TODO
			if os.Getenv(zoox.BuiltInEnvMonitorSentryEnabled) != "true" {
				middleware.InitSentry(middleware.InitSentryOption{
					Dsn:   app.Config.Monitor.Sentry.DSN,
					Debug: app.Config.Monitor.Sentry.Debug,
				})
			}

			app.Use(middleware.Sentry(func(opt *middleware.SentryOption) {
				opt.Repanic = true
				opt.WaitForDelivery = app.Config.Monitor.Sentry.WaitForDelivery
				opt.Timeout = app.Config.Monitor.Sentry.Timeout
			}))
		}
	})

	app.SetBeforeDestroy(func() {
		middleware.FinishSentry()
	})

	return app
}
