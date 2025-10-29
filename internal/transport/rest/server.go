package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/config"
	v1admin "github.com/tuanta7/hydros/internal/transport/rest/admin/v1"
	v1public "github.com/tuanta7/hydros/internal/transport/rest/public/v1"
)

type Server struct {
	cfg           *config.Config
	router        *gin.Engine
	server        *http.Server
	clientHandler *v1admin.ClientHandler
	oauthHandler  *v1public.OAuthHandler
}

func NewServer(cfg *config.Config, clientHandler *v1admin.ClientHandler, oauthHandler *v1public.OAuthHandler) *Server {
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
	// Authorization Service - OAuth APIs
	s.router.GET("/oauth/authorize", s.oauthHandler.HandleAuthorizeRequest)
	s.router.POST("/oauth/token", s.oauthHandler.HandleTokenRequest)
	s.router.POST("/oauth/introspect", s.oauthHandler.HandleIntrospectionRequest)
	s.router.POST("/oauth/revoke", nil)
	s.router.GET("/oauth/logout", nil)
	s.router.POST("/oauth/logout", nil)

	// Authorization Service - Admin APIs
	adminRouter := s.router.Group("/admin/api/v1")
	adminRouter.GET("/clients", s.clientHandler.List)
	adminRouter.POST("/clients", s.clientHandler.Create)

	// Identity Service
	s.router.GET("/self-service/login", nil)  // ui
	s.router.POST("/self-service/login", nil) // submit
	s.router.GET("/self-service/consent", nil)
	s.router.POST("/self-service/consent", nil)

	s.server.Handler = s.router
}
