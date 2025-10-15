package aead

import "context"

type Cipher interface {
	Encrypt(ctx context.Context, plaintext, additionalData []byte) (ciphertext string, err error)
	Decrypt(ctx context.Context, ciphertext string, additionalData []byte) (plaintext []byte, err error)
}
