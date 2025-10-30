package jwt

import (
	"errors"
	"strings"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
)

type Configurator interface {
	core.AccessTokenIssuerProvider
}

type DefaultSigner struct {
	config     Configurator
	privateKey *core.JSONWebKey
}

func NewSigner(cfg Configurator) (*DefaultSigner, error) {
	return &DefaultSigner{
		config: cfg,
	}, nil
}

func (s DefaultSigner) Generate(request *core.TokenRequest, tokenType core.TokenType) (string, string, error) {
	claims := &Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    s.config.GetAccessTokenIssuer(),
			Subject:   request.Session.GetSubject(),
			Audience:  gojwt.ClaimStrings(request.GrantedAudience),
			IssuedAt:  gojwt.NewNumericDate(x.NowUTC()),
			ExpiresAt: gojwt.NewNumericDate(request.Session.GetExpiresAt(tokenType)),
		},
		ClientID: request.Client.GetID(),
		Scope:    strings.Join(request.GrantedScope, " "),
	}

	// TODO: Finish implementation
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS512, claims)
	signedToken, err := token.SignedString([]byte("PRIVATE-KEY-OR-SECRET"))
	if err != nil {
		return "", "", err
	}

	return signedToken, string(token.Signature), nil
}

func (s DefaultSigner) GetSignature(token string) string {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return ""
	}
	return parts[2]
}

func (s DefaultSigner) Validate(token string) (err error) {
	parser := gojwt.Parser{}
	claims := &Claims{}

	t, err := parser.ParseWithClaims(token, claims, func(t *gojwt.Token) (any, error) {
		return []byte("PUBLIC-KEY-OR-SECRET"), nil
	})
	if err != nil {
		return err
	}

	if t.Method == gojwt.SigningMethodNone {
		return errors.New("token signing method none is not allowed")
	}

	return nil
}
