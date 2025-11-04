package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ViewHandler struct{}

func NewViewHandler() *ViewHandler {
	return &ViewHandler{}
}

func (v *ViewHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}
