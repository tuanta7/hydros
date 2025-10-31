package domain

import (
	"time"
)

type Set string
type Algorithm string

const (
	IDTokenSet     Set = "id-token"
	AccessTokenSet Set = "access-token"

	AlgorithmHS256 Algorithm = "HS256"
	AlgorithmHS512 Algorithm = "HS512"
	AlgorithmRS256 Algorithm = "RS256"
	AlgorithmRS512 Algorithm = "RS512"

	// AlgorithmES256 and AlgorithmES512 are not supported yet
	AlgorithmES256 Algorithm = "ES256"
	AlgorithmES512 Algorithm = "ES512"
)

var (
	KeySizeBytes = map[Algorithm]int{
		AlgorithmRS256: 256,
		AlgorithmRS512: 512,
		AlgorithmHS256: 32,
		AlgorithmHS512: 64,
	}
)

type JSONWebKey struct {
	KeyID     string    `json:"kid" db:"kid"`
	Set       Set       `json:"set" db:"set"`
	Key       string    `json:"key" db:"key"`
	KeyType   string    `json:"kty" db:"-"`
	Algorithm Algorithm `json:"algorithm" db:"algorithm"`
	Use       string    `json:"use" db:"use"`
	Active    bool      `json:"active" db:"active"` // only one key of each set can be active at a time
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (dj JSONWebKey) ColumnMap() map[string]any {
	return map[string]any{
		"kid":        dj.KeyID,
		"set":        dj.Set,
		"key":        dj.Key,
		"algorithm":  dj.Algorithm,
		"use":        dj.Use,
		"active":     dj.Active,
		"created_at": dj.CreatedAt,
	}
}
