package zoox

// Default returns a new default zoox.
func Default() *Application {
	app := New()
	app.Use(Recovery())
	app.Use(Logger())
	return app
}
