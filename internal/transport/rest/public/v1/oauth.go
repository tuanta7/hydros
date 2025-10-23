package v1

import (
	"context"

	"github.com/tuanta7/hydros/internal/core"
	"github.com/tuanta7/hydros/internal/domain"
)

type ClientUC interface {
	GetClient(ctx context.Context, id string) (*domain.Client, error)
}

type OAuthHandler struct {
	clientUC          ClientUC
	authorizeHandlers []core.AuthorizeHandler
	tokenHandlers     []core.TokenHandler
}

func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{}
}
