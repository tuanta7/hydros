package login

import "context"

type DefaultIdentityProvider struct{}

func NewDefaultIdentityProvider() *DefaultIdentityProvider {
	return &DefaultIdentityProvider{}
}

func (d *DefaultIdentityProvider) Login(ctx context.Context, credentials *Credentials) error {
	return nil
}
