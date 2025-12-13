package v1

import (
	"context"
	stderr "errors"
	"net/http"
	"net/url"
	"time"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/internal/errors"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/internal/session"
	"github.com/tuanta7/hydros/pkg/dbtype"
	"github.com/tuanta7/hydros/pkg/mapx"
)

func (h *OAuthHandler) requestLogin(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
) error {
	if ar.Prompt.IncludeAll("login") {
		return h.forwardLoginRequest(ctx, w, r, ar, nil)
	}

	loginSession, err := h.checkSession(ctx, r)
	if stderr.Is(err, errors.ErrNoAuthenticationSessionFound) {
		return h.forwardLoginRequest(ctx, w, r, ar, nil)
	} else if err != nil {
		return err
	}

	if ar.MaxAge > -1 && time.Time(loginSession.AuthenticatedAt).UTC().Add(time.Duration(ar.MaxAge)*time.Second).Before(x.NowUTC()) {
		if ar.Prompt.IncludeAll("none") {
			return core.ErrLoginRequired.WithHint("Request failed because prompt is set to 'none' and authentication time reached 'max_age'.")
		}
		return h.forwardLoginRequest(ctx, w, r, ar, loginSession)
	}

	// TODO: support token hint
	return h.forwardLoginRequest(ctx, w, r, ar, loginSession)
}

func (h *OAuthHandler) verifyLogin(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	verifier string,
) (*flow.Flow, error) {
	f, err := h.flowUC.DecodeFlow(ctx, verifier, flow.AsLoginVerifier)
	if err != nil {
		return nil, err
	}

	err = f.InvalidateLoginRequest()
	if err != nil {
		return nil, core.ErrInvalidRequest.WithDebug(err.Error())
	}

	if f.LoginError.IsError() {
		f.LoginError.SetDefaults(flow.LoginRequestDeniedErrorName)
		return nil, f.LoginError.ToRFCError()
	}

	err = session.ValidateCSRFSession(r, h.store, session.LoginCSRFCookieKey, f.LoginCSRF)
	if err != nil {
		return nil, err
	}

	if f.LoginSkip && !f.LoginRemember {
		return nil, core.ErrServerError.WithHint("The login request was previously remembered and can only be forgotten using the reject feature.")
	}

	sessionID := f.LoginSessionID.String()

	if !f.LoginSkip && f.LoginRemember {
		if time.Time(f.LoginAuthenticatedAt).IsZero() {
			return nil, core.ErrServerError.WithHint("Expected the handled login request to contain a valid authenticated_at value but it was zero.")
		}

		err = h.sessionUC.ConfirmLoginSession(ctx, &session.LoginSession{
			ID:                        sessionID,
			Subject:                   f.Subject,
			AuthenticatedAt:           f.LoginAuthenticatedAt,
			IdentityProviderSessionID: f.IdentityProviderSessionID,
			Remember:                  true,
		})
		if err != nil {
			return nil, err
		}
	}

	if !f.LoginSkip && !f.LoginRemember {
		// if the user logged in, then does not want to remember the login session
		err = h.revokeLoginSession(w, r)
		if err != nil {
			return nil, err
		}
	}

	if (f.LoginSkip && !f.LoginExtendSessionLifetime) || !f.LoginRemember {
		// if the login was skipped, and the session should not be extended, or if the user does not want to remember
		// the login session, we can skip the cookie setting
		return f, nil
	}

	cookie, _ := h.store.Get(r, h.cfg.SessionCookieKey())
	cookie.Values[CookieLoginSessionIDKey] = sessionID
	cookie.Options.HttpOnly = true
	cookie.Options.Path = h.cfg.SessionCookiePath()
	cookie.Options.SameSite = h.cfg.SessionCookieSameSiteMode()
	cookie.Options.Secure = h.cfg.SessionCookieSecure()
	if f.LoginRememberFor > 0 {
		cookie.Options.MaxAge = f.LoginRememberFor
	}
	if err := cookie.Save(r, w); err != nil {
		return nil, err
	}

	return f, nil
}

func (h *OAuthHandler) forwardLoginRequest(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
	loginSession *session.LoginSession,
) error {
	sessionID := ""
	subject := ""
	authenticatedAt := time.Time{}

	if loginSession != nil {
		sessionID = loginSession.ID
		subject = loginSession.Subject
		authenticatedAt = time.Time(loginSession.AuthenticatedAt)

		if (subject == "" && authenticatedAt.IsZero()) || (subject != "" && authenticatedAt.IsZero()) {
			return core.ErrServerError.WithHint("Subject and authenticated_at must be set together.")
		}
	}

	skip := false

	// if both subject and authenticated_at are set, we can skip the login
	if subject != "" {
		skip = true
	}

	// if both are empty, we have to enforce the login
	if !skip && ar.Prompt.IncludeAll("none") {
		return core.ErrLoginRequired.WithHint("Prompt 'none' was requested, but no existing login loginSession was found.")
	}

	csrf := x.RandomUUID()

	cl := client.SanitizedClientFromRequest(ar)
	f := &flow.Flow{
		ID:                      x.RandomUUID(),
		LoginCSRF:               csrf,
		LoginSkip:               skip,
		LoginWasHandled:         false,
		LoginAuthenticatedAt:    dbtype.NullTime(authenticatedAt),
		RequestedAt:             x.NowUTC().Truncate(time.Second),
		RequestURL:              r.URL.String(), // TODO: get proper authorize request url when behind a reverse proxy
		RequestedScope:          []string(ar.RequestedScope),
		RequestedAudience:       []string(ar.RequestedAudience),
		Client:                  cl,
		ClientID:                cl.GetID(),
		Subject:                 subject,
		LoginSessionID:          dbtype.NullString(sessionID),
		State:                   flow.StateLoginInitialized,
		ForcedSubjectIdentifier: "",
		Context:                 []byte("{}"),
		OIDCContext:             []byte("{}"),
	}

	err := session.CreateCSRFSession(w, r, h.cfg, h.store, session.LoginCSRFCookieKey, csrf, h.cfg.GetConsentRequestMaxAge())
	if err != nil {
		return err
	}

	encodedFlow, err := h.flowUC.EncodeFlow(ctx, f, flow.AsLoginChallenge)
	if err != nil {
		return err
	}

	redirectTo := h.cfg.GetLoginPageURL()
	if ar.Prompt.IncludeAll("registration") {
		redirectTo = h.cfg.GetRegistrationURL()
	}

	params := url.Values{}
	params.Set("login_challenge", encodedFlow)
	redirectTo.RawQuery = params.Encode()

	http.Redirect(w, r, redirectTo.String(), http.StatusFound)
	return errors.ErrAbortOAuth2Request
}

func (h *OAuthHandler) revokeLoginSession(w http.ResponseWriter, r *http.Request) error {
	sid, err := h.revokeAuthenticationCookie(w, r)
	if err != nil {
		return err
	}

	if sid == "" {
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
