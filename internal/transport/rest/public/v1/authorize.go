package v1

import (
	"context"
	stderr "errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/internal/errors"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/internal/jwk"
	"github.com/tuanta7/hydros/internal/session"
	"github.com/tuanta7/hydros/pkg/helper/mapx"

	"github.com/tuanta7/hydros/pkg/logger"
	"go.uber.org/zap"
)

const (
	CookieLoginSessionIDKey = "sid"
)

type OAuthHandler struct {
	cfg       *config.Config
	store     *sessions.CookieStore
	oauth2    core.OAuth2Provider
	jwkUC     *jwk.UseCase
	sessionUC session.UseCase
	flowUC    *flow.UseCase
	logger    *logger.Logger
}

func NewOAuthHandler(
	cfg *config.Config,
	store *sessions.CookieStore,
	oauth2 core.OAuth2Provider,
	jwkUC *jwk.UseCase,
	sessionUC session.UseCase,
	flowUC *flow.UseCase,
	logger *logger.Logger,
) *OAuthHandler {
	return &OAuthHandler{
		cfg:       cfg,
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

	f, err := h.handleAuthorizeRequest(ctx, c.Writer, c.Request, ar)
	if stderr.Is(err, errors.ErrAbortOAuth2Request) {
		return
	} else if err != nil {
		h.writeAuthorizeError(c, ar, err)
		return
	}

	ar.GrantedAudience = ar.GrantedAudience.Append(f.GrantedAudience...)
	ar.GrantedScope = ar.GrantedScope.Append(f.GrantedScope...)

	authorizeResponse, err := h.oauth2.NewAuthorizeResponse(ctx, ar, &session.Session{
		IDTokenSession: &oidc.IDTokenSession{
			Subject: f.Subject, // id of authenticated user
		},
		Flow: f,
	})
	if err != nil {
		h.writeAuthorizeError(c, ar, err)
		return
	}

	h.oauth2.WriteAuthorizeResponse(ctx, c.Writer, ar, authorizeResponse)
}

func (h *OAuthHandler) handleAuthorizeRequest(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	req *core.AuthorizeRequest,
) (*flow.Flow, error) {
	loginVerifier := strings.TrimSpace(req.Form.Get("login_verifier"))
	consentVerifier := strings.TrimSpace(req.Form.Get("consent_verifier"))
	if loginVerifier == "" && consentVerifier == "" {
		return nil, h.requestLogin(ctx, w, r, req)
	} else if loginVerifier != "" {
		f, err := h.verifyLogin(ctx, w, r, loginVerifier)
		if err != nil {
			return nil, err
		}

		return nil, h.requestConsent(ctx, w, r, req, f)
	}

	return h.verifyConsent(ctx, r, consentVerifier)
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

	return loginSession, nil
}
