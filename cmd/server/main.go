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

	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/internal/transport"
	"github.com/tuanta7/hydros/internal/transport/rest"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "Hydros",
		Usage: "OIDC and OAuth2.1 Provider",
		Action: func(ctx context.Context, command *cli.Command) error {
			cfg := config.LoadConfig(".env")
			fmt.Println("Loaded config:", cfg.Lifetime.AuthorizationCode)

			restServer := rest.NewServer(cfg)

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

	if err := cmd.Run(context.Background(), os.Args); err != nil {
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
