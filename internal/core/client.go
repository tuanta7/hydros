package core

import "github.com/golang-jwt/jwt/v5"

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
	GetJWKs() *jwt.VerificationKeySet
	GetJWKsURI() string
	GetTokenEndpointAuthMethod() string
}
