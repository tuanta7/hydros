package config

type ObfuscationConfig struct {
	EncryptSessionData bool `koanf:"encrypt_session_data" default:"false"`
}
