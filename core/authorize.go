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
	Request
	OIDCAuthorizeRequest
	ResponseTypes       Arguments `json:"response_types" form:"response_types"`
	CodeChallenge       string    `json:"code_challenge" form:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method" form:"code_challenge_method"`
	RedirectURI         *url.URL  `json:"redirect_uri" form:"redirect_uri"`
	State               string    `json:"state" form:"state"`
}

type OIDCAuthorizeRequest struct {
	ResponseMode        ResponseMode `json:"response_mode" form:"response_mode"`
	DefaultResponseMode ResponseMode `json:"default_response_mode" form:"-"`
	Prompt              Arguments    `json:"prompt" form:"prompt"`
	Nonce               string       `json:"nonce" form:"nonce"`
	MaxAge              int64        `json:"max_age" form:"max_age"`
}

func NewAuthorizeRequest() *AuthorizeRequest {
	return &AuthorizeRequest{
		ResponseTypes: Arguments{},
		Request:       *NewRequest(),
		RedirectURI:   nil, // must be set to nil for redirect detection to work properly
		OIDCAuthorizeRequest: OIDCAuthorizeRequest{
			ResponseMode: ResponseModeDefault,
		},
	}
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

// NewAuthorizeRequest processes an OAuth2 authorization request and validates parameters format
func (o *OAuth2) NewAuthorizeRequest(ctx context.Context, req *http.Request) (*AuthorizeRequest, error) {
	ar := NewAuthorizeRequest()

	form, err := x.BindForm(req)
	if err != nil {
		return ar, ErrInvalidRequest.
			WithHint("Unable to parse HTTP body, make sure to send a properly formatted form request body.").
			WithWrap(err)
	}
	ar.Form = form

	if len(form.Get("registration")) > 0 {
		return ar, ErrRegistrationNotSupported
	}

	responseMode := ar.Form.Get("response_mode")
	if ar.ResponseMode, err = parseResponseMode(responseMode); err != nil {
		return ar, ErrUnsupportedResponseMode.
			WithHint("Request with unsupported response_mode \"%s\".", responseMode).
			WithWrap(err)
	}

	if ar.ResponseMode == ResponseModeDefault {
		// Since the /authorize endpoint is now only used for the authorization code grant type, we can safely assume
		// that the response type is always "query". For other grant types, the default response mode is "fragment".
		ar.DefaultResponseMode = ResponseModeQuery
	}

	if ar.Client, err = o.store.GetClient(ctx, form.Get("client_id")); err != nil {
		return ar, ErrInvalidClient.
			WithHint("The requested OAuth 2.0 Client does not exist.").
			WithWrap(err).
			WithDebug(err.Error())
	}

	if ar.RedirectURI, err = parseRedirectURI(ar, ar.Client.GetRedirectURIs()); err != nil {
		return ar, err
	}

	if ar.State = form.Get("state"); len(ar.State) < o.config.GetMinParameterEntropy() {
		return ar, ErrInvalidState.
			WithHint("Request parameter 'state' must be at least be %d characters long.", o.config.GetMinParameterEntropy())
	}

	audiences := ar.Form["audience"]
	if len(audiences) > 1 {
		// POST requests format the audience as a JSON array
		ar.RequestedAudience = x.RemoveEmpty(audiences)
	} else if len(audiences) == 1 {
		// GET requests format the audience as a space-separated list
		ar.RequestedAudience = x.SplitSpace(audiences[0])
	} else {
		ar.RequestedAudience = []string{}
	}

	ar.RequestedScope = x.SplitSpace(form.Get("scope"))
	ar.ResponseTypes = x.SplitSpace(form.Get("response_type"))
	ar.CodeChallenge = form.Get("code_challenge")
	ar.CodeChallengeMethod = form.Get("code_challenge_method")

	for _, th := range o.authorizeHandlers {
		// HandleAuthorizeRequest only verifies the minimum requirements for the request to avoid overhead check before
		// the login step, the rest of the checks are done after logging in.
		he := th.HandleAuthorizeRequest(ctx, ar)
		if he != nil {
			return ar, he
		}
	}

	return ar, nil
}

func parseRedirectURI(ar *AuthorizeRequest, registeredURIs []string) (*url.URL, error) {
	raw := ar.Form.Get("redirect_uri")
	if raw == "" && ar.RequestedScope.IncludeAll("openid") {
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

func parseResponseMode(rm string) (ResponseMode, error) {
	switch r := ResponseMode(rm); r {
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

type AuthorizeResponse struct {
	Code   string `json:"code" form:"code"`
	State  string `json:"state" form:"state"`
	Issuer string `json:"issuer" form:"iss"`
}

func NewAuthorizeResponse() *AuthorizeResponse {
	return &AuthorizeResponse{}
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

func (o *OAuth2) WriteAuthorizeResponse(ctx context.Context, rw http.ResponseWriter, req *AuthorizeRequest, resp *AuthorizeResponse) {
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	var redirectURIString string
	switch req.ResponseMode {
	case ResponseModeFormPost:
		rw.Header().Set("Content-Type", "text/html;charset=UTF-8")
		o.FormPostResponse(req.RedirectURI.String(), rw)
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
		o.FormPostResponse(req.RedirectURI.String(), rw)
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
