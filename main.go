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
	"github.com/tuanta7/hydros/core/token/hmac"
	pgsource "github.com/tuanta7/hydros/internal/datasource/postgres"
	redissource "github.com/tuanta7/hydros/internal/datasource/redis"
	restadminv1 "github.com/tuanta7/hydros/internal/transport/rest/admin/v1"
	restpublicv1 "github.com/tuanta7/hydros/internal/transport/rest/public/v1"
	clientuc "github.com/tuanta7/hydros/internal/usecase/client"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
	"github.com/tuanta7/hydros/pkg/adapter/redis"
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

	pgClient, err := postgres.NewClient(cfg.Postgres.DSN())
	panicErr(err)
	defer pgClient.Close()

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
	hmacStrategy, err := hmac.NewHMAC([]byte(cfg.GlobalSecret), cfg.KeyEntropy)
	panicErr(err)

	oauthCore := core.NewOAuth2(cfg, clientUC,
		[]core.AuthorizeHandler{},
		[]core.TokenHandler{
			oauth.NewClientCredentialsGrantHandler(cfg, hmacStrategy, sessionStorage),
		},
		[]core.IntrospectionHandler{
			oauth.NewTokenIntrospectionHandler(cfg, hmacStrategy, nil, sessionStorage, nil),
		},
	)

	c := &cli.Command{
		Name:  "hydros",
		Usage: "OIDC and OAuth2.1 Provider",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "with-identity",
				Aliases:  []string{"idp"},
				Usage:    "run with identity provider endpoint enabled",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			cmd.NewCreateClientsCommand(clientUC),
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			clientHandler := restadminv1.NewClientHandler(clientUC)
			oauthHandler := restpublicv1.NewOAuthHandler(cfg, oauthCore, logger)

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

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
