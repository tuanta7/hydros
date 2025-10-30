package config

import "fmt"

type OAuthConfig struct {
	DisableRefreshTokenValidation bool   `koanf:"disable_refresh_token_validation"`
	AccessTokenFormat             string `koanf:"access_token_format"` // must be "jwt" or "opaque"
	AccessTokenIssuer             string `koanf:"access_token_issuer"`
}

func (c *Config) IsDisableRefreshTokenValidation() bool {
	return c.OAuth.DisableRefreshTokenValidation
}

func (c *Config) GetAccessTokenIssuer() string {
	if c.OAuth.AccessTokenIssuer == "" {
		// TODO: find a better way to get the default issuer URL
		return fmt.Sprintf("https://%s:%s", c.RestServerHost, c.RestServerPort)
	}
	return c.OAuth.AccessTokenIssuer
}

func (c *Config) GetAccessTokenFormat() string {
	if c.OAuth.AccessTokenFormat == "jwt" {
		return "jwt"
	}

	return "opaque"
}
