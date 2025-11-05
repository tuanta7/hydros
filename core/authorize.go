package core

import (
	"context"
	"errors"
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
	Prompt              string       `json:"prompt" form:"prompt"`
	Nonce               string       `json:"nonce" form:"nonce"`
	MaxAge              int64        `json:"max_age" form:"max_age"`
	Request
}

func (r *AuthorizeRequest) IsRedirectURIValid() bool {
	if r.RedirectURI == nil {
		return false
	}

	if r.Client == nil {
		return false
	}

	redirectURI, err := x.MatchRedirectURI(r.RedirectURI.String(), r.Client.GetRedirectURIs())
	if err != nil {
		return false
	}

	return x.IsValidRedirectURI(redirectURI.String())
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
	Code   string `json:"code" form:"code"`
	State  string `json:"state" form:"state"`
	Scope  string `json:"scope" form:"scope"`
	Issuer string `json:"issuer" form:"iss"`
}

func NewAuthorizeResponse() *AuthorizeResponse {
	return &AuthorizeResponse{}
}

func (o *OAuth2) NewAuthorizeRequest(ctx context.Context, req *http.Request) (*AuthorizeRequest, error) {
	authorizeRequest := NewAuthorizeRequest()

	form, err := x.BindForm(req)
	if err != nil {
		return authorizeRequest, ErrInvalidRequest.WithHint("Unable to parse HTTP body, make sure to send a properly formatted form request body.").WithWrap(err)
	}
	authorizeRequest.Form = form

	authorizeRequest.ResponseMode, err = parseResponseMode(authorizeRequest)
	if err != nil {
		return authorizeRequest, ErrUnsupportedResponseMode.WithHint("Request with unsupported response_mode \"%s\".", form.Get("response_mode")).WithWrap(err)
	}

	if authorizeRequest.ResponseMode == ResponseModeDefault {
		// Since the /authorize endpoint is now only used for the authorization code grant type, we can safely assume
		// that the response type is always "query". For other grant types, the default response mode is "fragment".
		authorizeRequest.DefaultResponseMode = ResponseModeQuery
	}

	client, err := o.store.GetClient(ctx, form.Get("client_id"))
	if err != nil {
		return authorizeRequest, ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.").WithWrap(err).WithDebug(err.Error())
	}
	authorizeRequest.Client = client

	redirectURI, err := o.parseRedirectURI(authorizeRequest, client.GetRedirectURIs())
	if err != nil {
		return authorizeRequest, err
	}
	authorizeRequest.RedirectURI = redirectURI

	if len(form.Get("registration")) > 0 {
		return authorizeRequest, ErrRegistrationNotSupported
	}

	authorizeRequest.State = form.Get("state")
	if len(authorizeRequest.State) < o.config.GetMinParameterEntropy() {
		return authorizeRequest, ErrInvalidState.WithHint("Request parameter 'state' must be at least be %d characters long to ensure sufficient entropy.", o.config.GetMinParameterEntropy())
	}

	authorizeRequest.ResponseTypes = x.SplitSpace(form.Get("response_type"))
	authorizeRequest.Scope = x.SplitSpace(form.Get("scope"))
	authorizeRequest.Audience = getAudience(authorizeRequest)

	authorizeRequest.CodeChallenge = form.Get("code_challenge")
	authorizeRequest.CodeChallengeMethod = form.Get("code_challenge_method")

	for _, th := range o.authorizeHandlers {
		// HandleAuthorizeRequest only verifies the minimum requirements for the request to avoid overhead check before
		// the login step, the rest of the checks are done after logging in.
		he := th.HandleAuthorizeRequest(ctx, authorizeRequest)
		if he != nil {
			return authorizeRequest, he
		}
	}

	return authorizeRequest, nil
}

func (o *OAuth2) parseRedirectURI(authorizeRequest *AuthorizeRequest, registeredURIs []string) (*url.URL, error) {
	raw := authorizeRequest.Form.Get("redirect_uri")
	if raw == "" && authorizeRequest.Scope.IncludeAll("openid") {
		return nil, ErrInvalidRequest.WithHint("The 'redirect_uri' parameter is required when using OpenID Connect 1.0.")
	}

	// get redirect uri if exists
	redirectURI, err := x.MatchRedirectURI(raw, registeredURIs)
	if err != nil {
		return nil, ErrInvalidRequest.WithHint("The 'redirect_uri' parameter does not match any of the OAuth 2.0 Client's pre-registered redirect urls.")
	}

	// check if redirect uri is valid
	if !x.IsValidRedirectURI(redirectURI.String()) {
		return nil, ErrInvalidRequest.WithHint("The redirect URI '%s' contains an illegal character (for example #) or is otherwise invalid.", redirectURI.String())
	}

	return redirectURI, nil
}

