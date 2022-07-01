package zoox

import "fmt"

var defaultGroupsFns = map[string]func(r *RouterGroup){}

// Default returns a new default zoox.
func Default() *Application {
	app := New()

	app.Use(HealthCheck())
	app.Use(Recovery())
	app.Use(Logger())

	for gprefix, gfn := range defaultGroupsFns {
		fmt.Println("zoox: default group:", gprefix)
		g := app.Group(gprefix)
		gfn(g)
	}

	return app
}

// DefaultGroup ...
func DefaultGroup(prefix string, fn func(r *RouterGroup)) {
	if _, ok := defaultGroupsFns[prefix]; ok {
		panic(fmt.Errorf("zoox: default group (%s) already registered", prefix))
	}

	defaultGroupsFns[prefix] = fn
}
