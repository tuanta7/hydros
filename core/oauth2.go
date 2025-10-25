package core

import (
	"context"
)

type TokenType string

const (
	BearerToken = "Bearer"

	AccessToken   TokenType = "access_token"
	RefreshToken  TokenType = "refresh_token"
	AuthorizeCode TokenType = "authorize_code"
	IDToken       TokenType = "id_token"
)

type AuthorizeHandler interface {
	HandleAuthorizeRequest(ctx context.Context, req *AuthorizeRequest, res *AuthorizeResponse) error
}

type TokenHandler interface {
	HandleTokenRequest(ctx context.Context, req *TokenRequest, res *TokenResponse) error
}
