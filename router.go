package zoox

import (
	"fmt"
	"strings"

	"github.com/go-zoox/core-utils/safe"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox/components/context/param"
	route "github.com/go-zoox/zoox/components/router"
)

type router struct {
	roots    *safe.Map
	handlers *safe.Map
}

func newRouter() *router {
	return &router{
		roots:    safe.NewMap(),
		handlers: safe.NewMap(),
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
	logger.Info("[router] register: %8s %s", method, path)

	parts := parsePath(path)

	key := fmt.Sprintf("%s %s", method, path)
	if ok := r.roots.Has(method); !ok {
		r.roots.Set(method, &route.Node{})
	}

	r.roots.Get(method).(*route.Node).Insert(path, parts, 0)
	r.handlers.Set(key, handler)
}

func (r *router) getRoute(method string, path string) (*route.Node, map[string]string) {
	searchParts := parsePath(path)
	params := make(map[string]string)
	if ok := r.roots.Has(method); !ok {
		return nil, nil
	}

	root := r.roots.Get(method).(*route.Node)
	if n := root.Search(searchParts, 0); n != nil {
		if n.IsWild() {
			parts := parsePath(n.Path)
			for i, part := range parts {
				if part[0] == ':' {
					// pattern: /user/:name
					params[part[1:]] = searchParts[i]
				} else if part[0] == '{' && part[len(part)-1] == '}' {
					// pattern: /user/{name}
					params[part[1:len(part)-1]] = searchParts[i]
				} else if part[0] == '*' && len(part) > 1 {
					// pattern: /file/*
					params[part[1:]] = strings.Join(searchParts[i:], "/")
					break
				}
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
		if ok := r.handlers.Has(key); ok {
			handler, ok := r.handlers.Get(key).([]HandlerFunc)
			if ok {
				ctx.handlers = append(ctx.handlers, handler...)
			} else {
				ctx.handlers = append(ctx.handlers, ctx.App.notfound)
			}
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
