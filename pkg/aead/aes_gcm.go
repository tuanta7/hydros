package aead

type AESGCM struct {
	Key        []byte
	RotatedKey [][]byte
}
