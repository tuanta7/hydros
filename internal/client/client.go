package client

import (
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/pkg/dbtype"
)

// Client represents an OAuth2.1 and IDToken Connect client.
// TODO: Support RFC7591 dynamic client registration
type Client struct {
	ID            string             `json:"id" db:"id"`
	Name          string             `json:"name" db:"name"`
	Description   string             `json:"description" db:"description"`
	Secret        string             `json:"secret,omitempty" db:"secret"`
	Scope         string             `json:"scope" db:"scope"`
	RedirectURIs  dbtype.StringArray `json:"redirect_uris" db:"redirect_uris"`
	GrantTypes    dbtype.StringArray `json:"grant_types" db:"grant_types"`
	ResponseTypes dbtype.StringArray `json:"response_types" db:"response_types"`
	Audience      dbtype.StringArray `json:"audience" db:"audience"`
	RequestURIs   dbtype.StringArray `json:"request_uris,omitempty" db:"request_uris"`
	// JWKs and JWKsURI are mutually exclusive, they are used for authenticate clients using the private_key_jwt method.
	JWKs                        *dbtype.JWKSet `json:"jwks,omitempty" db:"jwks"`
	JWKsURI                     string         `json:"jwks_uri,omitempty" db:"jwks_uri"`
	TokenEndpointAuthMethod     string         `json:"token_endpoint_auth_method,omitempty" db:"token_endpoint_auth_method"`
	TokenEndpointAuthSigningAlg string         `json:"token_endpoint_auth_signing_alg,omitempty" db:"token_endpoint_auth_signing_alg"`
	CreatedAt                   time.Time      `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at,omitempty" db:"updated_at"`
}

func (c *Client) GetID() string {
	return c.ID
}

func (c *Client) GetHashedSecret() []byte {
	return []byte(c.Secret)
}

func (c *Client) GetRedirectURIs() []string {
	return c.RedirectURIs
}

func (c *Client) GetGrantTypes() core.Arguments {
	if len(c.GrantTypes) == 0 {
		return core.Arguments{"authorization_code"}
	}

	return core.Arguments(c.GrantTypes)
}

func (c *Client) GetResponseTypes() core.Arguments {
	return core.Arguments(c.ResponseTypes)
}

func (c *Client) GetScopes() core.Arguments {
	return strings.Fields(c.Scope)
}

func (c *Client) IsPublic() bool {
	return c.TokenEndpointAuthMethod == core.ClientAuthenticationMethodNone
}

func (c *Client) GetAudience() core.Arguments {
	return core.Arguments(c.Audience)
}

func (c *Client) GetRequestURIs() []string {
	return c.RequestURIs
}

func (c *Client) GetJWKs() *jose.JSONWebKeySet {
	if c.JWKs == nil {
		return nil
	}
	return c.JWKs.JSONWebKeySet
}

func (c *Client) GetJWKsURI() string {
	return c.JWKsURI
}

func (c *Client) GetTokenEndpointAuthMethod() string {
	return c.TokenEndpointAuthMethod
}

func (c *Client) GetResponseModes() []core.ResponseMode {
	// fixed for now
	return []core.ResponseMode{
		core.ResponseModeDefault,
		core.ResponseModeQuery,
		core.ResponseModeFormPost,
		core.ResponseModeFragment,
	}
}

func (c *Client) ColumnMap() map[string]any {
	return map[string]any{
		"id":                              c.ID,
		"name":                            c.Name,
		"description":                     c.Description,
		"secret":                          c.Secret,
		"scope":                           c.Scope,
		"redirect_uris":                   c.RedirectURIs,
		"grant_types":                     c.GrantTypes,
		"response_types":                  c.ResponseTypes,
		"audience":                        c.Audience,
		"request_uris":                    c.RequestURIs,
		"jwks":                            c.JWKs,
		"jwks_uri":                        c.JWKsURI,
		"token_endpoint_auth_method":      c.TokenEndpointAuthMethod,
		"token_endpoint_auth_signing_alg": c.TokenEndpointAuthSigningAlg,
		"created_at":                      c.CreatedAt,
		"updated_at":                      c.UpdatedAt,
	}
}
