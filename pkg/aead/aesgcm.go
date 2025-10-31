package aead

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type AESGCM struct {
	key []byte
}

func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != 32 {
		// Enforce AES-256 key length
		return nil, fmt.Errorf("key must be exactly 32 long bytes, got %d bytes", len(key))
	}

	return &AESGCM{key}, nil
}

func (c *AESGCM) Encrypt(ctx context.Context, plaintext, additionalData []byte) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, additionalData)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (c *AESGCM) Decrypt(ctx context.Context, encodedCiphertext string, additionalData []byte) ([]byte, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	plaintext, err := gcm.Open(nil, ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():], additionalData)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
