package token

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/session"
	"github.com/tuanta7/hydros/pkg/aead"
)

const (
	PKCE = "pkce"
	OIDC = "oidc"
)

type RequestSessionStorage struct {
	cfg  *config.Config
	aead aead.Cipher
	pg   *RequestSessionRepo
	rd   *RequestSessionCache
}

func NewRequestSessionStorage(
	cfg *config.Config,
	aead aead.Cipher,
	pg *RequestSessionRepo,
	rd *RequestSessionCache,
) *RequestSessionStorage {
	return &RequestSessionStorage{
		cfg:  cfg,
		aead: aead,
		pg:   pg,
		rd:   rd,
	}
}

func (r *RequestSessionStorage) CreateAccessTokenSession(ctx context.Context, signature string, req *core.Request) error {
	s, err := r.sessionFromRequest(ctx, signature, req, core.AccessToken)
	if err != nil {
		return err
	}

	return r.pg.Create(ctx, core.AccessToken, s)
}

func (r *RequestSessionStorage) GetAccessTokenSession(ctx context.Context, signature string, session core.Session) (*core.Request, error) {
	s, err := r.pg.GetBySignature(ctx, core.AccessToken, signature)
	if err != nil {
		return nil, err
	}

	return s.ToRequest(ctx, signature, session, core.AccessToken, r.aead)
}

func (r *RequestSessionStorage) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return r.pg.DeleteBySignature(ctx, core.AccessToken, signature)
}

func (r *RequestSessionStorage) CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, req *core.Request) (err error) {
	return nil
}

func (r *RequestSessionStorage) GetRefreshTokenSession(ctx context.Context, signature string, session core.Session) (*core.Request, error) {
	return nil, nil
}

func (r *RequestSessionStorage) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	return nil
}

func (r *RequestSessionStorage) RotateRefreshToken(ctx context.Context, requestID string, signature string) (err error) {
	return nil
}

func (r *RequestSessionStorage) CreateAuthorizeCodeSession(ctx context.Context, code string, req *core.Request) (err error) {
	s, err := r.sessionFromRequest(ctx, code, req, core.AuthorizationCode)
	if err != nil {
		return err
	}

	return r.rd.CreateCodeSession(ctx, code, s)
}

func (r *RequestSessionStorage) GetAuthorizationCodeSession(ctx context.Context, code string, session core.Session) (*core.Request, error) {
	s, err := r.rd.GetCodeSession(ctx, code)
	if err != nil {
		return nil, err
	}

	return s.ToRequest(ctx, code, session, core.AuthorizationCode, r.aead)
}

func (r *RequestSessionStorage) InvalidateAuthorizeCodeSession(ctx context.Context, signature string) (err error) {
	return r.rd.DeleteCodeSession(ctx, signature)
}

func (r *RequestSessionStorage) GetPKCERequestSession(ctx context.Context, authorizeCode string, session core.Session) (*core.Request, error) {
	s, err := r.pg.GetBySignature(ctx, PKCE, authorizeCode)
	if err != nil {
		return nil, err
	}

	return s.ToRequest(ctx, authorizeCode, session, PKCE, r.aead)
}

func (r *RequestSessionStorage) CreatePKCERequestSession(ctx context.Context, authorizeCode string, req *core.Request) error {
	s, err := r.sessionFromRequest(ctx, authorizeCode, req, PKCE)
	if err != nil {
		return err
	}

	return r.pg.Create(ctx, PKCE, s)
}

func (r *RequestSessionStorage) DeletePKCERequestSession(ctx context.Context, authorizeCode string) error {
	return r.pg.DeleteBySignature(ctx, PKCE, authorizeCode)
}

func (r *RequestSessionStorage) GetOpenIDConnectSession(ctx context.Context, authorizeCode string, session core.Session) (*core.Request, error) {
	s, err := r.pg.GetBySignature(ctx, OIDC, authorizeCode)
	if err != nil {
		return nil, err
	}

	return s.ToRequest(ctx, authorizeCode, session, OIDC, r.aead)
}

func (r *RequestSessionStorage) CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, req *core.Request) error {
	s, err := r.sessionFromRequest(ctx, authorizeCode, req, OIDC)
	if err != nil {
		return err
	}

	return r.pg.Create(ctx, OIDC, s)
}

func (r *RequestSessionStorage) DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) error {
	return r.pg.DeleteBySignature(ctx, OIDC, authorizeCode)
}

func (r *RequestSessionStorage) sessionFromRequest(
	ctx context.Context,
	signature string,
	req *core.Request,
	tokenType core.TokenType,
) (*RequestSessionData, error) {
	s, err := json.Marshal(req.Session)
	if err != nil {
		return nil, err
	}

	if r.cfg.Obfuscation.EncryptSessionData {
		ciphertext, err := r.aead.Encrypt(ctx, s, nil)
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
