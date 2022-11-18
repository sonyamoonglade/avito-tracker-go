package http

import (
	"net/http"
	"time"
)

type Server struct {
	h http.Handler
}

func NewServer(h http.Handler) *Server {
	return &Server{
		h: h,
	}
}

func (s *Server) Run(addr string, readTimeout time.Duration, writeTimeout time.Duration) error {
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Handler:      s.h,
	}

	return server.ListenAndServe()
}
