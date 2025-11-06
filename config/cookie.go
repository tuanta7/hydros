package config

import "net/http"

type CookieConfig struct {
	Domain      string `koanf:"domain"`
	Path        string `koanf:"path"`
	SameSite    string `koanf:"same_site"`
	Secure      bool   `koanf:"secure"`
	SessionName string `koanf:"session_name"`
	// KeyPairs    [][]byte `koanf:"key_pairs"`
}

func (c *Config) CookieKeyPairs() [][]byte {
	var keyPairs [][]byte
	keyPairs = append(keyPairs, c.GetGlobalSecret())
	return keyPairs
}

func (c *Config) CookieSessionName() string {
	return c.Cookie.SessionName
}

func (c *Config) CookieDomain() string {
	return c.Cookie.Domain
}

func (c *Config) CookiePath() string {
	return c.Cookie.Path
}

func (c *Config) CookieSameSiteMode() http.SameSite {
	switch c.Cookie.SameSite {
	case "strict":
		// only sent if the request is coming from the same site that set it
		return http.SameSiteStrictMode
	case "lax":
		// sent for same-site requests and for cross-site requests initiated by a top-level navigation
		return http.SameSiteLaxMode
	case "none":
		// sent with all cross-site requests, regardless of the method or navigation type
		// secure=true is required
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

func (c *Config) CookieSecure() bool {
	return c.Cookie.Secure
}
