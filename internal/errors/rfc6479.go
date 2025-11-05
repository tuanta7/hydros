package errors

import (
	"net/http"

	"github.com/tuanta7/hydros/core"
)

var (
	ErrNotFound = &core.RFC6749Error{
		CodeField:        http.StatusNotFound,
		ErrorField:       http.StatusText(http.StatusNotFound),
		DescriptionField: "Unable to locate the requested resource",
	}
	ErrConflict = &core.RFC6749Error{
		CodeField:        http.StatusConflict,
		ErrorField:       http.StatusText(http.StatusConflict),
		DescriptionField: "Unable to process the requested resource because of conflict in the current state",
	}
	ErrUnsupportedKeyAlgorithm = &core.RFC6749Error{
		CodeField:        http.StatusBadRequest,
		ErrorField:       http.StatusText(http.StatusBadRequest),
		DescriptionField: "Unsupported key algorithm",
	}
	ErrUnsupportedEllipticCurve = &core.RFC6749Error{
		CodeField:        http.StatusBadRequest,
		ErrorField:       http.StatusText(http.StatusBadRequest),
		DescriptionField: "Unsupported elliptic curve",
	}
	ErrMinimalRsaKeyLength = &core.RFC6749Error{
		CodeField:        http.StatusBadRequest,
		ErrorField:       http.StatusText(http.StatusBadRequest),
		DescriptionField: "Unsupported RSA key length",
	}
	ErrInvalidClientMetadata = &core.RFC6749Error{
		DescriptionField: "The value of one of the Client Metadata fields is invalid and the server has rejected this request. Note that an Authorization Server MAY choose to substitute a valid value for any requested parameter of a Client's Metadata.",
		ErrorField:       "invalid_client_metadata",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidRedirectURI = &core.RFC6749Error{
		DescriptionField: "The value of one or more redirect_uris is invalid.",
		ErrorField:       "invalid_redirect_uri",
		CodeField:        http.StatusBadRequest,
	}
	ErrInvalidRequest = &core.RFC6749Error{
		DescriptionField: "The request is missing a required parameter, includes an unsupported parameter value (other than grant type), repeats a parameter, includes multiple credentials, utilizes more than one mechanism for authenticating the client, or is otherwise malformed.",
		ErrorField:       "invalid_request",
		CodeField:        http.StatusBadRequest,
	}
)
