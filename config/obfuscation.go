package config

import "github.com/tuanta7/hydros/core"

type ObfuscationConfig struct {
	EncryptSessionData bool        `koanf:"encrypt_session_data"`
	BCryptCost         int         `koanf:"bcrypt_cost"`
	SecretHasher       core.Hasher `koanf:"-" ` // bcrypt, argon2, pbkdf2, etc.
}

func (c *Config) GetSecretsHasher() core.Hasher {
	if c.Obfuscation.SecretHasher == nil {
		// default to bcrypt
		return core.NewBCryptHasher(c.Obfuscation.BCryptCost)
	}

	return c.Obfuscation.SecretHasher
}
