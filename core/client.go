package core

import "github.com/go-jose/go-jose/v4"

type Client interface {
	GetID() string
	GetHashedSecret() []byte
	GetRedirectURIs() []string
	GetGrantTypes() Arguments
	GetResponseTypes() Arguments
	GetScopes() Arguments
	IsPublic() bool
	GetAudience() Arguments
}

type OpenIDConnectClient interface {
	GetRequestURIs() []string
	GetJWKs() *jose.JSONWebKeySet
	GetJWKsURI() string
	GetTokenEndpointAuthMethod() string
}
