package jwt

import "github.com/golang-jwt/jwt/v5"

type Headers struct {
	Algorithm jwt.SigningMethod `json:"alg"`
}
