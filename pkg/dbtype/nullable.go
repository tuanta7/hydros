package dbtype

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type NullTime time.Time

func (ns *NullTime) Scan(value interface{}) error {
	var v sql.NullTime
	if err := (&v).Scan(value); err != nil {
		return err
	}
	*ns = NullTime(v.Time)
	return nil
}

func (ns NullTime) MarshalJSON() ([]byte, error) {
	var t *time.Time
	if !time.Time(ns).IsZero() {
		tt := time.Time(ns)
		t = &tt
	}
	return json.Marshal(t)
}

func (ns *NullTime) UnmarshalJSON(data []byte) error {
	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	*ns = NullTime(t)
	return nil
}

func (ns NullTime) Value() (driver.Value, error) {
	return sql.NullTime{Valid: !time.Time(ns).IsZero(), Time: time.Time(ns)}.Value()
}

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

func (ns NullDuration) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return int64(ns.Duration), nil
}

func (ns NullDuration) MarshalJSON() ([]byte, error) {
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

type NullString string

func (ns NullString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ns))
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	if ns == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, (*string)(ns))
}

func (ns *NullString) Scan(value interface{}) error {
	var v sql.NullString
	if err := (&v).Scan(value); err != nil {
		return err
	}
	*ns = NullString(v.String)
	return nil
}

func (ns NullString) Value() (driver.Value, error) {
	if len(ns) == 0 {
		return sql.NullString{}.Value()
	}
	return sql.NullString{Valid: true, String: string(ns)}.Value()
}

func (ns NullString) String() string {
	return string(ns)
}
