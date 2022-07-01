package zoox

import "fmt"

// DefaultMiddlewares is the default global middleware
var DefaultMiddlewares = map[string]func(app *Application){
	// Logger,
}

// DefaultMiddleware
func DefaultMiddleware(name string, fn func(app *Application)) {
	DefaultMiddlewares[name] = fn
}

var DefaultGroupsFns = map[string]func(r *RouterGroup){}

// DefaultGroup ...
func DefaultGroup(prefix string, fn func(r *RouterGroup)) {
	if _, ok := DefaultGroupsFns[prefix]; ok {
		panic(fmt.Errorf("zoox: default group (%s) already registered", prefix))
	}

	DefaultGroupsFns[prefix] = fn
}
