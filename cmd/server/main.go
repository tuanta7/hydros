package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/knadh/koanf/v2"
	"github.com/tuanta7/oauth-server/config"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "OAuth Server",
		Usage: "",
		Action: func(ctx context.Context, command *cli.Command) error {
			cfg := config.LoadEnvConfig()
			cfg.LoadJSONConfig(koanf.New("."))

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
