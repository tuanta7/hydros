package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/oauth-server/internal/core"
)

func (h *OAuthHandler) Authorize(c *gin.Context) {
	var request *core.AuthorizeRequest
	err := c.BindQuery(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	client, err := h.clientUC.GetClient(c.Request.Context(), request.Client.GetID())
	if err != nil {
		return
	}

	resp := &core.AuthorizeResponse{}
	for _, ah := range h.authorizers {
		_ = ah.HandleAuthorizeRequest(c.Request.Context(), request, resp)
	}
}
