package test

import (
	"testing"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oauth"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/core/handler/pkce"
	"github.com/tuanta7/hydros/core/signer/hmac"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/internal/jwk"
	"github.com/tuanta7/hydros/internal/session"
	"github.com/tuanta7/hydros/internal/token"
	restadminv1 "github.com/tuanta7/hydros/internal/transport/rest/admin/v1"
	restpublicv1 "github.com/tuanta7/hydros/internal/transport/rest/public/v1"
	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/logger"
	"github.com/tuanta7/hydros/pkg/postgres"
)

type App struct {
	Config        *config.Config
	ClientUC      *client.UseCase
	FlowUC        *flow.UseCase
	OAuthCore     *core.OAuth2
	ClientHandler *restadminv1.ClientHandler
	FlowHandler   *restadminv1.FlowHandler
	OAuthHandler  *restpublicv1.OAuthHandler
	FormHandler   *restpublicv1.FormHandler
	TokenStorage  *token.RequestSessionStorage
	Cleanup       func()
}

func SetupTestApp(t *testing.T) *App {
	t.Helper()

	cfg := &config.Config{
		Version:        "1.0.0-test",
		LogLevel:       "debug",
		ReleaseMode:    "debug",
		RestServerHost: "localhost",
		RestServerPort: "8080",
		GRPCServerHost: "localhost",
		GRPCServerPort: "9090",
		Obfuscation: config.ObfuscationConfig{
			AESSecretKey: "test-secret-key-32-chars-long!!!",
		},
		HMAC: config.HMACConfig{
			GlobalSecret: "test-global-secret-64-chars-long!!!test-global-secret-64-chars-long!!!",
		},
	}

	zl, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		t.Fatalf("Failed to create zl: %v", err)
	}

	aeadAES, err := aead.NewAESGCM([]byte(cfg.Obfuscation.AESSecretKey))
	if err != nil {
		t.Fatalf("Failed to create AEAD: %v", err)
	}

	pgClient, err := postgres.NewClientFromPool(GetPgxPool())
	if err != nil {
		t.Fatalf("Failed to create postgres client: %v", err)
	}

	// Initialize repositories and use cases
	jwkRepo := jwk.NewKeyRepository(pgClient)
	jwkUC := jwk.NewUseCase(cfg, aeadAES, jwkRepo, zl)

	clientRepo := client.NewClientRepository(pgClient)
	clientUC := client.NewUseCase(cfg, clientRepo, zl)

	tokenRepo := token.NewRequestSessionRepo(pgClient)
	tokenStorage := token.NewRequestSessionStorage(cfg, aeadAES, tokenRepo)

	flowRepo := flow.NewFlowRepository(pgClient)
	flowUC := flow.NewUseCase(cfg, flowRepo, aeadAES, zl)

	loginSessionRepo := session.NewSessionRepository(pgClient)
	loginSessionUC := session.NewUseCase(loginSessionRepo)

	hmacSigner, err := hmac.NewSigner(cfg)
	if err != nil {
		t.Fatalf("Failed to create HMAC signer: %v", err)
	}

	jwtSigner, err := jwt.NewSigner(cfg, jwkUC.GetOrCreateJWKFn(jwk.AccessTokenSet))
	if err != nil {
		t.Fatalf("Failed to create JWT signer: %v", err)
	}

	tokenStrategy := oauth.NewJWTStrategy(cfg, hmacSigner, jwtSigner)

	idTokenSigner, err := jwt.NewSigner(cfg, jwkUC.GetOrCreateJWKFn(jwk.IDTokenSet))
	if err != nil {
		t.Fatalf("Failed to create ID Token signer: %v", err)
	}
	idTokenStrategy := oidc.NewIDTokenStrategy(cfg, idTokenSigner)

	// Setup OAuth handlers
	oauthCore := core.NewOAuth2(cfg, clientUC,
		oauth.NewAuthorizationCodeGrantHandler(cfg, tokenStrategy, tokenStorage),
		oidc.NewOpenIDConnectAuthorizationCodeFlowHandler(cfg, idTokenStrategy, tokenStorage),
		pkce.NewProofKeyForCodeExchangeHandler(cfg, tokenStrategy, tokenStorage),
		oauth.NewClientCredentialsGrantHandler(cfg, tokenStrategy, tokenStorage),
		oauth.NewJWTIntrospectionHandler(tokenStrategy),
		oauth.NewTokenIntrospectionHandler(cfg, tokenStrategy, tokenStorage),
	)

	// Setup handlers
	cookieStore := session.NewCookieStore(cfg)
	clientHandler := restadminv1.NewClientHandler(clientUC)
	flowHandler := restadminv1.NewFlowHandler(flowUC)
	formHandler := restpublicv1.NewFormHandler(cfg, flowUC)
	oauthHandler := restpublicv1.NewOAuthHandler(cfg, cookieStore, oauthCore, jwkUC, loginSessionUC, flowUC, zl)

	cleanup := func() {
		_ = zl.Sync()
	}

	return &App{
		Config:        cfg,
		ClientUC:      clientUC,
		FlowUC:        flowUC,
		OAuthCore:     oauthCore,
		ClientHandler: clientHandler,
		FlowHandler:   flowHandler,
		OAuthHandler:  oauthHandler,
		FormHandler:   formHandler,
		TokenStorage:  tokenStorage,
		Cleanup:       cleanup,
	}
}
