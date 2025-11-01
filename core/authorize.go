package core

import (
	"context"
	"net/http"
	"net/url"

	"github.com/tuanta7/hydros/core/x"
)

type ResponseMode string

const (
	ResponseModeDefault  ResponseMode = ""
	ResponseModeFormPost ResponseMode = "form_post"
	ResponseModeQuery    ResponseMode = "query"
	ResponseModeFragment ResponseMode = "fragment"
)

type AuthorizeRequest struct {
	State               string       `json:"state" form:"state"`
	RedirectURI         *url.URL     `json:"redirect_uri" form:"redirect_uri"`
	ResponseTypes       Arguments    `json:"response_types" form:"response_types"`
	ResponseMode        ResponseMode `json:"response_mode" form:"response_mode"`
	DefaultResponseMode ResponseMode `json:"default_response_mode" form:"-"`
	CodeChallenge       string       `json:"code_challenge" form:"code_challenge"`
	CodeChallengeMethod string       `json:"code_challenge_method" form:"code_challenge_method"`
	Request
}

func NewAuthorizeRequest() *AuthorizeRequest {
	return &AuthorizeRequest{
		ResponseTypes: Arguments{},
		Request:       *NewRequest(),
		ResponseMode:  ResponseModeDefault,
		RedirectURI:   nil, // must be set to nil for redirect detection to work properly
	}
}

type AuthorizeResponse struct {
	Code  string
	State string
	Scope string
}

func (o *OAuth2) NewAuthorizeRequest(ctx context.Context, req *http.Request) (*AuthorizeRequest, error) {
	authorizeRequest := NewAuthorizeRequest()

	form, err := x.BindPostForm(req)
	if err != nil {
		return nil, ErrInvalidRequest.WithHint("Unable to parse HTTP body, make sure to send a properly formatted form request body.").WithWrap(err)
	}
	authorizeRequest.Form = form

	client, err := o.store.GetClient(ctx, form.Get("client_id"))
	if err != nil {
		return nil, ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.").WithWrap(err).WithDebug(err.Error())
	}
	authorizeRequest.Client = client

	authorizeRequest.State = form.Get("state")
	authorizeRequest.CodeChallenge = form.Get("code_challenge")
	authorizeRequest.CodeChallengeMethod = form.Get("code_challenge_method")

	// Since the /authorize endpoint is now only used for the authorization code grant type, we can safely assume
	// that the response type is always "query". For other grant types, the default response mode is "fragment".
	authorizeRequest.DefaultResponseMode = ResponseModeQuery

	if err = validateOpenIDConnectAuthorizeRequest(req, authorizeRequest); err != nil {
		return nil, err
	}

	if err = parseAndValidateResponseMode(req, authorizeRequest); err != nil {
		return nil, err
	}

	if err = parseAndValidateAuthorizeScope(req, authorizeRequest); err != nil {
		return nil, err
	}

	if len(authorizeRequest.State) < o.config.GetMinParameterEntropy() {
		return nil, ErrInvalidState.WithHint("Request parameter 'state' must be at least be %d characters long to ensure sufficient entropy.", o.config.GetMinParameterEntropy())
	}

	return authorizeRequest, nil
}

func (o *OAuth2) NewAuthorizeResponse(ctx context.Context, req *AuthorizeRequest, session Session) (*AuthorizeResponse, error) {
	return nil, nil
}

func (o *OAuth2) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, err error) {

}

func (o *OAuth2) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, resp *AuthorizeResponse) {
}
