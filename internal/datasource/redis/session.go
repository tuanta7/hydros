package redis

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	goredis "github.com/redis/go-redis/v9"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/datasource"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/adapter/redis"
	"github.com/tuanta7/hydros/pkg/aead"
)

type TokenSessionStorage struct {
	cfg   *config.Config
	aead  aead.Cipher
	redis redis.Client
}

func NewTokenSessionStorage(cfg *config.Config, rc redis.Client) *TokenSessionStorage {
	gob.Register(datasource.TokenRequestSession{})
	gob.Register(datasource.RefreshRequestSession{})

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

func (s *TokenSessionStorage) GetAccessTokenSession(
	ctx context.Context,
	signature string,
	session core.Session,
) (*core.TokenRequest, error) {
	key := s.prefixKey(core.AccessToken, signature)

	tokenBytes, err := s.redis.Get(ctx, key)
	if errors.Is(err, goredis.Nil) {
		return nil, core.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	var token datasource.TokenRequestSession
	err = gob.NewDecoder(bytes.NewReader(tokenBytes)).Decode(&token)
	if err != nil {
		return nil, err
	}

	return token.ToRequest(ctx, signature, session, core.AccessToken, s.aead)
}

func (s *TokenSessionStorage) CreateAccessTokenSession(ctx context.Context, signature string, req *core.TokenRequest) error {
	session, err := s.sessionFromRequest(ctx, signature, req, core.AccessToken)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(session)
	if err != nil {
		return err
	}

	key := s.prefixKey(core.AccessToken, signature)
	err = s.redis.Set(ctx, key, buf.Bytes(), s.cfg.GetAccessTokenLifetime())
	if err != nil {
		return err
	}

	return nil
}

func (s *TokenSessionStorage) sessionFromRequest(
	ctx context.Context,
	signature string,
	req *core.TokenRequest,
	tokenType core.TokenType,
) (*datasource.TokenRequestSession, error) {
	session, err := json.Marshal(req.Session)
	if err != nil {
		return nil, err
	}

	if s.cfg.Obfuscation.EncryptSessionData {
		ciphertext, err := s.aead.Encrypt(ctx, session, nil)
		if err != nil {
			return nil, err
		}
		session = []byte(ciphertext)
	}

	var challenge sql.NullString
	ss, ok := req.Session.(*domain.Session)
	if !ok && req.Session != nil {
		return nil, fmt.Errorf("expected request to be of type *Session, but got: %T", req.Session)
	} else if ok {
		if len(ss.Challenge) > 0 {
			challenge = sql.NullString{Valid: true, String: ss.Challenge}
		}
	}

	return &datasource.TokenRequestSession{
		Signature:         signature,
		RequestID:         req.ID,
		RequestedAt:       req.RequestedAt,
		ClientID:          req.Client.GetID(),
		Scope:             strings.Join(req.Scope, "|"),
		GrantedScope:      strings.Join(req.GrantedScope, "|"),
		Audience:          strings.Join(req.Audience, "|"),
		GrantedAudience:   strings.Join(req.GrantedAudience, "|"),
		Form:              req.Form.Encode(),
		Session:           session,
		Subject:           req.Session.GetSubject(),
		Active:            true,
		Challenge:         challenge,
		InternalExpiresAt: sql.NullTime{Valid: true, Time: req.Session.GetExpiresAt(tokenType).UTC()},
	}, nil
}

func (s *TokenSessionStorage) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	key := s.prefixKey(core.AccessToken, signature)
	return s.redis.Del(ctx, key)
}
