package v1

import (
	"net/http"

	"github.com/bytedance/gopkg/util/logger"
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
	request, err := h.oauth2.NewTokenRequest(c.Request.Context(), c.Request, session)
	if err != nil {
		logger.Error(err, zap.String("method", "oauth2.NewTokenRequest"))
		return
	}

	c.Redirect(http.StatusFound, request.RedirectURI)
}
