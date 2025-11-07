package session

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/pkg/mapx"
)

const (
	CSRFKey = "csrf_token"
)

type CookieConfigurator interface {
	SessionCookieKey() string
	SessionCookieDomain() string
	SessionCookiePath() string
	SessionCookieSameSiteMode() http.SameSite
	SessionCookieSecure() bool
	CookieKeyPairs() [][]byte
}

func NewCookieStore(cfg CookieConfigurator) *sessions.CookieStore {
	cookieStore := sessions.NewCookieStore(cfg.CookieKeyPairs()...)
	cookieStore.MaxAge(0)
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.Secure = cfg.SessionCookieSecure()

	if d := cfg.SessionCookieDomain(); d != "" {
		cookieStore.Options.Domain = d
	}

	if p := cfg.SessionCookiePath(); p != "" {
		cookieStore.Options.Path = p
	}

	return cookieStore
}

func CreateCSRFSession(
	w http.ResponseWriter, r *http.Request,
	cfg CookieConfigurator,
	store *sessions.CookieStore,
	name, value string,
	maxAge time.Duration,
) error {
	// return a new session on error so we can ignore errors
	session, _ := store.Get(r, name)

	session.Values["csrf"] = value
	session.Options.HttpOnly = true
	session.Options.Secure = cfg.SessionCookieSecure()
	session.Options.SameSite = cfg.SessionCookieSameSiteMode()
	session.Options.Domain = cfg.SessionCookieDomain()
	session.Options.MaxAge = int(maxAge.Seconds())

	if err := session.Save(r, w); err != nil {
		return err
	}

	return nil
}

func ValidateCSRFSession(r *http.Request, store *sessions.CookieStore, name, expectedCSRF string) error {
	cookie, err := store.Get(r, name)
	if err != nil {
		return core.ErrRequestForbidden.WithHint("CSRF session cookie could not be decoded.")
	}

	csrf, err := mapx.GetString(cookie.Values, "csrf")
	if err != nil {
		return core.ErrRequestForbidden.WithHint("No CSRF value available in the session cookie.")
	}

	if csrf != expectedCSRF {
		return core.ErrRequestForbidden.WithHint("The CSRF value from the token does not match the CSRF value from the data store.")
	}

	return nil
}
