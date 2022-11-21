package http

import (
	"net/http"
)

type Router interface {
	Route(path, method string, h http.HandlerFunc)
	Handler() http.Handler
}

type muxRouter struct {
	Routes map[string]http.HandlerFunc
}

func NewMuxRouter() Router {
	return &muxRouter{
		Routes: make(map[string]http.HandlerFunc),
	}
}

func (r *muxRouter) Route(path, method string, h http.HandlerFunc) {
	// ignore path (default mux)
	r.Routes[path] = validateMethod(h, method)
}

func (r *muxRouter) Handler() http.Handler {
	mux := http.NewServeMux()
	for route, fn := range r.Routes {
		mux.HandleFunc(route, fn)
	}

	return mux
}

func validateMethod(h http.HandlerFunc, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		h.ServeHTTP(w, r)
	}
}
