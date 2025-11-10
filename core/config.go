package core

import (
	"hash"
	"net/url"
	"time"
)

type DebugModeProvider interface {
	IsDebugging() bool
}

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

type IDTokenLifetimeProvider interface {
	GetIDTokenLifetime() time.Duration
}

type SecretsHasherProvider interface {
	GetSecretsHasher() Hasher
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

type IDTokenIssuerProvider interface {
	GetIDTokenIssuer() string
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

type MinParameterEntropyProvider interface {
	GetMinParameterEntropy() int
}

type EnablePKCEPlainChallengeMethodProvider interface {
	IsEnablePKCEPlainChallengeMethod() bool
}

type AllowedPromptsProvider interface {
	GetAllowedPrompts() []string
}

type RedirectSecureCheckerProvider interface {
	GetRedirectSecureChecker() func(*url.URL) bool
}
