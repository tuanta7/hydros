package v1

import (
	stderr "errors"
	"net/http"

	"github.com/tuanta7/hydros/internal/errors"
	"github.com/tuanta7/hydros/pkg/helper/mapx"
)

func (h *OAuthHandler) revokeLoginSession(w http.ResponseWriter, r *http.Request) error {
	sid, err := h.revokeAuthenticationCookie(w, r)
	if err != nil {
		return err
	} else if sid == "" {
		return nil
	}

	_, err = h.sessionUC.DeleteLoginSession(r.Context(), sid)
	if stderr.Is(err, errors.ErrNotFound) {
		return nil
	}

	return err
}

func (h *OAuthHandler) revokeAuthenticationCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, _ := h.store.Get(r, h.cfg.SessionCookieKey())
	sid, _ := mapx.GetString(cookie.Values, CookieLoginSessionIDKey)

	cookie.Values[CookieLoginSessionIDKey] = ""
	cookie.Options.HttpOnly = true
	cookie.Options.Path = h.cfg.SessionCookiePath()
	cookie.Options.SameSite = h.cfg.SessionCookieSameSiteMode()
	cookie.Options.Secure = h.cfg.SessionCookieSecure()
	cookie.Options.Domain = h.cfg.SessionCookieDomain()
	cookie.Options.MaxAge = -1

	if err := cookie.Save(r, w); err != nil {
		return "", err
	}

	return sid, nil
}
