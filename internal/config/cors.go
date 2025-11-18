package config

type CORSConfig struct {
	AllowOrigins     []string `koanf:"allow_origins"`
	AllowMethods     []string `koanf:"allow_methods"`
	AllowHeaders     []string `koanf:"allow_headers"`
	AllowCredentials bool     `koanf:"allow_credentials"`
	MaxAge           int      `koanf:"max_age"`
	ExposeHeaders    []string `koanf:"expose_headers"`
}
