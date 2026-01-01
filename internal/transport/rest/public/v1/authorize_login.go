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
	"github.com/tuanta7/hydros/pkg/helper/mapx"
	"go.uber.org/zap"
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

func (h *OAuthHandler) checkSession(ctx context.Context, r *http.Request) (*session.LoginSession, error) {
	cookie, err := h.store.Get(r, h.cfg.SessionCookieKey())
	if err != nil {
		h.logger.Error("cookie store returned an error.",
			zap.Error(err),
			zap.String("method", "store.Get"),
		)
		return nil, errors.ErrNoAuthenticationSessionFound
	}

	sid := mapx.GetStringDefault(cookie.Values, CookieLoginSessionIDKey, "")
	if sid == "" {
		h.logger.Debug("cookie exists but session value is empty.", zap.String("method", "cookie.Values"))
		return nil, errors.ErrNoAuthenticationSessionFound
	}

	loginSession, err := h.sessionUC.GetRememberedLoginSession(ctx, nil, sid)
	if stderr.Is(err, errors.ErrNotFound) {
		h.logger.Debug("cookie exists and session value exist but are not remembered any more.",
			zap.Error(err),
			zap.String("method", "sessionUC.GetRememberedLoginSession"),
		)
		return nil, errors.ErrNoAuthenticationSessionFound
	} else if err != nil {
		return nil, err
	}

	subject := loginSession.Subject
	authenticatedAt := time.Time(loginSession.AuthenticatedAt)

	if (subject == "" && authenticatedAt.IsZero()) || (subject != "" && authenticatedAt.IsZero()) {
		return nil, core.ErrServerError.WithHint("Subject and authenticated_at must be set together.")
	}

	return loginSession, nil
}

func (h *OAuthHandler) forwardLoginRequest(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
	loginSession *session.LoginSession,
) error {
	if loginSession == nil {
		loginSession = session.NewLoginSession()
	}

	// if both subject and authenticated_at are set, we can skip the login
	skip := loginSession.Subject != "" && !time.Time(loginSession.AuthenticatedAt).IsZero()

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
		LoginAuthenticatedAt:    loginSession.AuthenticatedAt,
		RequestedAt:             x.NowUTC().Truncate(time.Second),
		RequestURL:              r.URL.String(), // TODO: get proper authorize request url when behind a reverse proxy
		RequestedScope:          []string(ar.RequestedScope),
		RequestedAudience:       []string(ar.RequestedAudience),
		Client:                  cl,
		ClientID:                cl.GetID(),
		Subject:                 loginSession.Subject,
		LoginSessionID:          dbtype.NullString(loginSession.ID),
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

	if err = f.InvalidateLoginRequest(); err != nil {
		return nil, core.ErrInvalidRequest.WithDebug(err.Error())
	}

	if f.LoginError.IsError() {
		f.LoginError.SetDefaults(flow.LoginRequestDeniedErrorName)
		return nil, f.LoginError.ToRFCError()
	}

	if err = session.ValidateCSRFSession(r, h.store, session.LoginCSRFCookieKey, f.LoginCSRF); err != nil {
		return nil, err
	}

	if f.LoginSkip && !f.LoginRemember {
		return nil, core.ErrServerError.WithHint("The login request was previously remembered and can only be forgotten using the reject feature.")
	}

	sessionID := f.LoginSessionID.String()

	// if the login session is not skipped, update or save the login session
	if !f.LoginSkip {
		if time.Time(f.LoginAuthenticatedAt).IsZero() {
			return nil, core.ErrServerError.WithHint("Expected the handled login request to contain a valid authenticated_at value but it was zero.")
		}

		err = h.sessionUC.ConfirmLoginSession(ctx, &session.LoginSession{
			ID:                        sessionID,
			Subject:                   f.Subject,
			AuthenticatedAt:           f.LoginAuthenticatedAt,
			IdentityProviderSessionID: f.IdentityProviderSessionID,
			Remember:                  f.LoginRemember,
		})
		if err != nil {
			return nil, err
		}
	}

	if (f.LoginSkip && !f.LoginExtendSessionLifetime) || !f.LoginRemember {
		// if the login was skipped, and the session should not be extended,
		// or if the user does not want to remember the login session,
		// skip the cookie setting
		return f, nil
	}

	cookie, _ := h.store.Get(r, h.cfg.SessionCookieKey())
	cookie.Values[CookieLoginSessionIDKey] = sessionID
	if f.LoginRememberFor > 0 {
		cookie.Options.MaxAge = f.LoginRememberFor
	}
	cookie.Options.HttpOnly = true
	cookie.Options.Path = h.cfg.SessionCookiePath()
	cookie.Options.SameSite = h.cfg.SessionCookieSameSiteMode()
	cookie.Options.Secure = h.cfg.SessionCookieSecure()
	if err = cookie.Save(r, w); err != nil {
		return nil, err
	}

	return f, nil
}
