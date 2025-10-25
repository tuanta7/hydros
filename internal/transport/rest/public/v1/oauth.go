package v1

import (
	"github.com/tuanta7/hydros/internal/usecase/client"
	"github.com/tuanta7/hydros/internal/usecase/oauth"
)

type OAuthHandler struct {
	clientUC client.UseCase
	oauthUC  oauth.UseCase
}

func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{}
}
