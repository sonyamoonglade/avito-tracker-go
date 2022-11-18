package routing

import (
	"net/http"
	internalHttp "parser/internal/http"
)

type Router struct {
	Routes map[string]http.HandlerFunc
}

func New() *Router {
	return &Router{
		Routes: make(map[string]http.HandlerFunc),
	}
}

func (r *Router) Add(path string, h http.HandlerFunc) {
	r.Routes[path] = h
}

func (r *Router) Handler() http.Handler {
	mux := http.NewServeMux()
	for route, fn := range r.Routes {
		mux.HandleFunc(route, fn)
	}

	return mux
}

func (r *Router) InitializeHttpEndpoints(handlers *internalHttp.Handlers) {
	r.Add("/subscribe", handlers.NewSubscription)
}
