package core

import (
	"net/http"
)

type RFC6749Error struct {
	ErrorField       string
	DescriptionField string
	HintField        string
	CodeField        int
}

func (e RFC6749Error) Error() string {
	return e.ErrorField
}

const (
	errUnknownErrorName = "error"
)

var (
	ErrUnknownRequest = &RFC6749Error{
		ErrorField:       errUnknownErrorName,
		DescriptionField: "The handler is not responsible for this request.",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidScope = &RFC6749Error{
		ErrorField:       "invalid_scope",
		DescriptionField: "The requested scope is invalid, unknown, or malformed.",
		CodeField:        http.StatusBadRequest,
	}
)
