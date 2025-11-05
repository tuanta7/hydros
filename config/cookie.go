package config

type CookieConfig struct {
	Domain      string `koanf:"domain"`
	Path        string `koanf:"path"`
	SameSite    string `koanf:"same_site"`
	Secure      bool   `koanf:"secure"`
	SessionName string `koanf:"session_name"`
}

func (c *Config) SessionCookieName() string {
	return c.Cookie.SessionName
}
