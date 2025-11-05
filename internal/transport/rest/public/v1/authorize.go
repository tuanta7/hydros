package v1

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/internal/usecase/jwk"
	"github.com/tuanta7/hydros/internal/usecase/session"
	"github.com/tuanta7/hydros/pkg/mapx"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type OAuthHandler struct {
	cfg       *config.Config
	store     *sessions.CookieStore
	oauth2    core.OAuth2Provider
	jwkUC     *jwk.UseCase
	sessionUC session.UseCase
	logger    *zapx.Logger
}

func NewOAuthHandler(
	cfg *config.Config,
	store *sessions.CookieStore,
	oauth2 core.OAuth2Provider,
	jwkUC *jwk.UseCase,
	sessionUC session.UseCase,
	logger *zapx.Logger,
) *OAuthHandler {
	return &OAuthHandler{
		cfg:       cfg,
		store:     store,
		oauth2:    oauth2,
		jwkUC:     jwkUC,
		sessionUC: sessionUC,
		logger:    logger,
	}
}

func (h *OAuthHandler) HandleAuthorizeRequest(c *gin.Context) {
	ctx := c.Request.Context()
	authorizeRequest, err := h.oauth2.NewAuthorizeRequest(ctx, c.Request)
	if err != nil {
		h.writeAuthorizeError(c, authorizeRequest, err)
		return
	}

	flow, err := h.handleLogin(ctx, c.Writer, c.Request, authorizeRequest)
	if err != nil {
		h.writeAuthorizeError(c, authorizeRequest, err)
		return
	}

	_, err = h.handleConsent(ctx, c.Writer, c.Request, authorizeRequest, flow)
	if err != nil {
		return
	}

	authorizeResponse, err := h.oauth2.NewAuthorizeResponse(ctx, authorizeRequest, &domain.Session{
		IDTokenSession: &oidc.IDTokenSession{
			Subject: flow.Subject, // id of authenticated user
		},
	})
	if err != nil {
		h.writeAuthorizeError(c, authorizeRequest, err)
		return
	}

	h.oauth2.WriteAuthorizeResponse(ctx, c.Writer, authorizeRequest, authorizeResponse)
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
) (*domain.Flow, error) {
	loginVerifier := strings.TrimSpace(req.Form.Get("login_verifier"))
	if loginVerifier == "" {
		return nil, h.requestLogin(ctx, w, r, req)
	}

	return h.verifyLogin()
}

func (h *OAuthHandler) initLoginRequest(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
) error {
	return nil
}

func (h *OAuthHandler) checkSession(ctx context.Context, r *http.Request) (*domain.LoginSession, error) {
	cookie, err := h.store.Get(r, h.cfg.SessionCookieName())
	if err != nil {
		h.logger.Error("cookie store returned an error.",
			zap.Error(err),
			zap.String("method", "store.Get"),
		)
		return nil, domain.ErrNoAuthenticationSessionFound
	}

	sid := mapx.GetStringDefault(cookie.Values, domain.CookieAuthenticationSIDName, "")
	if sid == "" {
		h.logger.Error("cookie exists but session value is empty.", zap.String("method", "cookie.Values"))
		return nil, domain.ErrNoAuthenticationSessionFound
	}

	loginSession, err := h.sessionUC.GetRememberedLoginSession(ctx, nil, sid)
	if err != nil {
		h.logger.Error("cookie exists and session value exist but are not remembered any more.",
			zap.Error(err),
			zap.String("method", "sessionUC.GetRememberedLoginSession"),
		)
		return nil, err
	}

	return loginSession, nil
}

func (h *OAuthHandler) requestLogin(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
) error {
	if ar.Prompt == "login" {
		return h.initLoginRequest(ctx, w, r, ar)
	}

	loginSession, err := h.checkSession(ctx, r)
	if errors.Is(err, domain.ErrNoAuthenticationSessionFound) {
		return h.initLoginRequest(ctx, w, r, ar)
	} else if err != nil {
		return err
	}

	if ar.MaxAge > -1 && loginSession.AuthenticatedAt.Time.UTC().Add(time.Duration(ar.MaxAge)*time.Second).Before(x.NowUTC()) {
		if ar.Prompt == "none" {
			return core.ErrLoginRequired.WithHint("Request failed because prompt is set to 'none' and authentication time reached 'max_age'.")
		}
		return h.initLoginRequest(ctx, w, r, ar)
	}

	// TODO: support token hint

	return h.initLoginRequest(ctx, w, r, ar)
}

func (h *OAuthHandler) verifyLogin() (*domain.Flow, error) {
	return nil, nil
}

func (h *OAuthHandler) handleConsent(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	req *core.AuthorizeRequest,
	flow *domain.Flow,
) (*domain.Flow, error) {
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
	flow *domain.Flow,
) error {
	return nil
}

func (h *OAuthHandler) verifyConsent() (*domain.Flow, error) {
	return nil, nil
}
