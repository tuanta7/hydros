package domain

import "errors"

const (
	CookieAuthenticationSIDName = "sid"
)

var (
	ErrAbortOAuth2Request             = errors.New("the OAuth 2.0 Authorization request must be aborted")
	ErrNoPreviousConsentFound         = errors.New("no previous OAuth 2.0 Consent could be found for this access request")
	ErrNoAuthenticationSessionFound   = errors.New("no previous login session was found")
	ErrHintDoesNotMatchAuthentication = errors.New("subject from hint does not match subject from session")
)
