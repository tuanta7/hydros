package v1

import (
	"context"

	"github.com/tuanta7/oauth-server/internal/core"
	"github.com/tuanta7/oauth-server/internal/domain"
)

type ClientUC interface {
	GetClient(ctx context.Context, id string) (*domain.Client, error)
}

type OAuthHandler struct {
	clientUC    ClientUC
	authorizers core.AuthorizeHandlerChain
}

func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{}
}
