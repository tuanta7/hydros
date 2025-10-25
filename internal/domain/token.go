package domain

import (
	"database/sql"
	"time"
)

type AccessToken struct {
	Signature   string    `db:"signature"`
	RequestID   string    `db:"request_id"`
	RequestedAt time.Time `db:"requested_at"`
	ClientID    string    `db:"client_id"`
	Subject     string    `db:"subject"`
	Active      bool      `db:"active"`
}

type RefreshToken struct {
	Signature            string         `db:"signature"`
	AccessTokenSignature sql.NullString `db:"access_token_signature"`
	RequestID            string         `db:"request_id"`
	RequestedAt          time.Time      `db:"requested_at"`
	ClientID             string         `db:"client_id"`
	Subject              string         `db:"subject"`
	Active               bool           `db:"active"`
}
