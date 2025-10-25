package strategy

import "github.com/tuanta7/hydros/core"

// TokenStrategy defines the methods used for managing token or authorization code
type TokenStrategy interface {
	Generate(request *core.TokenRequest) (token string, signature string, err error)
	GetSignature(token string) string
	Validate(token string) (err error)
}
