package core

import "github.com/golang-jwt/jwt/v5"

var (
	AlgorithmMap = map[string]jwt.SigningMethod{
		"RS512": jwt.SigningMethodRS512,
		"RS256": jwt.SigningMethodRS256,
		"HS512": jwt.SigningMethodHS512,
		"HS256": jwt.SigningMethodHS256,
		"ES256": jwt.SigningMethodES256,
		"ES512": jwt.SigningMethodES512,
	}
)
