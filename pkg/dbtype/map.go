package dbtype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// MapStringAny represents JSON in Go
type MapStringAny map[string]any

// Scan implements the Scanner interface.
func (n *MapStringAny) Scan(value interface{}) error {
	v := fmt.Sprintf("%s", value)
	if len(v) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(v), n)
}

func (n MapStringAny) Value() (driver.Value, error) {
	value, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	return string(value), nil
}
