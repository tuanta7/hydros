package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type IDTokenClaims struct {
	jwt.RegisteredClaims
	Nonce           string         `json:"nonce"`
	RequestedAt     time.Time      `json:"rat"`
	AuthTime        time.Time      `json:"auth_time"`
	ACR             string         `json:"acr"`
	AMR             []string       `json:"amr"`
	CodeHash        string         `json:"c_hash"`
	AccessTokenHash string         `json:"at_hash"`
	Extra           map[string]any `json:"ext"`
}

type Claims struct {
	jwt.RegisteredClaims
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
	AuthTime int64  `json:"auth_time,omitempty"`
	ACR      string `json:"acr,omitempty"`
	AMR      string `json:"amr,omitempty"`
}
