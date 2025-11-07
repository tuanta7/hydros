package flow

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tuanta7/hydros/core"
)

const (
	LoginRequestDeniedErrorName   = "login request denied"
	ConsentRequestDeniedErrorName = "consent request denied"
)

type RequestDeniedError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Hint             string `json:"error_hint"`
	Code             int    `json:"status_code"`
	Debug            string `json:"error_debug"`
	Valid            bool   `json:"valid"`
}

func (e *RequestDeniedError) IsError() bool {
	return e != nil && e.Valid
}

func (e *RequestDeniedError) SetDefaults(name string) {
	if e.Error == "" {
		e.Error = name
	}

	if e.Code == 0 {
		e.Code = http.StatusBadRequest
	}
}

func (e *RequestDeniedError) ToRFCError() *core.RFC6749Error {
	if e.Error == "" {
		e.Error = "request_denied"
	}

	if e.Code == 0 {
		e.Code = core.ErrInvalidRequest.CodeField
	}

	return &core.RFC6749Error{
		ErrorField:       e.Error,
		DescriptionField: e.ErrorDescription,
		HintField:        e.Hint,
		CodeField:        e.Code,
		DebugField:       e.Debug,
	}
}

func (e *RequestDeniedError) Scan(value any) error {
	v := fmt.Sprintf("%s", value)
	if len(v) == 0 || v == "{}" {
		return nil
	}

	if err := json.Unmarshal([]byte(v), e); err != nil {
		return err
	}

	e.Valid = true
	return nil
}

func (e *RequestDeniedError) Value() (driver.Value, error) {
	if !e.IsError() {
		return "{}", nil
	}

	value, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return string(value), nil
}
