package login

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type DefaultStrategy struct{}

func NewDefaultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

func (d *DefaultStrategy) Login(ctx context.Context, credentials *Credentials) error {
	if credentials.Username != "admin@example.com" || credentials.Password != "password" {
		return core.ErrRequestUnauthorized.WithHint("Invalid username or password")
	}

	return nil
}
