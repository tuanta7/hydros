package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/core"
)

func (h *OAuthHandler) Token(c *gin.Context) {
	var request *core.TokenRequest
	err := c.BindQuery(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	client, err := h.clientUC.Get(c.Request.Context(), request.Client.GetID())
	if err != nil {
		return
	}
	request.Client = client

	response := &core.TokenResponse{}
	_ = h.oauthUC.HandleTokenEndpoint(c.Request.Context(), request, response)

}
