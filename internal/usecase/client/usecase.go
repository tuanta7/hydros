package client

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/helper"
	"github.com/tuanta7/hydros/pkg/timex"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type UseCase struct {
	cfg        *config.Config
	clientRepo Repository
	logger     *zapx.Logger
}

func NewUseCase(cfg *config.Config, clientRepo Repository, logger *zapx.Logger) *UseCase {
	return &UseCase{
		cfg:        cfg,
		clientRepo: clientRepo,
		logger:     logger,
	}
}

func (u *UseCase) ListClients(ctx context.Context, page, pageSize uint64) ([]core.Client, error) {
	clients, err := u.clientRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	var result []core.Client
	for _, client := range clients {
		result = append(result, client)
	}

	return result, nil
}

func (u *UseCase) CreateClient(ctx context.Context, client *domain.Client) error {
	secret := ""
	if client.Secret != "" {
		secret = client.Secret
	} else {
		secret = helper.GenerateSecret(26)
	}

	hashedSecret, err := u.cfg.GetSecretsHasher().Hash(ctx, []byte(secret))
	if err != nil {
		return err
	}

	client.Secret = string(hashedSecret)

	if client.ID == "" {
		client.ID = strings.Replace(uuid.NewString(), "-", "", -1)
	}

	if client.TokenEndpointAuthMethod == "" {
		client.TokenEndpointAuthMethod = "none"
	}

	if client.TokenEndpointAuthSigningAlg == "" {
		client.TokenEndpointAuthSigningAlg = "none"
	}

	client.CreatedAt = timex.NowUTC().Round(time.Second)
	client.UpdatedAt = client.CreatedAt

	err = u.clientRepo.Create(ctx, client)
	if err != nil {
		u.logger.Error("error while creating client",
			zap.Error(err),
			zap.String("method", "clientRepo.Create"),
		)
		return err
	}

	if !client.IsPublic() {
		// let the creator know the secret for only once
		client.Secret = secret
	}

	return nil
}

func (u *UseCase) GetClient(ctx context.Context, id string) (core.Client, error) {
	client, err := u.clientRepo.Get(ctx, id)
	if err != nil {
		u.logger.Error("cannot get client",
			zap.Error(err),
			zap.String("method", "clientRepo.Get"),
		)
		return nil, err
	}

	return client, nil
}
