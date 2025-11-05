package token

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/pkg/aead"
)

// RequestSessionData is used to store request session in the database.
// It is the direct replacement of OAuth2RequestSQL in ory/hydra.
type RequestSessionData struct {
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
	InternalExpiresAt sql.NullTime `db:"-" json:"-"`
}

func (s *RequestSessionData) ColumnMap() map[string]any {
	return map[string]any{
		"signature":        s.Signature,
		"request_id":       s.RequestID,
		"requested_at":     s.RequestedAt,
		"client_id":        s.ClientID,
		"scope":            s.Scope,
		"granted_scope":    s.GrantedScope,
		"audience":         s.Audience,
		"granted_audience": s.GrantedAudience,
		"form_data":        s.Form,
		"session_data":     s.Session,
		"subject":          s.Subject,
		"active":           s.Active,
		"challenge":        s.Challenge,
		// "expires_at":       s.InternalExpiresAt,
	}
}

func (s *RequestSessionData) ToRequest(
	ctx context.Context,
	signature string,
	session core.Session,
	tokenType core.TokenType,
	aead aead.Cipher,
) (*core.Request, error) {
	jsonSession := s.Session
	if !gjson.ValidBytes(jsonSession) {
		var err error
		jsonSession, err = aead.Decrypt(ctx, string(s.Session), nil)
		if err != nil {
			return nil, err
		}
	}

	if session != nil {
		// use the session parameter to help reconstruct the session data since we only have the JSON formated data
		if err := json.Unmarshal(jsonSession, session); err != nil {
			return nil, err
		}
	} else {
		// if the session parameter is nil, we can't reconstruct the session data, so we just ignore it and
		// return the request with a nil session
	}

	form, err := url.ParseQuery(s.Form)
	if err != nil {
		return nil, err
	}

	return &core.Request{
		ID:              s.RequestID,
		RequestedAt:     s.RequestedAt,
		Scope:           strings.Split(s.Scope, "|"),
		GrantedScope:    strings.Split(s.GrantedScope, "|"),
		Audience:        strings.Split(s.Audience, "|"),
		GrantedAudience: strings.Split(s.GrantedAudience, "|"),
		Form:            form,
		Client: &client.Client{
			// I have not figured out how to get the full client object from the database like hydra,
			// so just use the ID for now.
			ID: s.ClientID,
		},
		Session: session,
	}, nil
}
