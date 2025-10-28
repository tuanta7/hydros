package core

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
	GetJWKs() *JSONWebKeySet
	GetJWKsURI() string
	GetTokenEndpointAuthMethod() string
}
