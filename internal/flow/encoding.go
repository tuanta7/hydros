package flow

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"

	"github.com/tuanta7/hydros/pkg/aead"
)

type AdditionalData []byte

var (
	AsLoginChallenge   AdditionalData = []byte("login_challenge")
	AsLoginVerifier    AdditionalData = []byte("login_verifier")
	AsConsentChallenge AdditionalData = []byte("consent_challenge")
	AsConsentVerifier  AdditionalData = []byte("consent_verifier")
)

func EncodeFlow(ctx context.Context, cipher aead.Cipher, f *Flow, data AdditionalData) (string, error) {
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

func DecodeFlow(ctx context.Context, cipher aead.Cipher, encoded string, data AdditionalData) (*Flow, error) {
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
