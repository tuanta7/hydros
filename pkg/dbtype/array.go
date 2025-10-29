package dbtype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
)

type StringArray []string

// Scan implements the Scanner interface.
func (m *StringArray) Scan(value interface{}) error {
	val := fmt.Sprintf("%s", value)
	if len(val) == 0 {
		val = "[]"
	}

	if parsed := gjson.Parse(val); parsed.Type == gjson.Null {
		val = "[]"
	} else if !parsed.IsArray() {
		return fmt.Errorf("expected JSON value to be an array but got type: %s", parsed.Type.String())
	}

	return errors.Join(json.Unmarshal([]byte(val), &m))
}

// Value implements the driver Valuer interface.
func (m StringArray) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "[]", nil
	}

	encoded, err := json.Marshal(&m)
	return string(encoded), errors.Join(err)
}
