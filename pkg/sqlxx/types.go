package sqlxx

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tidwall/gjson"
)

type StringSliceJSONFormat []string

// Scan implements the Scanner interface.
func (m *StringSliceJSONFormat) Scan(value interface{}) error {
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
func (m StringSliceJSONFormat) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "[]", nil
	}

	encoded, err := json.Marshal(&m)
	return string(encoded), errors.Join(err)
}

type NullDuration struct {
	Duration time.Duration
	Valid    bool
}

// Scan implements the Scanner interface.
func (ns *NullDuration) Scan(value interface{}) error {
	d := sql.NullInt64{}
	if err := d.Scan(value); err != nil {
		return err
	}

	ns.Duration = time.Duration(d.Int64)
	ns.Valid = d.Valid
	return nil
}

// Value implements the driver Valuer interface.
func (ns NullDuration) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return int64(ns.Duration), nil
}

// MarshalJSON returns m as the JSON encoding of m.
func (ns NullDuration) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(ns.Duration.String())
}

// UnmarshalJSON sets *m to a copy of data.
func (ns *NullDuration) UnmarshalJSON(data []byte) error {
	if ns == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}

	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	p, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	ns.Duration = p
	ns.Valid = true
	return nil
}
