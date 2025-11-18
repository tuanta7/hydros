package config

import (
	"crypto/sha512"
	"hash"
)

type HMACConfig struct {
	GlobalSecret  string           `koanf:"global_secret" json:"-" validate:"required"`
	RotatedSecret string           `koanf:"rotated_secret" json:"-"`
	KeyEntropy    int              `koanf:"key_entropy"`
	Hasher        func() hash.Hash `koanf:"-" json:"-"`
}

func (c *Config) GetTokenEntropy() int {
	if c.HMAC.KeyEntropy < 64 {
		return 64
	}

	return c.HMAC.KeyEntropy
}

func (c *Config) GetHMACHasher() func() hash.Hash {
	if c.HMAC.Hasher != nil {
		return c.HMAC.Hasher
	}

	return sha512.New512_256
}

func (c *Config) GetGlobalSecret() []byte {
	return []byte(c.HMAC.GlobalSecret)
}
