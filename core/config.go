package core

import (
	"time"
)

type AuthorizeCodeLifetimeProvider interface {
	GetAuthorizeCodeLifetime() time.Duration
}

type AccessTokenLifetimeProvider interface {
	GetAccessTokenLifetime() time.Duration
}

type RefreshTokenLifetimeProvider interface {
	GetRefreshTokenLifetime() time.Duration
}
