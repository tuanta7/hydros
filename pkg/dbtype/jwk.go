package dbtype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/go-jose/go-jose/v4"
)

// JWKSet or JKWs is a JSON array containing the public keys a server
// uses to verify the signature on JWTs it receives.
type JWKSet struct {
	*jose.JSONWebKeySet
}

func (j *JWKSet) Scan(value any) error {
	v := fmt.Sprintf("%s", value)
	if len(v) == 0 {
		return nil
	}

	return json.Unmarshal([]byte(v), j)
}

func (j JWKSet) Value() (driver.Value, error) {
	value, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(value), nil
}
