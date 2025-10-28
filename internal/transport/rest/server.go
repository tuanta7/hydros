package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/internal/transport/rest/public/v1"
)

type Server struct {
	cfg          *config.Config
	router       *gin.Engine
	server       *http.Server
	oauthHandler *v1.OAuthHandler
}

func NewServer(cfg *config.Config, oauthHandler *v1.OAuthHandler) *Server {
	return &Server{
		cfg:    cfg,
		router: gin.New(),
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.RestServerHost, cfg.RestServerPort),
			Handler: nil,
		},
		oauthHandler: oauthHandler,
	}
}

func (s *Server) Run() error {
	gin.SetMode(s.cfg.ReleaseMode)
	return s.server.ListenAndServe()
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
	s.router.POST("/oauth/token", s.oauthHandler.Token)
	s.router.GET("/oauth/introspect", nil)

	s.router.POST("/oauth/revoke", nil)
	s.router.GET("/oauth/logout", nil)
	s.router.POST("/oauth/logout", nil)

	s.router.GET("/self-service/login", nil)

	s.server.Handler = s.router
}
