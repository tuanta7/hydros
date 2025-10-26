package redis

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/adapter/redis"
)

type TokenSessionStorage struct {
	cfg   *config.Config
	redis redis.Client
}

func NewTokenSessionStorage(cfg *config.Config, rc redis.Client) *TokenSessionStorage {
	gob.Register(domain.TokenRequestSession{})
	gob.Register(&domain.TokenRequestSession{})
	gob.Register(domain.RefreshRequestSession{})
	gob.Register(&domain.RefreshRequestSession{})

	return &TokenSessionStorage{
		cfg:   cfg,
		redis: rc,
	}
}

func (s *TokenSessionStorage) prefixKey(tokenType core.TokenType, signature string) string {
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

func (s *TokenSessionStorage) GetAccessTokenSession(ctx context.Context, signature string, session core.Session) (*core.TokenRequest, error) {
	key := s.prefixKey(core.AccessToken, signature)

	tokenBytes, err := s.redis.Get(ctx, key)
	if errors.Is(err, goredis.Nil) {
		return nil, core.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	var token domain.TokenRequestSession
	err = gob.NewDecoder(bytes.NewReader(tokenBytes)).Decode(&token)
	if err != nil {
		return nil, err
	}

	request := &core.TokenRequest{
		Request: core.Request{
			ID:          token.RequestID,
			RequestedAt: token.RequestedAt,
			Session:     session,
		},
	}

	return request, nil
}

func (s *TokenSessionStorage) CreateAccessTokenSession(ctx context.Context, signature string, req *core.TokenRequest) error {
	session := requestToSession(req, signature, core.AccessToken)

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(session)
	if err != nil {
		return err
	}

	key := s.prefixKey(core.AccessToken, signature)
	err = s.redis.Set(ctx, key, buf.Bytes(), s.cfg.Lifetime.AccessToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *TokenSessionStorage) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	key := s.prefixKey(core.AccessToken, signature)
	return s.redis.Del(ctx, key)
}

func requestToSession(req *core.TokenRequest, signature string, tokenType core.TokenType) *domain.TokenRequestSession {
	session := &domain.TokenRequestSession{
		Signature:         signature,
		RequestID:         req.ID,
		RequestedAt:       req.RequestedAt,
		ClientID:          req.Client.GetID(),
		Subject:           req.Session.GetSubject(),
		Active:            true,
		InternalExpiresAt: sql.NullTime{Time: req.Session.GetExpiresAt(tokenType).UTC(), Valid: true},
	}

	return session
}
