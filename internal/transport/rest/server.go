package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/tuanta7/hydros/config"
	v1admin "github.com/tuanta7/hydros/internal/transport/rest/admin/v1"
	v1public "github.com/tuanta7/hydros/internal/transport/rest/public/v1"
)

type Server struct {
	cfg           *config.Config
	router        *gin.Engine
	server        *http.Server
	cookieStore   *sessions.CookieStore
	clientHandler *v1admin.ClientHandler
	oauthHandler  *v1public.OAuthHandler
	flowHandler   *v1public.FlowHandler
}

func NewServer(cfg *config.Config,
	cookieStore *sessions.CookieStore,
	clientHandler *v1admin.ClientHandler,
	oauthHandler *v1public.OAuthHandler,
	flowHandler *v1public.FlowHandler,
) *Server {
	if !cfg.IsDebugging() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.Static("/static", "./static")
	engine.LoadHTMLGlob("./static/html/*")

	return &Server{
		cfg:    cfg,
		router: engine,
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.RestServerHost, cfg.RestServerPort),
			Handler: nil,
		},
		cookieStore:   cookieStore,
		clientHandler: clientHandler,
		oauthHandler:  oauthHandler,
		flowHandler:   flowHandler,
	}
}

func (s *Server) Run() error {
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
	s.router.GET("/self-service/login", s.flowHandler.LoginPage)
	s.router.POST("/self-service/login", nil)
	s.router.PUT("/self-service/login", s.flowHandler.UpdateAuthenticationStatus)
	s.router.GET("/self-service/consent", s.flowHandler.ConsentPage)
	s.router.POST("/self-service/consent", nil)
	s.router.PUT("/self-service/consent", s.flowHandler.UpdateConsentStatus)

	s.server.Handler = s.router
}
