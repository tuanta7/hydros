package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type IDTokenClaims struct {
	JTI                                 string         `json:"jti"`
	Issuer                              string         `json:"iss"`
	Subject                             string         `json:"sub"`
	Audience                            []string       `json:"aud"`
	Nonce                               string         `json:"nonce"`
	ExpiresAt                           time.Time      `json:"exp"`
	IssuedAt                            time.Time      `json:"iat"`
	RequestedAt                         time.Time      `json:"rat"`
	AuthenticatedAt                     time.Time      `json:"authenticated_at"`
	AccessTokenHash                     string         `json:"at_hash"`
	AuthenticationContextClassReference string         `json:"acr"`
	AuthenticationMethodsReferences     []string       `json:"amr"`
	CodeHash                            string         `json:"c_hash"`
	Extra                               map[string]any `json:"ext"`
}

type Claims struct {
	jwt.RegisteredClaims
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
	AuthTime int64  `json:"auth_time,omitempty"`
	ACR      string `json:"acr,omitempty"`
	AMR      string `json:"amr,omitempty"`
}
