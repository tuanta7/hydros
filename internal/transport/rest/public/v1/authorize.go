package v1

import (
	"errors"
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
	err = h.oauthUC.HandleTokenEndpoint(c.Request.Context(), request, response)
	if err != nil {
		var rfc6749Error *core.RFC6749Error
		if errors.As(err, &rfc6749Error) {
			c.JSON(rfc6749Error.CodeField, rfc6749Error.DescriptionField)
			return
		}

		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusFound, request.RedirectURI)
}
