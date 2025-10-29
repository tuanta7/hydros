package config

import "fmt"

type OAuthConfig struct {
	DisableRefreshTokenValidation bool   `koanf:"disable_refresh_token_validation"`
	TokenIssuer                   string `koanf:"token_issuer"`
}

func (c *Config) IsDisableRefreshTokenValidation() bool {
	return c.OAuth.DisableRefreshTokenValidation
}

func (c *Config) GetIssuerURL() string {
	if c.OAuth.TokenIssuer == "" {
		// TODO: find a better way to get the default issuer URL
		return fmt.Sprintf("https://%s:%s", c.RestServerHost, c.RestServerPort)
	}
	return c.OAuth.TokenIssuer
}
