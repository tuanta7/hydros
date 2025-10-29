package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/internal/domain"
	clientuc "github.com/tuanta7/hydros/internal/usecase/client"
)

type ClientHandler struct {
	clientUC *clientuc.UseCase
}

func NewClientHandler(clientUC *clientuc.UseCase) *ClientHandler {
	return &ClientHandler{
		clientUC: clientUC,
	}
}

func (h *ClientHandler) List(c *gin.Context) {
	clients, err := h.clientUC.ListClients(c.Request.Context(), 1, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	jsonPayload, err := json.Marshal(clients)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, jsonPayload)
}

func (h *ClientHandler) Create(c *gin.Context) {
	var client domain.Client
	if err := c.ShouldBindJSON(&c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.clientUC.CreateClient(c.Request.Context(), &client)
	if err != nil {
		return
	}

	c.Status(http.StatusCreated)
}
