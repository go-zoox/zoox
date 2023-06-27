package defaults

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

// Defaults returns a new default zoox.
func Defaults() *zoox.Application {
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

	// zoox.DefaultMiddleware("pprof", func(app *zoox.Application) {
	// 	app.Use(middleware.PProf())
	// })

	// zoox.DefaultMiddleware("cors", func(app *zoox.Application) {
	// 	app.Use(middleware.CORS())
	// })

	app := zoox.New()

	return app
}
