package v1

import (
	"context"
	"database/sql"
	stderr "errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/internal/errors"

	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/internal/jwk"
	"github.com/tuanta7/hydros/internal/session"

	"github.com/tuanta7/hydros/pkg/aead"
	"github.com/tuanta7/hydros/pkg/mapx"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

const (
	CookieAuthenticationSIDName = "sid"
)

type OAuthHandler struct {
	cfg       *config.Config
	aead      aead.Cipher
	store     *sessions.CookieStore
	oauth2    core.OAuth2Provider
	jwkUC     *jwk.UseCase
	sessionUC session.UseCase
	flowUC    *flow.UseCase
	logger    *zapx.Logger
}

func NewOAuthHandler(
	cfg *config.Config,
	aead aead.Cipher,
	store *sessions.CookieStore,
	oauth2 core.OAuth2Provider,
	jwkUC *jwk.UseCase,
	sessionUC session.UseCase,
	flowUC *flow.UseCase,
	logger *zapx.Logger,
) *OAuthHandler {
	return &OAuthHandler{
		cfg:       cfg,
		aead:      aead,
		store:     store,
		oauth2:    oauth2,
		jwkUC:     jwkUC,
		sessionUC: sessionUC,
		flowUC:    flowUC,
		logger:    logger,
	}
}

func (h *OAuthHandler) HandleAuthorizeRequest(c *gin.Context) {
	ctx := c.Request.Context()
	ar, err := h.oauth2.NewAuthorizeRequest(ctx, c.Request)
	if err != nil {
		h.writeAuthorizeError(c, ar, err)
		return
	}

	f, err := h.handleLogin(ctx, c.Writer, c.Request, ar)
	if stderr.Is(err, errors.ErrAbortOAuth2Request) {
		return
	} else if err != nil {
		h.writeAuthorizeError(c, ar, err)
		return
	}

	//_, err = h.handleConsent(ctx, c.Writer, c.Request, ar, flow)
	//if errors.Is(err, domain.ErrAbortOAuth2Request) {
	//	return
	//} else if err != nil {
	//	return
	//}

	authorizeResponse, err := h.oauth2.NewAuthorizeResponse(ctx, ar, &session.Session{
		IDTokenSession: &oidc.IDTokenSession{
			Subject: f.Subject, // id of authenticated user
		},
	})
	if err != nil {
		h.writeAuthorizeError(c, ar, err)
		return
	}

	h.oauth2.WriteAuthorizeResponse(ctx, c.Writer, ar, authorizeResponse)
}

func (h *OAuthHandler) writeAuthorizeError(c *gin.Context, req *core.AuthorizeRequest, err error) {
	if req.IsRedirectURIValid() {
		h.oauth2.WriteAuthorizeError(c.Request.Context(), c.Writer, req, err)
		return
	}

	rfcErr := core.ErrorToRFC6749Error(err)
	c.HTML(http.StatusBadRequest, "errors.html", gin.H{
		"Error":       rfcErr.ErrorField,
		"Description": rfcErr.DescriptionField,
		"Debug":       rfcErr.DebugField,
		"Hint":        rfcErr.HintField,
	})
}

func (h *OAuthHandler) handleLogin(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	req *core.AuthorizeRequest,
) (*flow.Flow, error) {
	loginVerifier := strings.TrimSpace(req.Form.Get("login_verifier"))
	if loginVerifier == "" {
		return nil, h.requestLogin(ctx, w, r, req)
	}

	return h.verifyLogin(ctx, w, r, req, loginVerifier)
}

func (h *OAuthHandler) checkSession(ctx context.Context, r *http.Request) (*session.LoginSession, error) {
	cookie, err := h.store.Get(r, h.cfg.SessionCookieName())
	if err != nil {
		h.logger.Error("cookie store returned an error.",
			zap.Error(err),
			zap.String("method", "store.Get"),
		)
		return nil, errors.ErrNoAuthenticationSessionFound
	}

	sid := mapx.GetStringDefault(cookie.Values, CookieAuthenticationSIDName, "")
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

	return loginSession, nil
}

func (h *OAuthHandler) forwardLoginRequest(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
	loginSession *session.LoginSession,
) error {
	sessionID := uuid.NewString()
	subject := ""
	authenticatedAt := sql.NullTime{}

	if loginSession != nil {
		sessionID = loginSession.ID
		subject = loginSession.Subject
		authenticatedAt = loginSession.AuthenticatedAt

		if subject == "" || authenticatedAt.Time.IsZero() {
			return core.ErrServerError.WithHint("subject and authenticated_at must be set together.")
		}
	}

	skip := false
	if subject != "" {
		skip = true
	}

	if ar.Prompt.IncludeAll("none") && !skip {
		return core.ErrLoginRequired.WithHint("Prompt 'none' was requested, but no existing login loginSession was found.")
	}

	loginVerifier := x.RandomUUID()
	loginChallenge := x.RandomUUID()
	loginCSRF := x.RandomUUID()

	cl := client.SanitizedClientFromRequest(ar)
	f := &flow.Flow{
		LoginChallenge:       loginChallenge,
		LoginVerifier:        loginVerifier,
		LoginCSRF:            loginCSRF,
		LoginSkip:            skip,
		LoginWasHandled:      false,
		LoginAuthenticatedAt: authenticatedAt,
		RequestedAt:          x.NowUTC().Truncate(time.Second),
		RequestURL:           r.URL.String(), // TODO: get proper authorize request url
		RequestedScope:       []string(ar.Scope),
		RequestedAudience:    []string(ar.Audience),
		Client:               cl,
		ClientID:             cl.GetID(),
		Subject:              subject,
		LoginSessionID: sql.NullString{
			Valid:  len(sessionID) > 0,
			String: sessionID,
		},
		State:                   flow.FlowStateLoginInitialized,
		ForcedSubjectIdentifier: "",
		Context:                 []byte("{}"),
		OIDCContext:             []byte("{}"),
	}

	// TODO: prevent csrf

	encodedFlow, err := f.EncodeToLoginChallenge(ctx, h.aead)
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

	if ar.MaxAge > -1 && loginSession.AuthenticatedAt.Time.UTC().Add(time.Duration(ar.MaxAge)*time.Second).Before(x.NowUTC()) {
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
	ar *core.AuthorizeRequest,
	verifier string,
) (*flow.Flow, error) {
	return nil, nil
}

func (h *OAuthHandler) handleConsent(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	req *core.AuthorizeRequest,
	flow *flow.Flow,
) (*flow.Flow, error) {
	consentVerifier := strings.TrimSpace(req.Form.Get("consent_verifier"))
	if consentVerifier == "" {
		return nil, h.requestConsent(ctx, w, r, req, flow)
	}

	return h.verifyConsent()
}

func (h *OAuthHandler) requestConsent(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
	flow *flow.Flow,
) error {
	return nil
}

func (h *OAuthHandler) verifyConsent() (*flow.Flow, error) {
	return nil, nil
}
