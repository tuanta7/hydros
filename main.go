package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/sessions"
	"github.com/tuanta7/hydros/cmd"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oauth"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/core/handler/pkce"
	"github.com/tuanta7/hydros/core/signer/hmac"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/internal/jwk"
	"github.com/tuanta7/hydros/internal/session"
	"github.com/tuanta7/hydros/internal/token"
	restadminv1 "github.com/tuanta7/hydros/internal/transport/rest/admin/v1"
	restpublicv1 "github.com/tuanta7/hydros/internal/transport/rest/public/v1"

	"github.com/tuanta7/hydros/pkg/adapter/postgres"
	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/zapx"

	"github.com/tuanta7/hydros/internal/transport"
	"github.com/tuanta7/hydros/internal/transport/rest"
	"github.com/urfave/cli/v3"
)

func main() {
	cfg := config.LoadConfig(".env")
	logger, err := zapx.NewLogger(cfg.LogLevel)
	panicErr(err)
	defer logger.Sync()

	aeadAES, err := aead.NewAESGCM([]byte(cfg.Obfuscation.AESSecretKey))
	panicErr(err)

	pgClient, err := postgres.NewClient(cfg.Postgres.DSN())
	panicErr(err)
	defer pgClient.Close()

	jwkRepo := jwk.NewKeyRepository(pgClient)
	jwkUC := jwk.NewUseCase(cfg, aeadAES, jwkRepo, logger)

	clientRepo := client.NewClientRepository(pgClient)
	clientUC := client.NewUseCase(cfg, clientRepo, logger)

	tokenRepo := token.NewRequestSessionRepo(pgClient)
	tokenStorageUC := token.NewRequestSessionStorage(cfg, aeadAES, tokenRepo)

	flowRepo := flow.NewFlowRepository(pgClient)
	flowUC := flow.NewUseCase(flowRepo, logger)

	loginSessionRepo := session.NewSessionRepository(pgClient)
	loginSessionUC := session.NewUseCase(loginSessionRepo)

	tokenStrategy, err := getTokenStrategy(context.Background(), cfg, jwkUC)
	panicErr(err)

	c := &cli.Command{
		Name:  "hydros",
		Usage: "OIDC and OAuth2.1 Provider",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "with-identity",
				Aliases:  []string{"idp"},
				Usage:    "enable identity provider endpoints",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			cmd.NewCreateClientsCommand(clientUC),
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			oauthAuthorizationCodeGrantHandler := oauth.NewAuthorizationCodeGrantHandler(cfg, tokenStrategy, tokenStorageUC)
			oidcAuthorizationCodeFlowHandler := oidc.NewOpenIDConnectAuthorizationCodeFlowHandler()
			pkceHandler := pkce.NewProofKeyForCodeExchangeHandler(cfg, tokenStrategy, tokenStorageUC)

			oauthCore := core.NewOAuth2(cfg, clientUC,
				[]core.AuthorizeHandler{
					oauthAuthorizationCodeGrantHandler,
					oidcAuthorizationCodeFlowHandler,
					pkceHandler,
				},
				[]core.TokenHandler{
					oauthAuthorizationCodeGrantHandler,
					oidcAuthorizationCodeFlowHandler,
					pkceHandler,
					oauth.NewClientCredentialsGrantHandler(cfg, tokenStrategy, tokenStorageUC),
				},
				[]core.IntrospectionHandler{
					oauth.NewJWTIntrospectionHandler(tokenStrategy),
					oauth.NewTokenIntrospectionHandler(cfg, tokenStrategy, tokenStorageUC),
				},
			)

			cookieStore := newCookieStore(cfg)
			clientHandler := restadminv1.NewClientHandler(clientUC)
			flowHandler := restpublicv1.NewFlowHandler(flowUC)
			oauthHandler := restpublicv1.NewOAuthHandler(cfg, aeadAES, cookieStore, oauthCore, jwkUC, loginSessionUC, flowUC, logger)

			restServer := rest.NewServer(cfg, cookieStore, clientHandler, oauthHandler, flowHandler)
			errCh := make(chan error)
			go func() {
				if err := restServer.Run(); err != nil {
					err = fmt.Errorf("error starting REST server: %w", err)
					errCh <- err
				}
			}()

			notifyCh := make(chan os.Signal, 1)
			signal.Notify(notifyCh, syscall.SIGINT, syscall.SIGTERM)

			select {
			case err := <-errCh:
				log.Println("Shutting down due to server error:", err)
				return shutdownServer(restServer)
			case <-notifyCh:
				log.Println("Shutting down gracefully...")
				_ = shutdownServer(restServer)
				return nil
			}
		},
	}

	if err := c.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func newCookieStore(cfg *config.Config) *sessions.CookieStore {
	// TODO: use better key strategy
	cookieStore := sessions.NewCookieStore(cfg.GetGlobalSecret())
	cookieStore.MaxAge(0)
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.Secure = cfg.Cookie.Secure

	if d := cfg.Cookie.Domain; d != "" {
		cookieStore.Options.Domain = d
	}

	if p := cfg.Cookie.Path; p != "" {
		cookieStore.Options.Path = p
	}
	return cookieStore
}

func shutdownServer(servers ...transport.Server) (err error) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, server := range servers {
		if se := server.Shutdown(shutdownCtx); se != nil {
			log.Println("Error during server shutdown:", se)
			err = errors.Join(err, se)
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func getTokenStrategy(ctx context.Context, cfg *config.Config, jwkUC *jwk.UseCase) (strategy.TokenStrategy, error) {
	hmacSigner, err := hmac.NewSigner(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.GetAccessTokenFormat() == "jwt" {
		getPrivateKeyFn := jwkUC.GetOrCreateJWKFn(jwk.AccessTokenSet)
		jwtSigner, err := jwt.NewSigner(cfg, getPrivateKeyFn)
		if err != nil {
			return nil, err
		}

		return strategy.NewJWTStrategy(hmacSigner, jwtSigner), nil
	}

	return strategy.NewHMACStrategy(cfg, hmacSigner), nil
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
