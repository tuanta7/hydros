package main

import (
	"context"
	"encoding/json"
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
	restpublicv1 "github.com/tuanta7/hydros/internal/transport/rest/public/v1"
	"github.com/tuanta7/hydros/internal/usecase/client"
	"github.com/tuanta7/hydros/pkg/adapter/postgres"
	"github.com/tuanta7/hydros/pkg/adapter/redis"
	"github.com/tuanta7/hydros/pkg/zapx"

	"github.com/tuanta7/hydros/internal/transport"
	"github.com/tuanta7/hydros/internal/transport/rest"
	"github.com/urfave/cli/v3"
)

func main() {
	cfg := config.LoadConfig(".env")
	if b, err := json.MarshalIndent(cfg, "", "\t"); err == nil {
		fmt.Printf("Config: %s", string(b))
		fmt.Println()
	}

	logger, err := zapx.NewLogger(cfg.LogLevel)
	panicErr(err)

	c := &cli.Command{
		Name:  "hydros",
		Usage: "OIDC and OAuth2.1 Provider",
		Commands: []*cli.Command{
			cmd.NewCreateClientsCommand(),
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			pgClient, err := postgres.NewClient(cfg.Postgres.DSN())
			panicErr(err)
			defer pgClient.Close()

			redisClient, err := redis.NewClient(
				context.Background(),
				fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
				redis.WithCredential(cfg.Redis.Username, cfg.Redis.Password),
				redis.WithDB(cfg.Redis.DB),
			)
			panicErr(err)
			defer redisClient.Close()

			clientRepo := pgsource.NewClientRepository(pgClient)
			clientUC := client.NewUseCase(cfg, clientRepo, logger)

			clients, err := clientRepo.List(context.Background(), 1, 10)
			fmt.Println(clients)
			panicErr(err)

			sessionStorage := redissource.NewTokenSessionStorage(cfg, redisClient)
			hmacStrategy, err := hmac.NewHMAC([]byte(cfg.GlobalSecret), cfg.KeyEntropy)
			panicErr(err)

			oauthCore := core.NewOAuth2(cfg, clientUC,
				[]core.AuthorizeHandler{},
				[]core.TokenHandler{
					oauth.NewClientCredentialsGrantHandler(cfg, hmacStrategy, sessionStorage),
				})
			oauthHandler := restpublicv1.NewOAuthHandler(oauthCore)

			restServer := rest.NewServer(cfg, oauthHandler)
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
