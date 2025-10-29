package datasource

import (
	"context"
	"database/sql"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/aead"
)

// TokenRequestSession is used to store the token request in the database.
// It is the direct replacement of OAuth2RequestSQL in ory/fosite
type TokenRequestSession struct {
	Signature       string         `db:"signature"`
	RequestID       string         `db:"request_id"`
	RequestedAt     time.Time      `db:"requested_at"`
	ClientID        string         `db:"client_id"`
	Scope           string         `db:"scope"`
	GrantedScope    string         `db:"granted_scope"`
	Audience        string         `db:"audience"`
	GrantedAudience string         `db:"granted_audience"`
	Form            string         `db:"form_data"`
	Session         []byte         `db:"session_data"`
	Subject         string         `db:"subject"`
	Active          bool           `db:"active"`
	Challenge       sql.NullString `db:"challenge"`

	// InternalExpiresAt denormalizes the expiry from the session to additionally store it as a row.
	InternalExpiresAt sql.NullTime `db:"expires_at" json:"-"`
}

func (s *TokenRequestSession) ToRequest(
	ctx context.Context,
	signature string,
	session core.Session,
	tokenType core.TokenType,
	aead aead.Cipher,
) (*core.TokenRequest, error) {
	jsonSession := s.Session
	if !gjson.ValidBytes(jsonSession) {
		var err error
		jsonSession, err = aead.Decrypt(ctx, string(s.Session), nil)
		if err != nil {
			return nil, err
		}
	}

	form, err := url.ParseQuery(s.Form)
	if err != nil {
		return nil, err
	}

	return &core.TokenRequest{
		Request: core.Request{
			ID:              s.RequestID,
			RequestedAt:     s.RequestedAt,
			Scope:           strings.Split(s.Scope, "|"),
			GrantedScope:    strings.Split(s.GrantedScope, "|"),
			Audience:        strings.Split(s.Audience, "|"),
			GrantedAudience: strings.Split(s.GrantedAudience, "|"),
			Form:            form,
			Client: &domain.Client{
				ID: s.ClientID,
			},
			Session: session,
		},
	}, nil
}

type RefreshRequestSession struct {
	TokenRequestSession
	FirstUsedAt          sql.NullTime   `db:"first_used_at"`
	AccessTokenSignature sql.NullString `db:"access_token_signature"`
}
