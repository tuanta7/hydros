package core

import (
	"context"
	"net/http"
	"net/url"
)

type TokenType string
type GrantType string

const (
	BearerToken = "Bearer"

	AccessToken       TokenType = "access_token"
	RefreshToken      TokenType = "refresh_token"
	AuthorizationCode TokenType = "authorize_code"
	IDToken           TokenType = "id_token"
	PKCECode          TokenType = "pkce_code"

	GrantTypeRefreshToken      GrantType = "refresh_token"
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	GrantTypeClientCredentials GrantType = "client_credentials"
)

// OAuth2 implements the OAuth2Provider interface.
type OAuth2 struct {
	config               Configurator
	store                Storage
	authorizeHandlers    []AuthorizeHandler
	tokenHandlers        []TokenHandler
	introspectionHandler []IntrospectionHandler
}

func NewOAuth2(
	config Configurator,
	store Storage,
	authorizeHandlers []AuthorizeHandler,
	tokenHandlers []TokenHandler,
	introspectionHandler []IntrospectionHandler,
) *OAuth2 {
	return &OAuth2{
		config:               config,
		store:                store,
		authorizeHandlers:    authorizeHandlers,
		tokenHandlers:        tokenHandlers,
		introspectionHandler: introspectionHandler,
	}
}

type OAuth2Provider interface {
	NewAuthorizeRequest(ctx context.Context, req *http.Request) (*AuthorizeRequest, error)
	NewAuthorizeResponse(ctx context.Context, req *AuthorizeRequest, session Session) (*AuthorizeResponse, error)
	WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, err error)
	WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, resp *AuthorizeResponse)

	AuthenticateClient(ctx context.Context, r *http.Request, form url.Values) (Client, error)
	NewTokenRequest(ctx context.Context, req *http.Request, session Session) (*TokenRequest, error)
	NewTokenResponse(ctx context.Context, req *TokenRequest) (*TokenResponse, error)
	WriteTokenError(ctx context.Context, rw http.ResponseWriter, req *TokenRequest, err error)
	WriteTokenResponse(ctx context.Context, rw http.ResponseWriter, req *TokenRequest, resp *TokenResponse)

	IntrospectToken(ctx context.Context, req *http.Request, session Session) (*IntrospectionResponse, error)
	WriteIntrospectionError(ctx context.Context, rw http.ResponseWriter, err error)
	WriteIntrospectionResponse(ctx context.Context, rw http.ResponseWriter, r *IntrospectionResponse)
}

type Configurator interface {
	DebugModeProvider
	MinParameterEntropyProvider
	SecretsHasherProvider
}

type Storage interface {
	GetClient(ctx context.Context, id string) (Client, error)
}

type AuthorizeHandler interface {
	HandleAuthorizeRequest(ctx context.Context, req *AuthorizeRequest) error
	HandleAuthorizeResponse(ctx context.Context, req *AuthorizeRequest, res *AuthorizeResponse) error
}

type TokenHandler interface {
	HandleTokenRequest(ctx context.Context, req *TokenRequest) error
	HandleTokenResponse(ctx context.Context, req *TokenRequest, res *TokenResponse) error
}

type IntrospectionHandler interface {
	IntrospectToken(context.Context, *IntrospectionRequest, *TokenRequest) (TokenType, error)
}
