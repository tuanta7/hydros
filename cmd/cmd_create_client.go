package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func NewCreateClientsCommand() *cli.Command {
	cmd := &cli.Command{
		Name:  "create-client",
		Usage: "create a new OAuth client",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Usage:    "client name",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			fmt.Println("create-client called; args:", os.Args[1:])
			return nil
		},
	}

	return cmd
}
