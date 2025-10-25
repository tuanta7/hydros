package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/adapters/redis"
)

type AccessTokenRepository struct {
	cfg   *config.Config
	redis redis.Client
}

func NewAccessTokenRepository(cfg *config.Config, rc redis.Client) *AccessTokenRepository {
	gob.Register(domain.AccessToken{})
	gob.Register(&domain.AccessToken{})

	return &AccessTokenRepository{
		cfg:   cfg,
		redis: rc,
	}
}

func (r *AccessTokenRepository) prefixKey(signature string) string {
	return fmt.Sprintf("access_token:%s", signature)
}

func (r *AccessTokenRepository) GetBySignature(ctx context.Context, signature string) (*domain.AccessToken, error) {
	key := r.prefixKey(signature)
	tokenBytes, err := r.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var token domain.AccessToken
	err = gob.NewDecoder(bytes.NewReader(tokenBytes)).Decode(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *AccessTokenRepository) Create(ctx context.Context, token *domain.AccessToken) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(token)
	if err != nil {
		return err
	}

	key := r.prefixKey(token.Signature)
	err = r.redis.Set(ctx, key, buf.Bytes(), r.cfg.Lifetime.AccessToken)
	if err != nil {
		return err
	}

	return nil
}

func (r *AccessTokenRepository) Delete(ctx context.Context, signature string) error {
	key := r.prefixKey(signature)
	return r.redis.Del(ctx, key)
}
