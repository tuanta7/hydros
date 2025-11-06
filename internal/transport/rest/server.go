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
	flowHandler   *v1admin.FlowHandler
	oauthHandler  *v1public.OAuthHandler
	formHandler   *v1public.FormHandler
}

func NewServer(cfg *config.Config,
	clientHandler *v1admin.ClientHandler,
	flowHandler *v1admin.FlowHandler,
	oauthHandler *v1public.OAuthHandler,
	formHandler *v1public.FormHandler,
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

		clientHandler: clientHandler,
		oauthHandler:  oauthHandler,
		formHandler:   formHandler,
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

	// Default forms and submit endpoints
	s.router.GET("/self-service/login", s.formHandler.LoginPage)
	s.router.POST("/self-service/login", s.formHandler.Login)
	s.router.GET("/self-service/consent", s.formHandler.ConsentPage)
	s.router.POST("/self-service/consent", nil)

	// Authorization Service - Admin APIs
	adminRouter := s.router.Group("/admin/api/v1")
	adminRouter.GET("/clients", s.clientHandler.List)
	adminRouter.POST("/clients", s.clientHandler.Create)

	// Identity Service
	adminRouter.GET("/login/flows", s.flowHandler.GetLoginFlow)
	adminRouter.PUT("/login/accept", s.flowHandler.AcceptLogin)
	adminRouter.PUT("/login/reject", s.flowHandler.RejectLogin)
	adminRouter.GET("/consent/flows", nil)
	adminRouter.PUT("/consent/accept", nil)
	adminRouter.PUT("/consent/reject", nil)

	s.server.Handler = s.router
}
