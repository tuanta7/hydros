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

type Strategy struct {
	entropy int
	secret  []byte
	// rotatedSecrets [][]byte // currently not used
	hasher func() hash.Hash
}

func NewHMAC(secret []byte, entropy int) (*Strategy, error) {
	if len(secret) < 64 {
		return nil, fmt.Errorf("secret for signing HMAC-SHA512/256 is expected to be at least 64 bytes long, got %d bytes", len(secret))
	}

	if entropy < 64 {
		return nil, fmt.Errorf("entropy for HMAC-SHA512/256 is expected to be at least 64 bytes, got %d bytes", entropy)
	}

	return &Strategy{
		entropy: entropy,
		secret:  secret,
		hasher:  sha512.New512_256, // default to HMAC-SHA512/256
	}, nil
}

func (s *Strategy) Generate(request *core.TokenRequest) (string, string, error) {
	tokenKey, err := randomBytes(s.entropy)
	if err != nil {
		return "", "", err
	}

	tokenKey = append(tokenKey, []byte(request.ID)...)
	signature := s.hmacSign(tokenKey)

	encodedSignature := b64.EncodeToString(signature)
	encodedTokenKey := fmt.Sprintf("%s.%s", b64.EncodeToString(tokenKey), encodedSignature)
	return encodedTokenKey, encodedSignature, nil
}

func (s *Strategy) GetSignature(token string) string {
	split := strings.Split(token, ".")
	if len(split) != 2 {
		return ""
	}
	return split[1]
}

func (s *Strategy) Validate(token string) (err error) {
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

	expectedSignature := s.hmacSign(decodedTokenKey)
	if !hmac.Equal(expectedSignature, decodedTokenSignature) {
		return core.ErrTokenSignatureMismatch
	}

	return nil
}

func (s *Strategy) hmacSign(tokenKey []byte) []byte {
	h := hmac.New(s.hasher, s.secret[:64])
	_, err := h.Write(tokenKey)
	if err != nil {
		panic(err) // Write to hash never returns an error
	}

	return h.Sum(nil)
}

func randomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}
