package core

import (
	"time"
)

type AuthorizationCodeLifetimeProvider interface {
	GetAuthorizationCodeLifetime() time.Duration
}

type AccessTokenLifetimeProvider interface {
	GetAccessTokenLifetime() time.Duration
}

type RefreshTokenLifetimeProvider interface {
	GetRefreshTokenLifetime() time.Duration
}
