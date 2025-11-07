package login

import "context"

type IdentityProvider interface {
	Login(ctx context.Context, credentials *Credentials) error
}
