package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tuanta7/hydros/core/x"
)

type RFC6749Error struct {
	ErrorField       string `json:"error"`
	DescriptionField string `json:"error_description"`
	HintField        string `json:"hint,omitempty"`
	CodeField        int    `json:"-"`
	DebugField       string `json:"-"`
	cause            error
}

func (e *RFC6749Error) Error() string {
	return e.ErrorField
}

func (e *RFC6749Error) WithHint(hint string, args ...interface{}) *RFC6749Error {
	err := *e
	err.HintField = fmt.Sprintf(hint, args...)
	return &err
}

func (e *RFC6749Error) WithWrap(cause error) *RFC6749Error {
	e.cause = cause
	return e
}

func (e *RFC6749Error) WithDebug(debug string) *RFC6749Error {
	err := *e
	err.DebugField = debug
	return &err
}

const (
	errUnknownErrorName = "error"
)

var (
	ErrServerError = &RFC6749Error{
		ErrorField:       "server_error",
		DescriptionField: "The authorization server encountered an unexpected condition that prevented it from fulfilling the request.",
		CodeField:        http.StatusInternalServerError,
	}
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
	ErrTokenExpired = &RFC6749Error{
		ErrorField:       "invalid_token",
		DescriptionField: "Token expired.",
		HintField:        "The token expired.",
		CodeField:        http.StatusUnauthorized,
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
	ErrInvalidRequest = &RFC6749Error{
		ErrorField:       "invalid_request",
		DescriptionField: "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed.",
		HintField:        "Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified.",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidClient = &RFC6749Error{
		ErrorField:       "invalid_client",
		DescriptionField: "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method).",
		CodeField:        http.StatusUnauthorized,
	}
	ErrRequestUnauthorized = &RFC6749Error{
		ErrorField:       "request_unauthorized",
		DescriptionField: "The request could not be authorized.",
		HintField:        "Check that you provided valid credentials in the right format.",
		CodeField:        http.StatusUnauthorized,
	}
)

func ErrorToRFC6749Error(err error) *RFC6749Error {
	var e *RFC6749Error
	if errors.As(err, &e) {
		return e
	}
	return &RFC6749Error{
		ErrorField:       errUnknownErrorName,
		DescriptionField: "The error is unrecognizable",
		DebugField:       err.Error(),
		CodeField:        http.StatusInternalServerError,
		cause:            err,
	}
}

func (o *OAuth2) writeError(ctx context.Context, rw http.ResponseWriter, err error) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	rfcErr := ErrorToRFC6749Error(err)

	jsonErr, err := json.Marshal(rfcErr)
	if err != nil {
		if o.config.IsDebugging() {
			errPayload := fmt.Sprintf(
				`{"error":"server_error","error_description":"%s"}`,
				x.EscapeJSONString(err.Error()),
			)
			http.Error(rw, errPayload, http.StatusInternalServerError)
		}
		http.Error(rw, `{"error":"server_error"}`, http.StatusInternalServerError)
	}

	rw.WriteHeader(rfcErr.CodeField)
	_, _ = rw.Write(jsonErr)
}
