package client

import (
	"time"

	"github.com/tuanta7/oauth-server/pkg/sqlxx"
)

type Client struct {
	// The ID is immutable. If no ID is provided, a UUID4 will be generated.
	ID   string `json:"client_id" db:"id"`
	Name string `json:"client_name" db:"client_name"`

	// The secret will be included in the creation request as cleartext, and then
	// never again. The secret is kept in hashed format and is not recoverable once lost.
	Secret        string                      `json:"client_secret,omitempty" db:"client_secret"`
	RedirectURIs  sqlxx.StringSliceJSONFormat `json:"redirect_uris" db:"redirect_uris"`
	GrantTypes    sqlxx.StringSliceJSONFormat `json:"grant_types" db:"grant_types"`
	ResponseTypes sqlxx.StringSliceJSONFormat `json:"response_types" db:"response_types"`
	Scope         string                      `json:"scope" db:"scope"`
	Audience      sqlxx.StringSliceJSONFormat `json:"audience" db:"audience"`
	CreatedAt     time.Time                   `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt     time.Time                   `json:"updated_at,omitempty" db:"updated_at"`
	Lifespans
}

type Lifespans struct{}
