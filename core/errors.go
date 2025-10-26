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
	ErrInvalidGrant = &RFC6749Error{
		ErrorField:       "invalid_grant",
		DescriptionField: "The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client.",
		CodeField:        http.StatusBadRequest,
	}
	ErrUnauthorizedClient = &RFC6749Error{
		ErrorField:       "unauthorized_client",
		DescriptionField: "The client is not authorized to request a token using this method.",
		HintField:        "Make sure that client id and secret are correctly specified and that the client exists.",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidTokenFormat = &RFC6749Error{
		ErrorField:       "invalid_token",
		DescriptionField: "Invalid token format.",
		HintField:        "Check that you provided a valid token in the right format.",
		CodeField:        http.StatusBadRequest,
	}
	ErrTokenSignatureMismatch = &RFC6749Error{
		ErrorField:       "token_signature_mismatch",
		DescriptionField: "Token signature mismatch.",
		HintField:        "Check that you provided  a valid token in the right format.",
		CodeField:        http.StatusBadRequest,
	}
	ErrNotFound = &RFC6749Error{
		ErrorField:       "not_found",
		DescriptionField: "Could not find the requested resource(s).",
		CodeField:        http.StatusNotFound,
	}
	ErrInactiveToken = &RFC6749Error{
		ErrorField:       "token_inactive",
		DescriptionField: "Token is inactive because it is malformed, expired or otherwise invalid.",
		HintField:        "Token validation failed.",
		CodeField:        http.StatusUnauthorized,
	}
)
