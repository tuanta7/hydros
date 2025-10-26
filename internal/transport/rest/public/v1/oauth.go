package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/core"
	clientuc "github.com/tuanta7/hydros/internal/usecase/client"
)

type OAuthHandler struct {
	clientUC             *clientuc.UseCase
	authorizeInteractors []core.AuthorizeInteractor
	tokenInteractors     []core.TokenInteractor
}

func NewOAuthHandler(
	clientUC *clientuc.UseCase,
	authorizeInteractors []core.AuthorizeInteractor,
	tokenInteractors []core.TokenInteractor,
) *OAuthHandler {
	return &OAuthHandler{
		clientUC:             clientUC,
		authorizeInteractors: authorizeInteractors,
		tokenInteractors:     tokenInteractors,
	}
}

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

	for _, ah := range h.tokenInteractors {
		err := ah.HandleTokenRequest(c.Request.Context(), request, response)
		if errors.Is(err, core.ErrUnknownRequest) {
			// skip to the next token interactor
			continue
		} else if err != nil {
			var rfc6749Error *core.RFC6749Error
			if errors.As(err, &rfc6749Error) {
				c.JSON(rfc6749Error.CodeField, rfc6749Error.DescriptionField)
				return
			}

			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.Redirect(http.StatusFound, request.RedirectURI)
}
