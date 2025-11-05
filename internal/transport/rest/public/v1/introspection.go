package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/internal/session"
	"go.uber.org/zap"
)

func (h *OAuthHandler) HandleIntrospectionRequest(c *gin.Context) {
	ctx := c.Request.Context()
	session := session.NewSession("")

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
