package http

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type HTTPServer struct {
	server *http.Server
	router Router
}

func NewHTTPServer(router Router, addr string, readTimeout time.Duration, writeTimeout time.Duration) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:         addr,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			Handler:      router.Handler(),
		},
		router: router,
	}
}

func (s *HTTPServer) Route(path, method string, h http.HandlerFunc) {
	s.router.Route(path, method, h)
}

func (s *HTTPServer) Run() error {
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
