package oidc

import (
	"context"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type IDTokenSigner interface {
	GenerateWithClaims(ctx context.Context, claims gojwt.Claims) (string, string, error)
	Validate(ctx context.Context, token string) (err error)
	GetSignature(token string) string
}
