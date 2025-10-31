package config

import "github.com/tuanta7/hydros/core/strategy"

const (
	MinParameterEntropy = 8
)

type OAuthConfig struct {
	ScopeStrategy                   string `koanf:"scope_strategy"`
	AudienceStrategy                string `koanf:"audience_strategy"`
	MinParameterEntropy             int    `koanf:"min_parameter_entropy"`
	DisableRefreshTokenValidation   bool   `koanf:"disable_refresh_token_validation"`
	AccessTokenFormat               string `koanf:"access_token_format"` // must be "jwt" or "opaque"
	EnforcePKCE                     bool   `koanf:"enforce_pkce"`
	EnforcePKCEForPublicClients     bool   `koanf:"enforce_pkce_for_public_clients"`
	DisablePKCEPlainChallengeMethod bool   `koanf:"disable_pkce_plain_challenge_method"`
}

func (c *Config) GetScopeStrategy() strategy.ScopeStrategy {
	switch c.OAuth.ScopeStrategy {
	case "exact":
		return strategy.ExactScopeStrategy
	case "hierarchical":
		return strategy.HierarchicScopeStrategy
	}

	return strategy.ExactScopeStrategy
}

func (c *Config) GetAudienceStrategy() strategy.AudienceStrategy {
	switch c.OAuth.AudienceStrategy {
	case "exact":
		return strategy.ExactAudienceStrategy
	}

	return strategy.ExactAudienceStrategy
}

func (c *Config) GetMinParameterEntropy() int {
	if c.OAuth.MinParameterEntropy == 0 {
		return MinParameterEntropy
	} else {
		return c.OAuth.MinParameterEntropy
	}
}

func (c *Config) IsDisableRefreshTokenValidation() bool {
	return c.OAuth.DisableRefreshTokenValidation
}

func (c *Config) GetAccessTokenFormat() string {
	if c.OAuth.AccessTokenFormat == "jwt" {
		return "jwt"
	}

	return "opaque"
}

func (c *Config) GetEnforcePKCE() bool {
	return c.OAuth.EnforcePKCE
}

func (c *Config) GetEnforcePKCEForPublicClients() bool {
	return c.OAuth.EnforcePKCEForPublicClients
}

func (c *Config) IsEnablePKCEPlainChallengeMethod() bool {
	return c.OAuth.DisablePKCEPlainChallengeMethod
}
