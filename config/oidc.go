package config

import (
	"net/url"
	"time"

	"github.com/tuanta7/hydros/core/x"
)

type OIDCConfig struct {
	AllowedPrompts  []string
	Issuer          string
	IDTokenLifetime time.Duration
}

func (c *Config) GetAllowedPrompts() []string {
	return c.OIDC.AllowedPrompts
}

func (c *Config) GetRedirectSecureChecker() func(*url.URL) bool {
	return x.IsURISecure
}

func (c *Config) GetIDTokenIssuer() string {
	return c.OIDC.Issuer
}

func (c *Config) GetIDTokenLifetime() time.Duration {
	if c.OIDC.IDTokenLifetime == 0 {
		return time.Hour
	}

	return c.OIDC.IDTokenLifetime
}
