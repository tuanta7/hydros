package config

import "fmt"

type JWTConfig struct {
	Algorithm         string
	KeySize           int
	AccessTokenIssuer string `koanf:"access_token_issuer"`
}

func (c *Config) GetAccessTokenIssuer() string {
	if c.JWT.AccessTokenIssuer == "" {
		// TODO: find a better way to get the default issuer URL
		return fmt.Sprintf("https://%s:%s", c.RestServerHost, c.RestServerPort)
	}
	return c.JWT.AccessTokenIssuer
}

func (c *Config) GetAccessTokenAlgorithm() string {
	supported := []string{"RS256", "RS512", "HS256", "HS512"}
	for _, alg := range supported {
		if alg == c.JWT.Algorithm {
			return alg
		}
	}

	return "RS256"
}
