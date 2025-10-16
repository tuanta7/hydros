package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tuanta7/oauth-server/config"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "OAuth Server",
		Usage: "",
		Action: func(ctx context.Context, command *cli.Command) error {
			cfg := &config.Config{}
			config.LoadJSONConfig(cfg)

			for {
				time.Sleep(10 * time.Second)
				fmt.Printf("Version: %s\n", cfg.Version)
			}
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
