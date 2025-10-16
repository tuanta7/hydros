package domain

import (
	"time"

	"github.com/tuanta7/oauth-server/pkg/sqlxx"
)

type Client struct {
	// The ID is immutable. If no ID is provided, a UUID4 will be generated.
	ID          string `json:"client_id" db:"id"`
	Name        string `json:"client_name" db:"client_name"`
	Description string `json:"description" db:"description"`

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
	Lifetimes
}

// Lifetimes is a struct that holds the custom lifetimes for each supported grant_type of a client.
type Lifetimes struct {
	// A maximum authorization code lifetime of 10 minutes is RECOMMENDED.
	// The authorization code is bound to the client_id, code_challenge and redirect_uri
	AuthorizationCodeLifetime sqlxx.NullDuration `json:"authorization_code_lifetime,omitempty" db:"authorization_code_lifetime"`

	AuthorizationCodeGrantIDTokenLifetime      sqlxx.NullDuration `json:"authorization_code_grant_id_token_lifetime,omitempty" db:"authorization_code_grant_id_token_lifetime"`
	AuthorizationCodeGrantAccessTokenLifetime  sqlxx.NullDuration `json:"authorization_code_grant_access_token_lifetime,omitempty" db:"authorization_code_grant_access_token_lifetime"`
	AuthorizationCodeGrantRefreshTokenLifetime sqlxx.NullDuration `json:"authorization_code_grant_refresh_token_lifetime,omitempty" db:"authorization_code_grant_refresh_token_lifetime"`
	RefreshTokenGrantIDTokenLifetime           sqlxx.NullDuration `json:"refresh_token_grant_id_token_lifetime,omitempty" db:"refresh_token_grant_id_token_lifetime"`
	RefreshTokenGrantAccessTokenLifetime       sqlxx.NullDuration `json:"refresh_token_grant_access_token_lifetime,omitempty" db:"refresh_token_grant_access_token_lifetime"`
	RefreshTokenGrantRefreshTokenLifetime      sqlxx.NullDuration `json:"refresh_token_grant_refresh_token_lifetime,omitempty" db:"refresh_token_grant_refresh_token_lifetime"`
	ClientCredentialsGrantAccessTokenLifetime  sqlxx.NullDuration `json:"client_credentials_grant_access_token_lifetime,omitempty" db:"client_credentials_grant_access_token_lifetime"`
}
