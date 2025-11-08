package config

import (
	"net/url"

	"github.com/tuanta7/hydros/core/x"
)

type OIDCConfig struct {
	AllowedPrompts []string
}

func (c *Config) GetAllowedPrompts() []string {
	return c.OIDC.AllowedPrompts
}

func (c *Config) GetRedirectSecureChecker() func(*url.URL) bool {
	return x.IsURISecure
}
