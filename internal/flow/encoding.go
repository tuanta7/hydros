package flow

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"

	"github.com/tuanta7/hydros/pkg/aead"
)

func (f *Flow) EncodeToLoginChallenge(ctx context.Context, cipher aead.Cipher) (string, error) {
	return EncodeFlow(ctx, cipher, f, []byte("login_challenge"))
}

func (f *Flow) EncodeToLoginVerifier(ctx context.Context, cipher aead.Cipher) (string, error) {
	return EncodeFlow(ctx, cipher, f, []byte("login_verifier"))
}

func (f *Flow) EncodeToConsentChallenge(ctx context.Context, cipher aead.Cipher) (string, error) {
	return EncodeFlow(ctx, cipher, f, []byte("consent_challenge"))
}

func (f *Flow) EncodeToConsentVerifier(ctx context.Context, cipher aead.Cipher) (string, error) {
	return EncodeFlow(ctx, cipher, f, []byte("consent_verifier"))
}

func EncodeFlow(ctx context.Context, cipher aead.Cipher, f *Flow, data []byte) (string, error) {
	if f.Client != nil {
		f.ClientID = f.Client.ID
	}

	var bb bytes.Buffer
	gz, err := gzip.NewWriterLevel(&bb, gzip.BestCompression)
	if err != nil {
		return "", err
	}

	if err = json.NewEncoder(gz).Encode(f); err != nil {
		return "", err
	}

	if err = gz.Close(); err != nil {
		return "", err
	}

	return cipher.Encrypt(ctx, bb.Bytes(), data)
}

func DecodeFlow(ctx context.Context, cipher aead.Cipher, encoded string, data []byte) (*Flow, error) {
	plain, err := cipher.Decrypt(ctx, encoded, data)
	if err != nil {
		return nil, err
	}

	var f Flow
	gz, err := gzip.NewReader(bytes.NewReader(plain))
	if err != nil {
		return nil, err
	}

	if err = json.NewDecoder(gz).Decode(&f); err != nil {
		return nil, err
	}

	if err = gz.Close(); err != nil {
		return nil, err
	}

	return &f, nil
}
