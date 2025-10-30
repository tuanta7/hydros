package core

import (
	"hash"
	"time"
)

type LifetimeConfigProvider interface {
	AccessTokenLifetimeProvider
	RefreshTokenLifetimeProvider
	AuthorizationCodeLifetimeProvider
}

type AuthorizationCodeLifetimeProvider interface {
	GetAuthorizationCodeLifetime() time.Duration
}

type AccessTokenLifetimeProvider interface {
	GetAccessTokenLifetime() time.Duration
}

type RefreshTokenLifetimeProvider interface {
	GetRefreshTokenLifetime() time.Duration
}

type SecretsHashingProvider interface {
	GetSecretsHasher() Hasher
}

type DebugModeProvider interface {
	IsDebugging() bool
}

type DisableRefreshTokenValidationProvider interface {
	IsDisableRefreshTokenValidation() bool
}

type AccessTokenFormatProvider interface {
	GetAccessTokenFormat() string
}

type AccessTokenIssuerProvider interface {
	GetAccessTokenIssuer() string
}

type TokenEntropyProvider interface {
	GetTokenEntropy() int
}

type GlobalSecretProvider interface {
	GetGlobalSecret() []byte
}

type HMACHashingProvider interface {
	GetHMACHasher() func() hash.Hash
}
