package core

import (
	"net/http"
	"strings"

	"github.com/tuanta7/hydros/core/x"
)

func validateOpenIDConnectAuthorizeRequest(r *http.Request, authorizeRequest *AuthorizeRequest) error {
	scope := authorizeRequest.Form.Get("scope")
	scopes := Arguments(x.RemoveEmpty(strings.Split(scope, " ")))
	if !scopes.IncludeOne("openid") {
		// this authorizeRequest is not an OpenID Connect request
		return nil
	}

	requestURI := authorizeRequest.Form.Get("request_uri")
	request := authorizeRequest.Form.Get("request")

	if request == "" && requestURI == "" {
		return ErrInvalidRequest.WithHint("OpenID Connect parameters 'request' and 'request_uri' were both given, but you can use at most one.")
	}

	if request != "" && requestURI != "" {
		return ErrInvalidRequest.WithHint("OpenID Connect parameters 'request' and 'request_uri' were both given, but you can use at most one.")
	}

	oidcClient, ok := authorizeRequest.Client.(OpenIDConnectClient)
	if !ok {
		if requestURI != "" {
			return ErrRequestNotSupported.WithHint("OpenID Connect 'request_uri' context was given, but the  OAuth 2.0 Client does not implement advanced OpenID Connect capabilities.")
		}
		return ErrRequestNotSupported.WithHint("OpenID Connect 'request' context was given, but the  OAuth 2.0 Client does not implement advanced OpenID Connect capabilities.")
	}

	if oidcClient.GetJWKs() == nil || oidcClient.GetJWKsURI() == "" {
		return ErrInvalidRequest.WithHint("OpenID Connect 'request' or 'request_uri' context was given, but the OAuth 2.0 Client does not have any JSON Web Keys registered.")
	}

	return nil
}

func parseAndValidateResponseMode(r *http.Request, request *AuthorizeRequest) error {
	switch responseMode := r.Form.Get("response_mode"); responseMode {
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

	if request.ResponseMode == ResponseModeDefault {
		// the response mode is not set, we need to use the default response mode later
		return nil
	}

	registeredResponseModes := request.Client.GetResponseModes()
	found := false
	for _, mode := range registeredResponseModes {
		if mode == request.ResponseMode {
			found = true
			break
		}
	}
	if !found {
		return ErrUnsupportedResponseMode.WithHint("The client is not allowed to request response_mode '%s'.", r.Form.Get("response_mode"))
	}

	return nil
}

func parseAndValidateAuthorizeScope(r *http.Request, request *AuthorizeRequest) error {
	return nil
}
