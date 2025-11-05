package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/internal/flow"
)

type FlowHandler struct {
	flowUC *flow.UseCase
}

func NewFlowHandler(uc *flow.UseCase) *FlowHandler {
	return &FlowHandler{
		flowUC: uc,
	}
}

func (h *FlowHandler) GetFlow(c *gin.Context) {
	challenge := c.Param("challenge")
	fmt.Println(challenge)
	c.JSON(http.StatusOK, gin.H{})
}

func (h *FlowHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (h *FlowHandler) Login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "login successfully",
	})
}

func (h *FlowHandler) UpdateAuthenticationStatus(c *gin.Context) {}

func (h *FlowHandler) ConsentPage(c *gin.Context) {
	c.HTML(http.StatusOK, "consent.html", nil)
}

func (h *FlowHandler) UpdateConsentStatus(c *gin.Context) {}
