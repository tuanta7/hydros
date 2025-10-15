package aead

// XChaCha20Poly1305 is an extension of ChaCha20Poly1305 that use 24-byte nonce
type XChaCha20Poly1305 struct {
	Key        []byte
	RotatedKey [][]byte
}
