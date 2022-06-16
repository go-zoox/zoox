package zoox

import (
	"net/http"
	"path"
	"time"
)

// RouterGroup is a group of routes.
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *RouterGroup
	app         *Application
}

// Group defines a new router group
func (g *RouterGroup) Group(prefix string) *RouterGroup {
	app := g.app
	newGroup := &RouterGroup{
		prefix: g.prefix + prefix,
		parent: g,
		app:    app,
	}
	g.app.groups = append(g.app.groups, newGroup)
	return newGroup
}

func (g *RouterGroup) addRoute(method string, path string, handler HandlerFunc) {
	pathX := g.prefix + path
	g.app.router.addRoute(method, pathX, handler)
}

// Get defines the method to add GET request
func (g *RouterGroup) Get(path string, handler HandlerFunc) {
	g.addRoute("GET", path, handler)
}

// Post defines the method to add POST request
func (g *RouterGroup) Post(path string, handler HandlerFunc) {
	g.addRoute("POST", path, handler)
}

// Put defines the method to add PUT request
func (g *RouterGroup) Put(path string, handler HandlerFunc) {
	g.addRoute("PUT", path, handler)
}

// Patch defines the method to add PATCH request
func (g *RouterGroup) Patch(path string, handler HandlerFunc) {
	g.addRoute("PATCH", path, handler)
}

// Delete defines the method to add DELETE request
func (g *RouterGroup) Delete(path string, handler HandlerFunc) {
	g.addRoute("DELETE", path, handler)
}

// Head defines the method to add HEAD request
func (g *RouterGroup) Head(path string, handler HandlerFunc) {
	g.addRoute("HEAD", path, handler)
}

// Use adds a middleware to the group
func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(g.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(ctx *Context) {
		file := ctx.Param("filepath")
		// Check if file exists and/or is not a directory
		if _, err := fs.Open(file); err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

// StaticOptions is the options for static method
type StaticOptions struct {
	Gzip         bool
	Md5          bool
	CacheControl string
	MaxAge       time.Duration
	Index        bool
	Suffix       string
}

// Static defines the method to serve static files
func (g *RouterGroup) Static(relativePath string, root string, options ...StaticOptions) {
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	pathX := path.Join(relativePath, "/*filepath")

	//
	g.Get(pathX, handler)
}
