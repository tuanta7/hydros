package config

import "net/http"

type SessionCookieConfig struct {
	Domain     string `koanf:"domain"`
	Path       string `koanf:"path"`
	SameSite   string `koanf:"same_site"`
	Secure     bool   `koanf:"secure"`
	SessionKey string `koanf:"session_key"`
	// KeyPairs    [][]byte `koanf:"key_pairs"`
}

func (c *Config) CookieKeyPairs() [][]byte {
	var keyPairs [][]byte
	keyPairs = append(keyPairs, c.GetGlobalSecret())
	return keyPairs
}

func (c *Config) SessionCookieKey() string {
	if c.SessionCookie.SessionKey == "" {
		return "login_session"
	}
	return c.SessionCookie.SessionKey
}

func (c *Config) SessionCookieDomain() string {
	return c.SessionCookie.Domain
}

func (c *Config) SessionCookiePath() string {
	return c.SessionCookie.Path
}

func (c *Config) SessionCookieSameSiteMode() http.SameSite {
	switch c.SessionCookie.SameSite {
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

func (c *Config) SessionCookieSecure() bool {
	return c.SessionCookie.Secure
}
