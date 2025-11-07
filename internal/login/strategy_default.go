package login

import "context"

type DefaultStrategy struct{}

func NewDefaultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

func (d *DefaultStrategy) Login(ctx context.Context, credentials *Credentials) error {
	return nil
}
