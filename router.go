package zoox

import (
	"fmt"
	"strings"

	"github.com/go-zoox/zoox/components/context/param"
	route "github.com/go-zoox/zoox/components/router"
)

type router struct {
	roots    map[string]*route.Node
	handlers map[string][]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*route.Node),
		handlers: make(map[string][]HandlerFunc),
	}
}

func parsePath(path string) []string {
	partsX := strings.Split(path, "/")

	parts := []string{}
	for _, part := range partsX {
		if part != "" {
			parts = append(parts, part)
			// * is a wildcard
			if part[0] == '*' {
				break
			}
		}
	}

	return parts
}

func (r *router) addRoute(method string, path string, handler ...HandlerFunc) {
	// logger.Info("Route add: %4s %s", method, path)

	parts := parsePath(path)

	key := fmt.Sprintf("%s %s", method, path)
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &route.Node{}
	}

	r.roots[method].Insert(path, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*route.Node, map[string]string) {
	searchParts := parsePath(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	if n := root.Search(searchParts, 0); n != nil {
		parts := parsePath(n.Path)
		for i, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[i]
			} else if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[i:], "/")
				break
			}
		}

		return n, params
	}

	return nil, nil
}

func (r *router) handle(ctx *Context) {
	n, params := r.getRoute(ctx.Method, ctx.Path)
	if n != nil {
		ctx.param = param.New(params)

		key := fmt.Sprintf("%s %s", ctx.Method, n.Path)
		if handler, ok := r.handlers[key]; ok {
			ctx.handlers = append(ctx.handlers, handler...)
		} else {
			ctx.handlers = append(ctx.handlers, ctx.App.notfound)
		}
	} else {
		ctx.handlers = append(ctx.handlers, ctx.App.notfound)
	}

	ctx.Next()

	if !ctx.Writer.Written() {
		ctx.Writer.Flush()
	}
}
