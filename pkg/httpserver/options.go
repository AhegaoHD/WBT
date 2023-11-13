package httpserver

import (
	"net"
	"strings"
	"time"
)

type Option func(*Server)

func Port(port string) Option {
	return func(s *Server) {
		if !strings.EqualFold("", port) {
			s.server.Addr = net.JoinHostPort("", port)
		}
	}
}

func ReadTimeout(timeout *time.Duration) Option {
	return func(s *Server) {
		if timeout != nil {
			s.server.ReadTimeout = *timeout * time.Second
		}
	}
}

func WriteTimeout(timeout *time.Duration) Option {
	return func(s *Server) {
		if timeout != nil {
			s.server.WriteTimeout = *timeout * time.Second
		}
	}
}

func ShutdownTimeout(timeout *time.Duration) Option {
	return func(s *Server) {
		if timeout != nil {
			s.shutdownTimeout = *timeout * time.Second
		}
	}
}
