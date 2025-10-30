package strategy

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type Signer interface {
	GetSignature(token string) string
	Generate(ctx context.Context, request *core.TokenRequest, tokenType core.TokenType) (token string, signature string, err error)
	Validate(ctx context.Context, token string) (err error)
}
