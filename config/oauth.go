package config

type OAuthConfig struct {
	DisableRefreshTokenValidation bool   `koanf:"disable_refresh_token_validation"`
	AccessTokenFormat             string `koanf:"access_token_format"` // must be "jwt" or "opaque"
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
