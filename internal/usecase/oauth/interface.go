package oauth

import (
	"context"

	"github.com/tuanta7/hydros/internal/domain"
)

type AccessTokenRepository interface {
	GetBySignature(ctx context.Context, signature string) (*domain.AccessToken, error)
	Create(ctx context.Context, token *domain.AccessToken) error
	Delete(ctx context.Context, signature string) error
}
