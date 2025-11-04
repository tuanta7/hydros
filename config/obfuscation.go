package config

import "github.com/tuanta7/hydros/core"

type ObfuscationConfig struct {
	EncryptSessionData bool        `koanf:"encrypt_session_data" json:"encrypt_session_data"`
	SecretHasher       core.Hasher `koanf:"-" json:"-"` // bcrypt, argon2, pbkdf2, etc.
	BCryptCost         int         `koanf:"bcrypt_cost" json:"bcrypt_cost"`
	AESSecretKey       string      `koanf:"aes_secret_key" json:"aes_secret_key"`
}

func (c *Config) GetSecretsHasher() core.Hasher {
	if c.Obfuscation.SecretHasher == nil {
		// default to bcrypt
		return core.NewBCryptHasher(c.Obfuscation.BCryptCost)
	}

	return c.Obfuscation.SecretHasher
}
