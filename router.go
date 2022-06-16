package zoox

import (
	"fmt"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
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

func (r *router) addRoute(method string, path string, handler HandlerFunc) {
	// logger.Info("Route add: %4s %s", method, path)

	parts := parsePath(path)

	key := fmt.Sprintf("%s %s", method, path)
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &node{}
	}

	r.roots[method].insert(path, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePath(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	if n := root.search(searchParts, 0); n != nil {
		parts := parsePath(n.path)
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
		ctx.Params = params

		key := fmt.Sprintf("%s %s", ctx.Method, n.path)
		if handler, ok := r.handlers[key]; ok {
			ctx.handlers = append(ctx.handlers, handler)
		} else {
			ctx.handlers = append(ctx.handlers, ctx.app.notfound)
		}
	} else {
		ctx.handlers = append(ctx.handlers, ctx.app.notfound)
	}

	ctx.Next()
}
