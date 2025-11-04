package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ViewHandler struct{}

func NewViewHandler() *ViewHandler {
	return &ViewHandler{}
}

func (v *ViewHandler) ErrorPage(c *gin.Context) {
	c.HTML(http.StatusOK, "errors.html", gin.H{
		"Error":       c.Query("error"),
		"Description": c.Query("error_description"),
		"Debug":       c.Query("debug"),
		"Hint":        c.Query("hint"),
		"Cause":       c.Query("cause"),
	})
}

func (v *ViewHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}
