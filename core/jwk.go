package core

import (
	"crypto/rsa"
)

type VerificationKey interface {
	*rsa.PublicKey | []byte
}
type JSONWebKey struct {
	Key       *rsa.PublicKey // fixed for now
	KeyID     string
	Algorithm string
	Use       string
}

type JSONWebKeySet struct {
	Keys []JSONWebKey
}
