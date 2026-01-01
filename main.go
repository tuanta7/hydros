package main

import (
	"context"
	"log"
	"os"

	"github.com/tuanta7/hydros/cmd"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oauth"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/core/handler/pkce"
	"github.com/tuanta7/hydros/core/signer/hmac"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/internal/jwk"
	"github.com/tuanta7/hydros/internal/login"
	"github.com/tuanta7/hydros/internal/session"
	"github.com/tuanta7/hydros/internal/token"
	restadminv1 "github.com/tuanta7/hydros/internal/transport/rest/admin/v1"
	restpublicv1 "github.com/tuanta7/hydros/internal/transport/rest/public/v1"
	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/postgres"
	"github.com/tuanta7/hydros/pkg/zapx"

	"github.com/tuanta7/hydros/internal/transport"
	"github.com/tuanta7/hydros/internal/transport/rest"
	"github.com/urfave/cli/v3"
)

func main() {
	cfg := config.LoadConfig(".env")
	zl, err := zapx.NewLogger(cfg.LogLevel)
	panicErr(err)
	defer zl.Sync()

	aeadAES, err := aead.NewAESGCM([]byte(cfg.Obfuscation.AESSecretKey))
	panicErr(err)

	pgClient, err := postgres.NewClient(cfg.Postgres.DSN())
	panicErr(err)
	defer pgClient.Close()

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

	tokenStrategy, err := getTokenStrategy(cfg, jwkUC)
	panicErr(err)

	idTokenSigner, err := jwt.NewSigner(cfg, jwkUC.GetOrCreateJWKFn(jwk.IDTokenSet))
	panicErr(err)

	idTokenStrategy := oidc.NewIDTokenStrategy(cfg, idTokenSigner)

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
			cmd.NewCleanCommand(),
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			oauthCore := core.NewOAuth2(cfg, clientUC,
				oauth.NewAuthorizationCodeGrantHandler(cfg, tokenStrategy, tokenStorage),
				oidc.NewOpenIDConnectAuthorizationCodeFlowHandler(cfg, idTokenStrategy, tokenStorage),
				pkce.NewProofKeyForCodeExchangeHandler(cfg, tokenStrategy, tokenStorage),
				oauth.NewClientCredentialsGrantHandler(cfg, tokenStrategy, tokenStorage),
				oauth.NewJWTIntrospectionHandler(tokenStrategy),
				oauth.NewTokenIntrospectionHandler(cfg, tokenStrategy, tokenStorage),
			)

			cookieStore := session.NewCookieStore(cfg)
			clientHandler := restadminv1.NewClientHandler(clientUC)
			flowHandler := restadminv1.NewFlowHandler(flowUC)

			defaultLoginStrategy := login.NewDefaultStrategy()
			formHandler := restpublicv1.NewFormHandler(cfg, flowUC, defaultLoginStrategy)
			oauthHandler := restpublicv1.NewOAuthHandler(cfg, cookieStore, oauthCore, jwkUC, loginSessionUC, flowUC, zl)

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

	if cfg.GetAccessTokenFormat() == config.AccessTokenFormatJWT {
		getPrivateKeyFn := jwkUC.GetOrCreateJWKFn(jwk.AccessTokenSet)
		jwtSigner, err := jwt.NewSigner(cfg, getPrivateKeyFn)
		if err != nil {
			return nil, err
		}

		return oauth.NewJWTStrategy(cfg, hmacSigner, jwtSigner), nil
	}

	return oauth.NewHMACStrategy(cfg, hmacSigner), nil
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
