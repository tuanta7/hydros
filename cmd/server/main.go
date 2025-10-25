package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tuanta7/hydros/config"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "Hydros",
		Usage: "OIDC and OAuth2.1 Provider",
		Action: func(ctx context.Context, command *cli.Command) error {
			cfg := config.LoadConfig(".env")
			fmt.Println("Loaded config:", cfg.Lifetime.AuthorizeCode)

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
