package client

import (
	"context"
	"errors"
	"time"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/pkg/dbtype"
	"github.com/tuanta7/hydros/pkg/helper/stringx"

	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type UseCase struct {
	cfg        *config.Config
	clientRepo Repository
	logger     *zapx.ZapLogger
}

func NewUseCase(cfg *config.Config, clientRepo Repository, logger *zapx.ZapLogger) *UseCase {
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

func (u *UseCase) CreateClient(ctx context.Context, client *Client) error {
	if client == nil {
		return errors.New("client cannot be nil")
	}

	secret := ""
	if client.Secret != "" {
		secret = client.Secret
	} else {
		secret = stringx.GenerateSecret(26)
	}

	hashedSecret, err := u.cfg.GetSecretsHasher().Hash(ctx, []byte(secret))
	if err != nil {
		return err
	}

	client.Secret = string(hashedSecret)

	if client.ID == "" {
		client.ID = x.RandomUUID()
	}

	if client.JWKs == nil {
		client.JWKs = &dbtype.JWKSet{}
	}

	if client.TokenEndpointAuthMethod == "" {
		client.TokenEndpointAuthMethod = "none"
	}

	if client.TokenEndpointAuthSigningAlg == "" {
		client.TokenEndpointAuthSigningAlg = "none"
	}

	client.CreatedAt = x.NowUTC().Round(time.Second)
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
