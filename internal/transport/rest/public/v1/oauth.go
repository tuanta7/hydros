package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type OAuthHandler struct {
	oauth2 core.OAuth2Provider
	logger *zapx.Logger
}

func NewOAuthHandler(oauth2 core.OAuth2Provider, logger *zapx.Logger) *OAuthHandler {
	return &OAuthHandler{
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

func (h *OAuthHandler) HandleIntrospectionRequest(c *gin.Context) {}
