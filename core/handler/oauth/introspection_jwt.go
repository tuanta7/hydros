package oauth

import (
	"context"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/strategy"
)

type JWTIntrospectionHandler struct {
	tokenStrategy strategy.AccessTokenStrategy
}

func NewJWTIntrospectionHandler(tokenStrategy strategy.AccessTokenStrategy) *JWTIntrospectionHandler {
	return &JWTIntrospectionHandler{
		tokenStrategy: tokenStrategy,
	}
}

func (h *JWTIntrospectionHandler) IntrospectToken(
	ctx context.Context,
	ir *core.IntrospectionRequest,
	tr *core.TokenRequest,
) (core.TokenType, error) {
	return "", nil
}
