package domain

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
)

type Set string
type Algorithm string

const (
	IDTokenSet     Set = "id-token"
	AccessTokenSet Set = "access-token"

	AlgorithmHS256 Algorithm = "HS256"
	AlgorithmHS512 Algorithm = "HS512"
	AlgorithmRS256 Algorithm = "RS256"
	AlgorithmRS512 Algorithm = "RS512"

	// AlgorithmES256 and AlgorithmES512 are not supported yet
	AlgorithmES256 Algorithm = "ES256"
	AlgorithmES512 Algorithm = "ES512"
)

type JSONWebKey struct {
	KeyID     string
	Key       any
	Algorithm Algorithm
	Use       string
	Set       Set
	Active    bool
}

func (dj JSONWebKey) GetPublicKey() any {
	switch t := dj.Key.(type) {
	case *rsa.PrivateKey:
		return t.Public()
	case *rsa.PublicKey:
		return t
	case string: // HS256, HS512
		return []byte(t)
	case []byte:
		return t
	default:
		return t
	}
}

func (dj JSONWebKey) GetPrivateKey() any {
	switch t := dj.Key.(type) {
	case *rsa.PrivateKey:
		return t
	case string: // HS256, HS512
		return []byte(t)
	case []byte:
		return t
	default:
		return t
	}
}

func (dj JSONWebKey) GetKeyID() string {
	return dj.KeyID
}

func (dj JSONWebKey) GetAlgorithm() jwt.SigningMethod {
	switch dj.Algorithm {
	case AlgorithmRS256:
		return jwt.SigningMethodRS256
	case AlgorithmRS512:
		return jwt.SigningMethodRS512
	case AlgorithmHS256:
		return jwt.SigningMethodHS256
	case AlgorithmHS512:
		return jwt.SigningMethodHS512
	}

	return jwt.SigningMethodNone
}

func (dj JSONWebKey) GetUse() string {
	return dj.Use
}
