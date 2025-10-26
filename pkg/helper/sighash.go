package helper

import (
	"crypto/sha512"
	"fmt"
)

// SignatureHash hashes the signature to prevent errors where the signature is
// longer than 128 characters (and thus doesn't fit into the pk).
func SignatureHash(signature string) string {
	return fmt.Sprintf("%x", sha512.Sum384([]byte(signature)))
}
