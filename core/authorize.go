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
	Code  string `json:"code" form:"code"`
	State string `json:"state" form:"state"`
	Scope string `json:"scope" form:"scope"`
}

func NewAuthorizeResponse() *AuthorizeResponse {
	return &AuthorizeResponse{}
}

func (o *OAuth2) NewAuthorizeRequest(ctx context.Context, req *http.Request) (*AuthorizeRequest, error) {
	authorizeRequest := NewAuthorizeRequest()

	form, err := x.BindForm(req)
	if err != nil {
		return nil, ErrInvalidRequest.WithHint("Unable to parse HTTP body, make sure to send a properly formatted form request body.").WithWrap(err)
	}
	authorizeRequest.Form = form

	client, err := o.store.GetClient(ctx, form.Get("client_id"))
	if err != nil {
		return nil, ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.").WithWrap(err).WithDebug(err.Error())
	}
	authorizeRequest.Client = client

	redirectURI, err := url.Parse(authorizeRequest.Form.Get("redirect_uri"))
	if err != nil {
		return nil, ErrInvalidRequest.WithHint("Invalid redirect_uri \"%s\".", authorizeRequest.Form.Get("redirect_uri")).WithWrap(err)
	}
	authorizeRequest.RedirectURI = redirectURI

	authorizeRequest.State = form.Get("state")
	authorizeRequest.Scope = x.SplitSpace(form.Get("scope"))
	authorizeRequest.ResponseTypes = x.SplitSpace(form.Get("response_type"))
	authorizeRequest.CodeChallenge = form.Get("code_challenge")
	authorizeRequest.CodeChallengeMethod = form.Get("code_challenge_method")

	if err = parseAudience(authorizeRequest); err != nil {
		return nil, err
	}

	if err = parseResponseMode(authorizeRequest); err != nil {
		return nil, err
	}

	if authorizeRequest.ResponseMode == ResponseModeDefault {
		// Since the /authorize endpoint is now only used for the authorization code grant type, we can safely assume
		// that the response type is always "query". For other grant types, the default response mode is "fragment".
		authorizeRequest.DefaultResponseMode = ResponseModeQuery
	}

	if len(form.Get("registration")) > 0 {
		return nil, ErrRegistrationNotSupported
	}

	for _, th := range o.authorizeHandlers {
		// HandleAuthorizeRequest verifies the minimum requirements for the request to avoid overhead check before the
		// login step, the rest of the checks are done after the login step.
		he := th.HandleAuthorizeRequest(ctx, authorizeRequest)
		if he != nil {
			return nil, he
		}
	}

	return authorizeRequest, nil
}

func parseResponseMode(request *AuthorizeRequest) error {
	switch responseMode := request.Form.Get("response_mode"); responseMode {
	case string(ResponseModeDefault):
		request.ResponseMode = ResponseModeDefault
	case string(ResponseModeFragment):
		request.ResponseMode = ResponseModeFragment
	case string(ResponseModeQuery):
		request.ResponseMode = ResponseModeQuery
	case string(ResponseModeFormPost):
		request.ResponseMode = ResponseModeFormPost
	default:
		return ErrUnsupportedResponseMode.WithHint("Request with unsupported response_mode \"%s\".", responseMode)
	}

	return nil
}

func parseAudience(request *AuthorizeRequest) error {
	audiences := request.Form["audience"]
	if len(audiences) > 1 {
		request.Audience = x.RemoveEmpty(audiences)
	} else if len(audiences) == 1 {
		request.Audience = x.SplitSpace(audiences[0])
	} else {
		request.Audience = []string{}
	}
	return nil
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
	errors := rfcErr.ToValues()
	errors.Set("state", req.State)

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
		redirectURIString = req.RedirectURI.String() + "#" + errors.Encode()
	default: // ResponseModeQuery
		for k, v := range errors {
			req.RedirectURI.Query()[k] = v
		}
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
		req.RedirectURI.Query().Set("code", resp.Code)
		req.RedirectURI.Query().Set("state", resp.State)
		redirectURIString = req.RedirectURI.String()
	}

	rw.Header().Set("Location", redirectURIString)
	rw.WriteHeader(http.StatusSeeOther)
}
