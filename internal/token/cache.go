package token

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/pkg/adapter/redis"
)

type RequestSessionCache struct {
	cfg   *config.Config
	redis redis.Client
}

func NewRequestSessionCache(cfg *config.Config, rc redis.Client) *RequestSessionCache {
	gob.Register(RequestSessionData{})
	return &RequestSessionCache{
		cfg:   cfg,
		redis: rc,
	}
}

func (c *RequestSessionCache) prefixKey(tokenType core.TokenType, signature string) string {
	switch tokenType {
	case core.AccessToken:
		return fmt.Sprintf("access_token:%s", signature)
	case core.RefreshToken:
		return fmt.Sprintf("refresh_token:%s", signature)
	case core.AuthorizationCode:
		return fmt.Sprintf("authorize_code:%s", signature)
	case core.IDToken:
		return fmt.Sprintf("id_token:%s", signature)
	}

	return signature
}

func (c *RequestSessionCache) GetCodeSession(ctx context.Context, code string) (*RequestSessionData, error) {
	key := c.prefixKey(core.AuthorizationCode, code)

	tokenBytes, err := c.redis.Get(ctx, key)
	if errors.Is(err, goredis.Nil) {
		return nil, core.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	var token *RequestSessionData
	err = gob.NewDecoder(bytes.NewReader(tokenBytes)).Decode(&token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (c *RequestSessionCache) CreateCodeSession(ctx context.Context, code string, s *RequestSessionData) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(s)
	if err != nil {
		return err
	}

	key := c.prefixKey(core.AuthorizationCode, code)
	err = c.redis.Set(ctx, key, buf.Bytes(), c.cfg.GetAuthorizationCodeLifetime())
	if err != nil {
		return err
	}

	return nil
}

func (c *RequestSessionCache) DeleteCodeSession(ctx context.Context, code string) error {
	key := c.prefixKey(core.AuthorizationCode, code)
	return c.redis.Del(ctx, key)
}
