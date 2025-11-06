package main

import (
	"context"
	"log"
	"os"

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
	flowUC := flow.NewUseCase(flowRepo, aeadAES, logger)

	loginSessionRepo := session.NewSessionRepository(pgClient)
	loginSessionUC := session.NewUseCase(loginSessionRepo)

	tokenStrategy, err := getTokenStrategy(cfg, jwkUC)
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

			cookieStore := session.NewCookieStore(cfg)
			clientHandler := restadminv1.NewClientHandler(clientUC)
			flowHandler := restadminv1.NewFlowHandler(flowUC)

			formHandler := restpublicv1.NewFormHandler(cfg, flowUC)
			oauthHandler := restpublicv1.NewOAuthHandler(cfg, cookieStore, oauthCore, jwkUC, loginSessionUC, flowUC, logger)

			restServer := rest.NewServer(cfg, clientHandler, flowHandler, oauthHandler, formHandler)
			return transport.RunServers(restServer)
		},
	}

	if err = c.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func getTokenStrategy(cfg *config.Config, jwkUC *jwk.UseCase) (strategy.TokenStrategy, error) {
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
