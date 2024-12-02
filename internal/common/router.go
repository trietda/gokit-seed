// This file is based on https://github.com/julienschmidt/httprouter/pull/89/files#diff-e105a475d2665ddcd60afe2fb46441e174123b1fb77044946040e7a18b9346a0

package common

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Router struct {
	*httprouter.Router
}

func NewRouter(r *httprouter.Router) *Router {
	if r != nil {
		return &Router{r}
	}

	return &Router{httprouter.New()}
}

// NewGroup adds a zero overhead group of routes that share a common root path.
func (r *Router) NewGroup(path string) *RouteGroup {
	return newRouteGroup(r, path)
}

type RouteGroup struct {
	r    *Router
	Path string
}

func newRouteGroup(r *Router, path string) *RouteGroup {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	//Strip traling / (if present) as all added sub paths must start with a /
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	return &RouteGroup{r: r, Path: path}
}

func NewRouteGroup(path string) *RouteGroup {
	return newRouteGroup(NewRouter(nil), path)
}

func (r *RouteGroup) NewGroup(path string) *RouteGroup {
	return newRouteGroup(r.r, r.SubPath(path))
}

func (r *RouteGroup) Handle(method, path string, handle httprouter.Handle) {
	r.r.Handle(method, r.SubPath(path), handle)
}

func (r *RouteGroup) Handler(method, path string, handler http.Handler) {
	r.r.Handler(method, r.SubPath(path), handler)
}

func (r *RouteGroup) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.r.HandlerFunc(method, r.SubPath(path), handler)
}

func (r *RouteGroup) GET(path string, handle httprouter.Handle) {
	r.Handle("GET", path, handle)
}

func (r *RouteGroup) HEAD(path string, handle httprouter.Handle) {
	r.Handle("HEAD", path, handle)
}

func (r *RouteGroup) OPTIONS(path string, handle httprouter.Handle) {
	r.Handle("OPTIONS", path, handle)
}

func (r *RouteGroup) POST(path string, handle httprouter.Handle) {
	r.Handle("POST", path, handle)
}

func (r *RouteGroup) PUT(path string, handle httprouter.Handle) {
	r.Handle("PUT", path, handle)
}

func (r *RouteGroup) PATCH(path string, handle httprouter.Handle) {
	r.Handle("PATCH", path, handle)
}

func (r *RouteGroup) DELETE(path string, handle httprouter.Handle) {
	r.Handle("DELETE", path, handle)
}

func (r *RouteGroup) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.r.ServeHTTP(w, req)
}

func (r *RouteGroup) SubPath(path string) string {
	if path[0] != '/' {
		panic("path must start with a '/'")
	}

	return r.Path + path
}
