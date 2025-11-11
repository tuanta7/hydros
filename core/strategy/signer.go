package strategy

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tuanta7/hydros/core"
)

type OpaqueSigner interface {
	Generate(ctx context.Context, request *core.Request) (token string, signature string, err error)
	Validate(ctx context.Context, token string) (err error)
	GetSignature(token string) string
}

type JWTSigner interface {
	Generate(ctx context.Context, claims jwt.Claims, headers ...map[string]any) (token string, signature string, err error)
	Validate(ctx context.Context, token string) (err error)
	GetSignature(token string) string
}
