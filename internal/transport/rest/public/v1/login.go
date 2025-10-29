package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/core"
)

type LoginHandler struct {
	oauth2 core.OAuth2Provider
}

func (h *LoginHandler) HandleLoginFlow(c *gin.Context) {
	ctx := c.Request.Context()
	authorizeRequest, err := h.oauth2.NewAuthorizeRequest(ctx, c.Request)
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Writer, authorizeRequest, err)
		return
	}

}
