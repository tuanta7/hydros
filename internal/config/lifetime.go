package config

import "time"

type LifetimeConfig struct {
	AuthorizationCode time.Duration `koanf:"authorize_code" default:"10m"`
	AccessToken       time.Duration `koanf:"access_token" default:"1h"`
	RefreshToken      time.Duration `koanf:"refresh_token" default:"720h"`
}

func (c *Config) GetRefreshTokenLifetime() time.Duration {
	if c.Lifetime.RefreshToken == 0 {
		return time.Hour * 24 * 30
	}
	return c.Lifetime.RefreshToken
}

func (c *Config) GetAuthorizationCodeLifetime() time.Duration {
	if c.Lifetime.AuthorizationCode == 0 {
		return time.Minute * 10
	}
	return c.Lifetime.AuthorizationCode
}

func (c *Config) GetAccessTokenLifetime() time.Duration {
	if c.Lifetime.AccessToken == 0 {
		return time.Hour
	}
	return c.Lifetime.AccessToken
}
