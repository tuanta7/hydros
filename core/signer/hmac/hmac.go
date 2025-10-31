package hmac

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/tuanta7/hydros/core"
)

var b64 = base64.URLEncoding.WithPadding(base64.NoPadding)

type SignerConfigurator interface {
	core.TokenEntropyProvider
	core.GlobalSecretProvider
	core.HMACHashingProvider
}

type DefaultSigner struct {
	config SignerConfigurator
}

func NewSigner(config SignerConfigurator) (*DefaultSigner, error) {
	secret := config.GetGlobalSecret()
	if len(secret) < 64 {
		return nil, fmt.Errorf("secret for signing HMAC-SHA512/256 is expected to be at least 64 bytes long, got %d bytes", len(secret))
	}

	return &DefaultSigner{
		config: config,
	}, nil
}

func (s *DefaultSigner) Generate(ctx context.Context, request *core.Request, _ core.TokenType) (string, string, error) {
	entropy := s.config.GetTokenEntropy()
	if entropy < 64 {
		entropy = 64
	}

	tokenKey, err := randomBytes(entropy)
	if err != nil {
		return "", "", err
	}

	secret := s.config.GetGlobalSecret()
	if len(secret) < 64 {
		return "", "", fmt.Errorf("secret for signing HMAC-SHA512/256 is expected to be at least 64 bytes long, got %d bytes", len(secret))
	}

	tokenKey = append(tokenKey, []byte(request.ID)...)
	signingKey := secret[:64]
	signature := s.hmacSign(tokenKey, signingKey)

	encodedSignature := b64.EncodeToString(signature)
	encodedTokenKey := fmt.Sprintf("%s.%s", b64.EncodeToString(tokenKey), encodedSignature)
	return encodedTokenKey, encodedSignature, nil
}

func (s *DefaultSigner) GetSignature(token string) string {
	split := strings.Split(token, ".")
	if len(split) != 2 {
		return ""
	}
	return split[1]
}

func (s *DefaultSigner) Validate(ctx context.Context, token string) (err error) {
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

	secret := s.config.GetGlobalSecret()
	if len(secret) < 64 {
		return fmt.Errorf("secret for signing HMAC-SHA512/256 is expected to be at least 64 bytes long, got %d bytes", len(secret))
	}

	expectedSignature := s.hmacSign(decodedTokenKey, secret[:64])
	if !hmac.Equal(expectedSignature, decodedTokenSignature) {
		return core.ErrTokenSignatureMismatch
	}

	return nil
}

func (s *DefaultSigner) hmacSign(tokenKey []byte, secret []byte) []byte {
	hasher := s.config.GetHMACHasher()
	if hasher == nil {
		hasher = sha512.New512_256
	}

	h := hmac.New(hasher, secret)
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
