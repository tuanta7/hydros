package strategy

import "github.com/tuanta7/hydros/core"

type Signer interface {
	Generate(request *core.TokenRequest, tokenType core.TokenType) (token string, signature string, err error)
	GetSignature(token string) string
	Validate(token string) (err error)
}
