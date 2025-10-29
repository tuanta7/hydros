package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/config"
	v1private "github.com/tuanta7/hydros/internal/transport/rest/private/v1"
	v1public "github.com/tuanta7/hydros/internal/transport/rest/public/v1"
)

type Server struct {
	cfg           *config.Config
	router        *gin.Engine
	server        *http.Server
	clientHandler *v1private.ClientHandler
	oauthHandler  *v1public.OAuthHandler
}

func NewServer(cfg *config.Config, clientHandler *v1private.ClientHandler, oauthHandler *v1public.OAuthHandler) *Server {
	return &Server{
		cfg:    cfg,
		router: gin.New(),
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.RestServerHost, cfg.RestServerPort),
			Handler: nil,
		},
		clientHandler: clientHandler,
		oauthHandler:  oauthHandler,
	}
}

func (s *Server) Run() error {
	gin.SetMode(s.cfg.ReleaseMode)
	s.RegisterRoutes()

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
	// Authorization Service
	s.router.GET("/oauth/authorize", s.oauthHandler.HandleAuthorizeRequest)
	s.router.POST("/oauth/token", s.oauthHandler.HandleTokenRequest)
	s.router.GET("/oauth/introspect", s.oauthHandler.HandleIntrospectionRequest)

	s.router.POST("/oauth/revoke", nil)
	s.router.GET("/oauth/logout", nil)
	s.router.POST("/oauth/logout", nil)

	s.router.GET("/clients", s.clientHandler.List)
	s.router.POST("/clients", s.clientHandler.Create)

	// Identity Service
	s.router.GET("/self-service/login", nil)  // ui
	s.router.POST("/self-service/login", nil) // submit
	s.router.GET("/self-service/consent", nil)
	s.router.POST("/self-service/consent", nil)

	s.server.Handler = s.router
}
