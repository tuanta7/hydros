package core

import "github.com/golang-jwt/jwt/v5"

// JSONWebKeySet or JKWs is a JSON array containing the public keys a server
// uses to verify the signature on JWTs it receives.
type JSONWebKeySet struct {
	jwt.VerificationKeySet `json:"keys"`
}

// JSONWebKey is a key format used to represent cryptographic keys.
type JSONWebKey interface {
	GetPrivateKey() any // private or secret key
	GetPublicKey() any
	GetKeyID() string
	GetAlgorithm() jwt.SigningMethod
	GetUse() string
}
