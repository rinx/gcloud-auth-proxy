package router

import (
	"net/http"
)

type Router interface {
	http.Handler
	Register(routes map[RoutePattern]http.Handler)
}

type RoutePattern string

type router struct {
	mux *http.ServeMux
}

func New() Router {
	return &router{
		mux: http.NewServeMux(),
	}
}

func (ro *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ro.mux.ServeHTTP(w, r)
}

func (ro *router) Register(routes map[RoutePattern]http.Handler) {
	for pattern, handler := range routes {
		ro.mux.Handle(string(pattern), handler)
	}
}
