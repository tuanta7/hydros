package dbtype

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type NullDuration struct {
	Duration time.Duration
	Valid    bool
}

func (ns *NullDuration) Scan(value interface{}) error {
	d := sql.NullInt64{}
	if err := d.Scan(value); err != nil {
		return err
	}

	ns.Duration = time.Duration(d.Int64)
	ns.Valid = d.Valid
	return nil
}

func (ns *NullDuration) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return int64(ns.Duration), nil
}

func (ns *NullDuration) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(ns.Duration.String())
}

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
