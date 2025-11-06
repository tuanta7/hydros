package session

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

type CookieConfigurator interface {
	CookieSessionName() string
	CookieDomain() string
	CookiePath() string
	CookieSameSiteMode() http.SameSite
	CookieSecure() bool
	CookieKeyPairs() [][]byte
}

func NewCookieStore(cfg CookieConfigurator) *sessions.CookieStore {
	cookieStore := sessions.NewCookieStore(cfg.CookieKeyPairs()...)
	cookieStore.MaxAge(0)
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.Secure = cfg.CookieSecure()

	if d := cfg.CookieDomain(); d != "" {
		cookieStore.Options.Domain = d
	}

	if p := cfg.CookiePath(); p != "" {
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

	if err := session.Save(r, w); err != nil {
		return err
	}

	return nil
}
