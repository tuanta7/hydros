package helper

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateSecret(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// rand.Read never returns an error
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(b)
}
