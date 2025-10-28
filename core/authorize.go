package core

import (
	"context"
	"net/http"
	"net/url"
)

type ResponseModeType string

const (
	ResponseModeDefault  ResponseModeType = ""
	ResponseModeFormPost ResponseModeType = "form_post"
	ResponseModeQuery    ResponseModeType = "query"
	ResponseModeFragment ResponseModeType = "fragment"
)

type AuthorizeRequest struct {
	State               string           `json:"state"`
	RedirectURI         *url.URL         `json:"redirectUri"`
	ResponseTypes       Arguments        `json:"responseTypes"`
	ResponseMode        ResponseModeType `json:"responseMode"`
	CodeChallenge       string           `json:"codeChallenge"`
	CodeChallengeMethod string           `json:"codeChallengeMethod"`
	Request
}

type AuthorizeResponse struct {
	Header     http.Header
	Parameters url.Values
	Code       string
}

func (o *OAuth2) NewAuthorizeRequest(ctx context.Context, req *http.Request) (*AuthorizeRequest, error) {
	return nil, nil
}

func (o *OAuth2) NewAuthorizeResponse(ctx context.Context, req *AuthorizeRequest, session Session) (*AuthorizeResponse, error) {
	return nil, nil
}

func (o *OAuth2) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, err error) {

}

func (o *OAuth2) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, resp *AuthorizeResponse) {
}
