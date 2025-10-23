package core

type TokenRequest struct {
	GrantType    Arguments
	Code         string
	CodeVerifier string
	RedirectURI  string
	Request
}

type TokenResponse struct{}
