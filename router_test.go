package zoox

import (
	"reflect"
	"testing"
)

func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/usersx/:nid", nil)
	r.addRoute("GET", "/users/{id}", nil)
	r.addRoute("GET", "/users/:id/profile", nil)
	r.addRoute("GET", "/users/:id/logs/:lid", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePath(t *testing.T) {
	if !reflect.DeepEqual(parsePath("/p/:name"), []string{"p", ":name"}) {
		t.Errorf("Expected [p,:name], got %v", parsePath("/p/:name"))
	}

	if !reflect.DeepEqual(parsePath("/p/*"), []string{"p", "*"}) {
		t.Errorf("Expected [p,*], got %v", parsePath("/p/*"))
	}

	if !reflect.DeepEqual(parsePath("/p/*name/*"), []string{"p", "*name"}) {
		t.Errorf("Expected [p,*name], got %v", parsePath("/p/*name/*"))
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/zoox")

	if n == nil {
		t.Fatal("Expected node, got nil")
	}

	if n.Path != "/hello/:name" {
		t.Errorf("Expected /hello/:name, got %s", n.Path)
	}

	if ps["name"] != "zoox" {
		t.Errorf("Expected zoox, got %s", ps["name"])
	}
}

func TestGetRouteMultiParams(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/users/1/logs/6")

	if n == nil {
		t.Fatal("Expected node, got nil")
	}

	if n.Path != "/users/:id/logs/:lid" {
		t.Errorf("Expected /users/:id/logs/:lid, got %s", n.Path)
	}

	if ps["id"] != "1" {
		t.Errorf("Expected 1, got %s", ps["id"])
	}

	if ps["lid"] != "6" {
		t.Errorf("Expected 6, got %s", ps["lid"])
	}
}

func TestGetRouteWithBrackets(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/users/1")

	if n == nil {
		t.Fatal("Expected node, got nil")
	}

	if n.Path != "/users/{id}" {
		t.Errorf("Expected /users/{id}, got %s", n.Path)
	}

	if ps["id"] != "1" {
		t.Errorf("Expected 1, got %v", ps["id"])
	}
}
