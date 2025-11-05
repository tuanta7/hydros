package jwk

import (
	"time"
)

type Set string

const (
	IDTokenSet     Set = "id-token"
	AccessTokenSet Set = "access-token"
)

// KeyData is used to store private/secret keys in the database.
// It is the direct replacement of SQLData in ory/hydra
type KeyData struct {
	KeyID     string    `json:"kid" db:"kid"`
	SetID     Set       `json:"sid" db:"sid"`
	Key       string    `json:"key" db:"key"`       // encrypted marshalled jose.JSONWebKey
	Active    bool      `json:"active" db:"active"` // only one key of each set can be active at a time
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (dj KeyData) ColumnMap() map[string]any {
	return map[string]any{
		"kid":        dj.KeyID,
		"sid":        dj.SetID,
		"key":        dj.Key,
		"active":     dj.Active,
		"created_at": dj.CreatedAt,
	}
}
