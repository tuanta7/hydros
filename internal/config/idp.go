package config

import (
	"net/url"
	"time"
)

type IdentityConfig struct {
	RegistrationURL      string        `koanf:"registration_url"`
	LoginPageURL         string        `koanf:"login_page_url"`
	ConsentPageURL       string        `koanf:"consent_page_url"`
	ConsentRequestMaxAge time.Duration `koanf:"consent_request_max_age"`
}

func (c *Config) GetConsentRequestMaxAge() time.Duration {
	if c.Identity.ConsentRequestMaxAge == 0 {
		return 10 * time.Minute
	}

	return c.Identity.ConsentRequestMaxAge
}

func (c *Config) GetRegistrationURL() *url.URL {
	def, _ := url.Parse("/registration")
	return urlWithDefault(c.Identity.RegistrationURL, def)
}

func (c *Config) GetLoginPageURL() *url.URL {
	def, _ := url.Parse("/self-service/login")
	return urlWithDefault(c.Identity.LoginPageURL, def)
}

func (c *Config) GetConsentPageURL() *url.URL {
	def, _ := url.Parse("/self-service/consent")
	return urlWithDefault(c.Identity.ConsentPageURL, def)
}

func urlWithDefault(s string, def *url.URL) *url.URL {
	parsed, err := url.Parse(s)
	if err == nil && parsed.String() != "" {
		return parsed
	}

	return def
}
