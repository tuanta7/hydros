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

	"github.com/tuanta7/hydros/cmd"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oauth"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/core/handler/pkce"
	"github.com/tuanta7/hydros/core/signer/hmac"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/core/strategy"
	pgsource "github.com/tuanta7/hydros/internal/datasource/postgres"
	redissource "github.com/tuanta7/hydros/internal/datasource/redis"
	"github.com/tuanta7/hydros/internal/domain"
	restadminv1 "github.com/tuanta7/hydros/internal/transport/rest/admin/v1"
	restpublicv1 "github.com/tuanta7/hydros/internal/transport/rest/public/v1"
	clientuc "github.com/tuanta7/hydros/internal/usecase/client"
	"github.com/tuanta7/hydros/internal/usecase/jwk"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
	"github.com/tuanta7/hydros/pkg/adapter/redis"
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

	jwkRepo := pgsource.NewKeyRepository(pgClient)
	jwkUC := jwk.NewUseCase(cfg, aeadAES, jwkRepo, logger)

	clientRepo := pgsource.NewClientRepository(pgClient)
	clientUC := clientuc.NewUseCase(cfg, clientRepo, logger)

	redisClient, err := redis.NewClient(
		context.Background(),
		fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		redis.WithCredential(cfg.Redis.Username, cfg.Redis.Password),
		redis.WithDB(cfg.Redis.DB),
	)
	panicErr(err)
	defer redisClient.Close()

	sessionStorage := redissource.NewTokenSessionStorage(cfg, redisClient)
	tokenStrategy, err := getTokenStrategy(context.Background(), cfg, jwkUC)
	panicErr(err)

	oauthAuthorizationCodeGrantHandler := oauth.NewAuthorizationCodeGrantHandler(cfg, tokenStrategy, sessionStorage)
	oidcAuthorizationCodeFlowHandler := oidc.NewOpenIDConnectAuthorizationCodeFlowHandler()
	pkceHandler := pkce.NewProofKeyForCodeExchangeHandler(cfg)

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
			oauth.NewClientCredentialsGrantHandler(cfg, tokenStrategy, sessionStorage),
		},
		[]core.IntrospectionHandler{
			oauth.NewJWTIntrospectionHandler(tokenStrategy),
			oauth.NewTokenIntrospectionHandler(cfg, tokenStrategy, sessionStorage),
		},
	)

	c := &cli.Command{
		Name:  "hydros",
		Usage: "OIDC and OAuth2.1 Provider",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "with-identity",
				Aliases:  []string{"idp"},
				Usage:    "enable identity provider endpoints ",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			cmd.NewCreateClientsCommand(clientUC),
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			clientHandler := restadminv1.NewClientHandler(clientUC)
			oauthHandler := restpublicv1.NewOAuthHandler(cfg, oauthCore, jwkUC, logger)

			restServer := rest.NewServer(cfg, clientHandler, oauthHandler)
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
		getPrivateKeyFn := jwkUC.GetOrCreateJWKFn(domain.AccessTokenSet)
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
