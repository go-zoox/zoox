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
