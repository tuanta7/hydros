package core

import (
	"context"
)

type AuthorizeHandlerChain []AuthorizeHandler

type AuthorizeHandler interface {
	HandleAuthorizeRequest(ctx context.Context, req *AuthorizeRequest, res *AuthorizeResponse) error
}

type TokenHandler interface {
	HandleTokenRequest(ctx context.Context, req *TokenRequest) error
}
