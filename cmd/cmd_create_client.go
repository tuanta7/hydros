package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/internal/usecase/client"
	"github.com/urfave/cli/v3"
)

func NewCreateClientsCommand(clientUC *client.UseCase) *cli.Command {
	cmd := &cli.Command{
		Name:  "create-client",
		Usage: "create a new OAuth client",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Usage:    "client name",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "path to JSON file with client data",
				Required: false,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			name := command.String("name")
			if name == "" {
				name = gofakeit.Name()
			}

			c := &domain.Client{
				Name:                    name,
				Description:             gofakeit.Comment(),
				Scope:                   "example:read",
				GrantTypes:              []string{string(core.GrantTypeClientCredentials)},
				Audience:                []string{"example.com"},
				TokenEndpointAuthMethod: core.ClientAuthenticationMethodBasic,
			}

			err := clientUC.CreateClient(context.Background(), c)
			if err != nil {
				return cli.Exit(err, 1)
			}

			jsonClient, _ := json.MarshalIndent(c, "", "\t")
			fmt.Println("New Client:", string(jsonClient))
			return nil
		},
	}

	return cmd
}
