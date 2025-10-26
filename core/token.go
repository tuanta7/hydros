package core

type TokenRequest struct {
	GrantType        Arguments `json:"grant_type" form:"grant_type"`
	HandledGrantType Arguments `json:"handled_grant_type" form:"handled_grant_type"`
	RedirectURI      string    `json:"redirect_uri" form:"redirect_uri"`
	Code             string    `json:"code" form:"code"`
	CodeVerifier     string    `json:"code_verifier" form:"code_verifier"`
	Request
}

type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in,omitempty"`
	Scope       Arguments `json:"scope,omitempty"`
	OIDCTokenResponse
}

type OIDCTokenResponse struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}
