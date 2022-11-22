package http

import (
	"context"
	"errors"
	"net/http"
	"parser/internal/domain/services"
	"time"
)

// Used to create HTTPServer instance
type ServerConfig struct {
	Router   Router
	Addr     string
	Services *services.Services

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type HTTPServer struct {
	server *http.Server
	router Router

	services *services.Services
}

func NewHTTPServer(cfg *ServerConfig) *HTTPServer {
	srv := &HTTPServer{
		server: &http.Server{
			Addr:         ":8000",
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			Handler:      cfg.Router.Handler(),
		},
		router:   cfg.Router,
		services: cfg.Services,
	}

	defer srv.routes()

	return srv
}

func (s *HTTPServer) routes() {

	// TODO: add .post() .get() as shortcuts
	rt := s.router.Route

	rt("/subscribe", http.MethodPost, s.Subscribe)
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
