package hmac

import (
	"crypto/sha512"
	"hash"
)

type HMAC struct {
	hasher hash.Hash
}

func NewHMAC(tokenEntropy int, secret []byte) (*HMAC, error) {
	return &HMAC{
		hasher: sha512.New,
	}, nil
}