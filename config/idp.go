package config

type IdentityConfig struct {
	LoginPageURL     string `koanf:"login_page_url"`
	LoginSubmitURL   string `koanf:"login_submit_url"`
	ConsentPageURL   string `koanf:"consent_page_url"`
	ConsentSubmitURL string `koanf:"consent_submit_url"`
}
