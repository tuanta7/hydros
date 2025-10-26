package domain

import (
	"database/sql"
	"time"

	"github.com/tuanta7/hydros/core/handler/oidc"
)

type Session struct {
	IDToken   *oidc.IDTokenSession `json:"id_token"`
	Extra     map[string]any       `json:"extra"`
	KeyID     string               `json:"kid"`
	ClientID  string               `json:"client_id"`
	Challenge string               `json:"challenge"`
	Flow      *Flow                `json:"-"`
}

type TokenRequestSession struct {
	Signature         string         `db:"signature"`
	Challenge         sql.NullString `db:"challenge"`
	RequestID         string         `db:"request_id"`
	RequestedAt       time.Time      `db:"requested_at"`
	ClientID          string         `db:"client_id"`
	Scopes            string         `db:"scope"`
	GrantedScope      string         `db:"granted_scope"`
	RequestedAudience string         `db:"requested_audience"`
	GrantedAudience   string         `db:"granted_audience"`
	Form              string         `db:"form_data"`
	Subject           string         `db:"subject"`
	Active            bool           `db:"active"`
	Session           []byte         `db:"session_data"`

	// InternalExpiresAt denormalizes the expiry from the session to additionally store it as a row.
	InternalExpiresAt sql.NullTime `db:"expires_at" json:"-"`
}

type RefreshRequestSession struct {
	TokenRequestSession
	FirstUsedAt          sql.NullTime   `db:"first_used_at"`
	AccessTokenSignature sql.NullString `db:"access_token_signature"`
}
