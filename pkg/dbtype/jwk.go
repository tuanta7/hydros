package dbtype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/tuanta7/hydros/core"
)

type JWKSet struct {
	*core.JSONWebKeySet
}

func (j *JWKSet) Scan(value any) error {
	v := fmt.Sprintf("%s", value)
	if len(v) == 0 {
		return nil
	}

	return json.Unmarshal([]byte(v), j)
}

func (j *JWKSet) Value() (driver.Value, error) {
	value, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(value), nil
}
