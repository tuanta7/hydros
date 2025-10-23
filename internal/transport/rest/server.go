package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/config"
)

type Server struct {
	router *gin.Engine
	server *http.Server
}

func NewServer(cfg config.Config) *Server {
	return &Server{
		router: gin.New(),
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.RestServerHost, cfg.RestServerPort),
			Handler: nil,
		},
	}
}

func (s *Server) Run() error {
	errCh := make(chan error)
	go func() {
		err := s.server.ListenAndServe()
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) RegisterRoutes() {
	s.router.GET("/oauth/authorize", nil)
	s.router.POST("/oauth/token", nil)
	s.router.GET("/oauth/introspect", nil)

	s.server.Handler = s.router
}
