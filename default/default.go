package zd

import (
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

// Default returns a new default zoox.
func Default() *zoox.Application {
	zoox.DefaultMiddleware("recovery", func(app *zoox.Application) {
		app.Use(middleware.Recovery())
	})

	zoox.DefaultMiddleware("logger", func(app *zoox.Application) {
		app.Use(middleware.Logger())
	})

	zoox.DefaultMiddleware("healthcheck", func(app *zoox.Application) {
		app.Use(middleware.HealthCheck())
	})

	app := zoox.New()

	return app
}
