package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type OAuthHandler struct {
	cfg    *config.Config
	oauth2 core.OAuth2Provider
	logger *zapx.Logger
}

func NewOAuthHandler(cfg *config.Config, oauth2 core.OAuth2Provider, logger *zapx.Logger) *OAuthHandler {
	return &OAuthHandler{
		cfg:    cfg,
		oauth2: oauth2,
		logger: logger,
	}
}

func (h *OAuthHandler) HandleAuthorizeRequest(c *gin.Context) {
	ctx := c.Request.Context()
	h.oauth2.WriteAuthorizeResponse(ctx, c.Writer, nil, nil)
}

func (h *OAuthHandler) HandleTokenRequest(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.NewSession("")
	tokenRequest, err := h.oauth2.NewTokenRequest(ctx, c.Request, session)
	if err != nil {
		h.logger.Error("error validating token request",
			zap.Error(err),
			zap.String("method", "oauth2.NewTokenRequest"),
		)
		h.oauth2.WriteTokenError(ctx, c.Writer, tokenRequest, err)
		return
	}

	tokenRequest.GrantedScope = tokenRequest.GrantedScope.Append(tokenRequest.Scope...)
	tokenRequest.GrantedAudience = tokenRequest.GrantedAudience.Append(tokenRequest.Audience...)

	if tokenRequest.GrantType.ExactOne(string(core.GrantTypeClientCredentials)) {
		session.Subject = tokenRequest.Client.GetID()
	}

	session.ClientID = tokenRequest.Client.GetID()
	session.IDTokenSession.Claims.Issuer = h.cfg.GetAccessTokenIssuer()
	session.IDTokenSession.Claims.IssuedAt = x.NowUTC()

	// TODO: Implement rfc8693 token exchange

	tokenResponse, err := h.oauth2.NewTokenResponse(ctx, tokenRequest)
	if err != nil {
		h.logger.Error("error populating token response",
			zap.Error(err),
			zap.String("method", "oauth2.NewTokenResponse"),
		)
		h.oauth2.WriteTokenError(ctx, c.Writer, tokenRequest, err)
		return
	}

	h.oauth2.WriteTokenResponse(ctx, c.Writer, tokenRequest, tokenResponse)
}

func (h *OAuthHandler) HandleIntrospectionRequest(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.NewSession("")

	resp, err := h.oauth2.IntrospectToken(ctx, c.Request, session)
	if err != nil {
		h.logger.Error("error while introspecting token",
			zap.Error(err),
			zap.String("method", "oauth2.IntrospectToken"),
		)
		h.oauth2.WriteIntrospectionError(ctx, c.Writer, err)
		return
	}

	h.oauth2.WriteIntrospectionResponse(ctx, c.Writer, resp)
}
