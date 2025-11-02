package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/internal/usecase/login"
)

type LoginHandler struct {
	loginUC *login.UseCase
}

func (h *LoginHandler) InitLogin(c *gin.Context) {}

func (h *LoginHandler) UpdateLoginStatus(c *gin.Context) {}
