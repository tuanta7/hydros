package hmac

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"strings"

	"github.com/tuanta7/hydros/core"
)

var b64 = base64.URLEncoding.WithPadding(base64.NoPadding)

type HMAC struct {
	entropy        int
	secret         []byte
	rotatedSecrets [][]byte // currently not used
	hasher         func() hash.Hash
}

func NewHMAC(secret []byte, entropy int) (*HMAC, error) {
	if len(secret) < 64 {
		return nil, fmt.Errorf("secret for signing HMAC-SHA512/256 is expected to be at least 64 bytes long, got %d bytes", len(secret))
	}

	if entropy < 64 {
		return nil, fmt.Errorf("entropy for HMAC-SHA512/256 is expected to be at least 64 bytes, got %d bytes", entropy)
	}

	return &HMAC{
		entropy: entropy,
		secret:  secret,
		hasher:  sha512.New512_256, // default to HMAC-SHA512/256
	}, nil
}

func (h *HMAC) Generate(request *core.TokenRequest) (string, string, error) {
	tokenKey, err := randomBytes(h.entropy)
	if err != nil {
		return "", "", err
	}

	tokenKey = append(tokenKey, []byte(request.ID)...)

	hs := hmac.New(h.hasher, h.secret[:64])
	_, err = hs.Write(tokenKey)
	if err != nil {
		panic(err) // Write to hash never returns an error
	}
	signature := hs.Sum(nil)

	encodedSignature := b64.EncodeToString(signature)
	encodedTokenKey := fmt.Sprintf("%s.%s", b64.EncodeToString(tokenKey), encodedSignature)
	return encodedTokenKey, encodedSignature, nil
}

func (h *HMAC) GetSignature(token string) string {
	split := strings.Split(token, ".")
	if len(split) != 2 {
		return ""
	}
	return split[1]
}

func (h *HMAC) Validate(token string) (err error) {
	tokenKey, tokenSignature, ok := strings.Cut(token, ".")
	if !ok {
		return core.ErrInvalidTokenFormat
	}

	if tokenKey == "" || tokenSignature == "" {
		return core.ErrInvalidTokenFormat
	}

	decodedTokenSignature, err := b64.DecodeString(tokenSignature)
	if err != nil {
		return err
	}

	decodedTokenKey, err := b64.DecodeString(tokenKey)
	if err != nil {
		return err
	}

	hs := hmac.New(h.hasher, h.secret[:64])
	_, err = hs.Write(decodedTokenKey)
	if err != nil {
		panic(err) // Write to hash never returns an error
	}
	expectedSignature := hs.Sum(nil)

	if !hmac.Equal(expectedSignature, decodedTokenSignature) {
		return core.ErrTokenSignatureMismatch
	}

	return nil
}

func randomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}
