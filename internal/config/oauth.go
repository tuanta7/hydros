package config

import "github.com/tuanta7/hydros/core/strategy"

const (
	MinParameterEntropy = 8

	MatchExact  = "exact"
	MatchPrefix = "prefix"

	AccessTokenFormatOpaque = "opaque"
	AccessTokenFormatJWT    = "jwt"
)

type OAuthConfig struct {
	ScopeStrategy                  string `koanf:"scope_strategy"`
	AudienceStrategy               string `koanf:"audience_strategy"`
	AccessTokenFormat              string `koanf:"access_token_format"`
	MinParameterEntropy            int    `koanf:"min_parameter_entropy"`
	DisableRefreshTokenValidation  bool   `koanf:"disable_refresh_token_validation"`
	EnablePKCEPlainChallengeMethod bool   `koanf:"enable_pkce_plain_challenge_method"`
}

func (c *Config) GetScopeStrategy() strategy.ScopeStrategy {
	switch c.OAuth.ScopeStrategy {
	case MatchExact:
		return strategy.ExactScopeStrategy
	case MatchPrefix:
		return strategy.PrefixScopeStrategy
	default:
		return strategy.ExactScopeStrategy
	}
}

func (c *Config) GetAudienceStrategy() strategy.AudienceStrategy {
	switch c.OAuth.AudienceStrategy {
	case MatchExact:
		return strategy.ExactAudienceStrategy
	default:
		return strategy.ExactAudienceStrategy
	}
}

func (c *Config) GetMinParameterEntropy() int {
	if c.OAuth.MinParameterEntropy < MinParameterEntropy {
		return MinParameterEntropy
	}

	return c.OAuth.MinParameterEntropy
}

func (c *Config) GetAccessTokenFormat() string {
	if c.OAuth.AccessTokenFormat == AccessTokenFormatJWT {
		return AccessTokenFormatJWT
	}

	return AccessTokenFormatOpaque
}

func (c *Config) IsDisableRefreshTokenValidation() bool {
	return c.OAuth.DisableRefreshTokenValidation
}

func (c *Config) IsEnablePKCEPlainChallengeMethod() bool {
	return c.OAuth.EnablePKCEPlainChallengeMethod
}
