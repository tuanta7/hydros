package aead

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type AESGCM struct {
	Key []byte
}

func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be exactly 32 long bytes, got %d bytes", len(key))
	}

	return &AESGCM{key}, nil
}

func (c *AESGCM) Encrypt(ctx context.Context, plaintext, additionalData []byte) (string, error) {
	ciphertext, err := encrypt(plaintext, c.Key, additionalData)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func encrypt(plaintext []byte, key []byte, additionalData []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, additionalData), nil
}
