package config

import "time"

type LifetimeConfig struct {
	AuthorizeCode time.Duration `koanf:"authorize_code" default:"10m"`
	AccessToken   time.Duration `koanf:"access_token" default:"1h"`
	RefreshToken  time.Duration `koanf:"refresh_token" default:"720h"`
}

func (c *Config) GetAuthorizeCodeLifetime() time.Duration {
	return c.Lifetime.AuthorizeCode
}
