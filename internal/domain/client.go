package domain

import (
	"time"

	"github.com/tuanta7/hydros/pkg/sqlx"
)

type Client struct {
	ID            string               `json:"client_id" db:"id"`
	Name          string               `json:"client_name" db:"client_name"`
	Description   string               `json:"description" db:"description"`
	Secret        string               `json:"client_secret,omitempty" db:"client_secret"`
	RedirectURIs  sqlx.StringSliceJSON `json:"redirect_uris" db:"redirect_uris"`
	GrantTypes    sqlx.StringSliceJSON `json:"grant_types" db:"grant_types"`
	ResponseTypes sqlx.StringSliceJSON `json:"response_types" db:"response_types"`
	Scope         string               `json:"scope" db:"scope"`
	Audience      sqlx.StringSliceJSON `json:"audience" db:"audience"`
	CreatedAt     time.Time            `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at,omitempty" db:"updated_at"`
	Lifetimes
}

type Lifetimes struct {
	AuthorizationCodeGrantIDTokenLifetime      sqlx.NullDuration `json:"authorization_code_grant_id_token_lifetime,omitempty" db:"authorization_code_grant_id_token_lifetime"`
	AuthorizationCodeGrantAccessTokenLifetime  sqlx.NullDuration `json:"authorization_code_grant_access_token_lifetime,omitempty" db:"authorization_code_grant_access_token_lifetime"`
	AuthorizationCodeGrantRefreshTokenLifetime sqlx.NullDuration `json:"authorization_code_grant_refresh_token_lifetime,omitempty" db:"authorization_code_grant_refresh_token_lifetime"`
	RefreshTokenGrantIDTokenLifetime           sqlx.NullDuration `json:"refresh_token_grant_id_token_lifetime,omitempty" db:"refresh_token_grant_id_token_lifetime"`
	RefreshTokenGrantAccessTokenLifetime       sqlx.NullDuration `json:"refresh_token_grant_access_token_lifetime,omitempty" db:"refresh_token_grant_access_token_lifetime"`
	RefreshTokenGrantRefreshTokenLifetime      sqlx.NullDuration `json:"refresh_token_grant_refresh_token_lifetime,omitempty" db:"refresh_token_grant_refresh_token_lifetime"`
	ClientCredentialsGrantAccessTokenLifetime  sqlx.NullDuration `json:"client_credentials_grant_access_token_lifetime,omitempty" db:"client_credentials_grant_access_token_lifetime"`
}
