package http

import (
	"net/http"
)

type Router interface {
	Route(path, method string, h http.HandlerFunc)
	Handler() http.Handler
}

type muxRouter struct {
	m      *http.ServeMux
	prefix string
}

func NewMuxRouter() Router {
	return &muxRouter{
		m:      http.NewServeMux(),
		prefix: "",
	}
}

func (r *muxRouter) SetGlobalPrefix(prefix string) {
	r.prefix = prefix
}

func (r *muxRouter) Route(path, method string, h http.HandlerFunc) {
<<<<<<< HEAD
	r.m.HandleFunc(r.prefix+path, validateMethod(h, method))
=======
	r.Routes[path] = validateMethod(h, method)
>>>>>>> b015edd8fca5456079496f321d715e627636dbd0
}

func (r *muxRouter) Handler() http.Handler {
	return r.m
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
