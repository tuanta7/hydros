package token

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
	"github.com/tuanta7/hydros/internal/session"
	"github.com/tuanta7/hydros/pkg/adapter/redis"
	"github.com/tuanta7/hydros/pkg/aead"
)

type RequestSessionCache struct {
	cfg   *config.Config
	aead  aead.Cipher
	redis redis.Client
}

func NewRequestSessionCache(cfg *config.Config, aead aead.Cipher, rc redis.Client) *RequestSessionCache {
	gob.Register(RequestSessionData{})

	return &RequestSessionCache{
		cfg:   cfg,
		aead:  aead,
		redis: rc,
	}
}

func (c *RequestSessionCache) prefixKey(tokenType core.TokenType, signature string) string {
	switch tokenType {
	case core.AccessToken:
		return fmt.Sprintf("access_token:%c", signature)
	case core.RefreshToken:
		return fmt.Sprintf("refresh_token:%c", signature)
	case core.AuthorizationCode:
		return fmt.Sprintf("authorize_code:%c", signature)
	case core.IDToken:
		return fmt.Sprintf("id_token:%c", signature)
	}

	return signature
}

func (c *RequestSessionCache) GetAccessTokenSession(
	ctx context.Context,
	signature string,
	session core.Session,
) (*core.Request, error) {
	key := c.prefixKey(core.AccessToken, signature)

	tokenBytes, err := c.redis.Get(ctx, key)
	if errors.Is(err, goredis.Nil) {
		return nil, core.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	var token RequestSessionData
	err = gob.NewDecoder(bytes.NewReader(tokenBytes)).Decode(&token)
	if err != nil {
		return nil, err
	}

	return token.ToRequest(ctx, signature, session, core.AccessToken, c.aead)
}

func (c *RequestSessionCache) CreateAccessTokenSession(ctx context.Context, signature string, req *core.Request) error {
	session, err := c.sessionFromRequest(ctx, signature, req, core.AccessToken)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(session)
	if err != nil {
		return err
	}

	key := c.prefixKey(core.AccessToken, signature)
	err = c.redis.Set(ctx, key, buf.Bytes(), c.cfg.GetAccessTokenLifetime())
	if err != nil {
		return err
	}

	return nil
}

func (c *RequestSessionCache) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	key := c.prefixKey(core.AccessToken, signature)
	return c.redis.Del(ctx, key)
}

func (c *RequestSessionCache) CreateAuthorizeCodeSession(
	ctx context.Context,
	signature string,
	req *core.Request,
) (err error) {
	session, err := c.sessionFromRequest(ctx, signature, req, core.AuthorizationCode)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(session)
	if err != nil {
		return err
	}

	key := c.prefixKey(core.AuthorizationCode, signature)
	err = c.redis.Set(ctx, key, buf.Bytes(), c.cfg.GetAuthorizationCodeLifetime())
	if err != nil {
		return err
	}

	return nil
}

func (c *RequestSessionCache) GetAuthorizationCodeSession(ctx context.Context, signature string, session core.Session) (*core.Request, error) {
	return nil, nil
}

func (c *RequestSessionCache) InvalidateAuthorizeCodeSession(ctx context.Context, signature string) (err error) {
	return nil
}

func (c *RequestSessionCache) sessionFromRequest(
	ctx context.Context,
	signature string,
	req *core.Request,
	tokenType core.TokenType,
) (*RequestSessionData, error) {
	s, err := json.Marshal(req.Session)
	if err != nil {
		return nil, err
	}

	if c.cfg.Obfuscation.EncryptSessionData {
		ciphertext, err := c.aead.Encrypt(ctx, s, nil)
		if err != nil {
			return nil, err
		}
		s = []byte(ciphertext)
	}

	var challenge sql.NullString
	ss, ok := req.Session.(*session.Session)
	if !ok && req.Session != nil {
		return nil, fmt.Errorf("expected request to be of type *Session, but got: %T", req.Session)
	} else if ok {
		if len(ss.Challenge) > 0 {
			challenge = sql.NullString{Valid: true, String: ss.Challenge}
		}
	}

	return &RequestSessionData{
		Signature:         signature,
		RequestID:         req.ID,
		RequestedAt:       req.RequestedAt,
		ClientID:          req.Client.GetID(),
		Scope:             strings.Join(req.Scope, "|"),
		GrantedScope:      strings.Join(req.GrantedScope, "|"),
		Audience:          strings.Join(req.Audience, "|"),
		GrantedAudience:   strings.Join(req.GrantedAudience, "|"),
		Form:              req.Form.Encode(),
		Session:           s,
		Subject:           req.Session.GetSubject(),
		Active:            true,
		Challenge:         challenge,
		InternalExpiresAt: sql.NullTime{Valid: true, Time: req.Session.GetExpiresAt(tokenType).UTC()},
	}, nil
}
