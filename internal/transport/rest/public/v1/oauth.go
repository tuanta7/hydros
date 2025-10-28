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

func NewOAuthHandler(oauth2 core.OAuth2Provider) *OAuthHandler {
	return &OAuthHandler{
		oauth2: oauth2,
	}
}

func (h *OAuthHandler) Token(c *gin.Context) {
	session := domain.NewSession()
	tokenRequest, err := h.oauth2.NewTokenRequest(c.Request.Context(), c.Request, session)
	if err != nil {
		h.logger.Error("error validating token request",
			zap.Error(err),
			zap.String("method", "oauth2.NewTokenRequest"),
		)
		return
	}

	// TODO: Implement rfc8693 token exchange

	tokenResponse, err := h.oauth2.NewTokenResponse(c.Request.Context(), tokenRequest)
	if err != nil {
		h.logger.Error("error populating token response",
			zap.Error(err),
			zap.String("method", "oauth2.NewTokenResponse"),
		)
		h.oauth2.WriteTokenError(c.Request.Context(), c.Writer, tokenRequest, err)
		return
	}

	h.oauth2.WriteTokenResponse(c.Request.Context(), c.Writer, tokenRequest, tokenResponse)
}