func parseResponseMode(authorizeRequest *AuthorizeRequest) (ResponseMode, error) {
	raw := authorizeRequest.Form.Get("response_mode")
	switch r := ResponseMode(raw); r {
	case ResponseModeDefault:
		return ResponseModeDefault, nil
	case ResponseModeFragment:
		return ResponseModeFragment, nil
	case ResponseModeQuery:
		return ResponseModeQuery, nil
	case ResponseModeFormPost:
		return ResponseModeFormPost, nil
	default:
		return "", errors.New("invalid response mode")
	}
}

func getAudience(request *AuthorizeRequest) []string {
	audiences := request.Form["audience"]
	if len(audiences) > 1 {
		return x.RemoveEmpty(audiences)
	}

	if len(audiences) == 1 {
		// GET requests format the audience as a space-separated list
		return x.SplitSpace(audiences[0])
	}

	return []string{}
}

func (o *OAuth2) NewAuthorizeResponse(ctx context.Context, req *AuthorizeRequest, session Session) (*AuthorizeResponse, error) {
	response := NewAuthorizeResponse()

	req.Session = session
	for _, th := range o.authorizeHandlers {
		if he := th.HandleAuthorizeResponse(ctx, req, response); he != nil {
			return nil, he
		}
	}

	if req.DefaultResponseMode == ResponseModeFragment && req.ResponseMode == ResponseModeQuery {
		return nil, ErrUnsupportedResponseMode.WithHint("Insecure response_mode '%s' for the response_type '%s'.", req.ResponseMode, req.ResponseTypes)
	}

	return response, nil
}

func (o *OAuth2) WriteAuthorizeError(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, err error) {
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	rfcErr := ErrorToRFC6749Error(err)
	errorsForm := rfcErr.ToValues(o.config.IsDebugging())
	errorsForm.Set("state", req.State)

	if !req.IsRedirectURIValid() {
		return
	}

	req.RedirectURI.Fragment = ""

	var redirectURIString string
	switch req.ResponseMode {
	case ResponseModeFormPost:
		rw.Header().Set("Content-Type", "text/html;charset=UTF-8")

		// TODO: implement form post response mode
		rw.WriteHeader(http.StatusNotImplemented)
		_, _ = rw.Write([]byte("<html><body>Form post response mode not implemented.</body></html>"))

		return
	case ResponseModeFragment:
		redirectURIString = req.RedirectURI.String() + "#" + errorsForm.Encode()
	default: // ResponseModeQuery
		req.RedirectURI.RawQuery = errorsForm.Encode()
		redirectURIString = req.RedirectURI.String()
	}

	rw.Header().Set("Location", redirectURIString)
	rw.WriteHeader(http.StatusSeeOther)
}

func (o *OAuth2) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, resp *AuthorizeResponse) {
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	var redirectURIString string
	switch req.ResponseMode {
	case ResponseModeFormPost:
		rw.Header().Set("Content-Type", "text/html;charset=UTF-8")

		// TODO: implement form post response mode
		rw.WriteHeader(http.StatusNotImplemented)
		_, _ = rw.Write([]byte("<html><body>Form post response mode not implemented.</body></html>"))

		return
	case ResponseModeFragment:
		req.RedirectURI.Fragment = ""
		params := url.Values{}
		params.Set("code", resp.Code)
		params.Set("state", resp.State)
		redirectURIString = req.RedirectURI.String() + "#" + params.Encode()
	default:
		params := url.Values{}
		params.Set("code", resp.Code)
		params.Set("state", resp.State)
		req.RedirectURI.RawQuery = params.Encode()
		redirectURIString = req.RedirectURI.String()
	}

	rw.Header().Set("Location", redirectURIString)
	rw.WriteHeader(http.StatusSeeOther)
}
