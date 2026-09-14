package http

import (
	"context"
	"net/http"
)

type Registrator interface {
	Register(mux *http.ServeMux)
}

type Server struct {
	server *http.Server
}

func NewServer(port string, registrators ...Registrator) *Server {
	mux := &http.ServeMux{}

	for _, reg := range registrators {
		reg.Register(mux)
	}

	return &Server{
		server: &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
