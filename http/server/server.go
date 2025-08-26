package server

import (
	"fmt"
	"net/http"
	"time"
)

type Router interface {
	Routes() *http.ServeMux
}

type Config struct {
	Port    int           `envconfig:"PORT" default:"8080"`
	Timeout time.Duration `envconfig:"TIMEOUT" default:"30s"`
}

type Server struct {
	*http.Server
	cfg Config
}

func New(cfg Config, router Router) *Server {
	server := &Server{
		Server: &http.Server{
			Handler:     router.Routes(),
			Addr:        fmt.Sprintf(":%d", cfg.Port),
			ReadTimeout: cfg.Timeout,
		},
	}

	return server
}
