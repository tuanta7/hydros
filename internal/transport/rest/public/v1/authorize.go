package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/internal/core"
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

	fmt.Println(client)

	resp := &core.AuthorizeResponse{}
	for _, ah := range h.authorizeHandlers {
		_ = ah.HandleAuthorizeRequest(c.Request.Context(), request, resp)
	}
}
