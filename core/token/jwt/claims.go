package jwt

import "time"

type IDTokenClaims struct {
	JTI                                 string         `json:"jti"`
	Issuer                              string         `json:"iss"`
	Subject                             string         `json:"sub"`
	Audience                            []string       `json:"aud"`
	Nonce                               string         `json:"nonce"`
	ExpiresAt                           time.Time      `json:"exp"`
	IssuedAt                            time.Time      `json:"iat"`
	RequestedAt                         time.Time      `json:"rat"`
	AuthTime                            time.Time      `json:"auth_time"`
	AccessTokenHash                     string         `json:"at_hash"`
	AuthenticationContextClassReference string         `json:"acr"`
	AuthenticationMethodsReferences     []string       `json:"amr"`
	CodeHash                            string         `json:"c_hash"`
	Extra                               map[string]any `json:"ext"`
}
