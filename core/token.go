package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/tuanta7/hydros/core/x"
)

type TokenRequest struct {
	GrantType        Arguments `json:"grant_type" form:"grant_type"`
	HandledGrantType Arguments `json:"handled_grant_type" form:"handled_grant_type"`
	RedirectURI      string    `json:"redirect_uri" form:"redirect_uri"`
	Code             string    `json:"code" form:"code"` // authorization code
	CodeVerifier     string    `json:"code_verifier" form:"code_verifier"`
	Request
}

func NewTokenRequest(session Session) *TokenRequest {
	r := &TokenRequest{
		GrantType:        Arguments{},
		HandledGrantType: Arguments{},
		Request:          *NewRequest(),
	}
	r.Session = session
	return r
}

type TokenResponse struct {
	AccessToken  string        `json:"access_token" form:"access_token"`
	TokenType    string        `json:"token_type" form:"token_type"`
	ExpiresIn    time.Duration `json:"expires_in,omitempty" form:"expires_in"`
	Scope        Arguments     `json:"scope,omitempty" form:"scope"`
	RefreshToken string        `json:"refresh_token,omitempty" form:"refresh_token"`
	IDToken      string        `json:"id_token,omitempty" form:"id_token"`
}

func NewTokenResponse() *TokenResponse {
	return &TokenResponse{}
}

func (o *OAuth2) NewTokenRequest(ctx context.Context, req *http.Request, session Session) (*TokenRequest, error) {
	if session == nil {
		return nil, errors.New("session must not be nil")
	}

	if req.Method != http.MethodPost {
		return nil, ErrInvalidRequest.WithHint("HTTP method is '%s', expected 'POST'.", req.Method)
	}

	form, err := x.BindPostForm(req)
	if err != nil {
		return nil, ErrInvalidRequest.WithHint("Unable to parse HTTP body, make sure to send a properly formatted form request body.").WithWrap(err)
	} else if len(form) == 0 {
		return nil, ErrInvalidRequest.WithHint("The POST body can not be empty.")
	}

	client, clientErr := o.AuthenticateClient(ctx, req, form)
	if clientErr != nil {
		return nil, clientErr
	}

	tokenRequest := NewTokenRequest(session)
	tokenRequest.Form = form
	tokenRequest.Client = client

	tokenRequest.GrantType = form["grant_type"]
	tokenRequest.RequestedAudience = form["audience"]
	tokenRequest.RequestedScope = form["scope"]

	tokenRequest.Code = form.Get("code")
	tokenRequest.CodeVerifier = form.Get("code_verifier")

	handled := false
	for _, th := range o.tokenHandlers {
		he := th.HandleTokenRequest(ctx, tokenRequest)
		if he == nil {
			handled = true
		} else if errors.Is(he, ErrUnknownRequest) {
			// this handler does not handle this grant type, try the next one
			continue
		} else if he != nil {
			return nil, he
		}
	}

	if !handled {
		return nil, ErrInvalidRequest.WithHint("The requested grant type is not supported by this authorization server.")
	}

	return tokenRequest, nil
}

func (o *OAuth2) NewTokenResponse(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	tokenResponse := NewTokenResponse()

	for _, th := range o.tokenHandlers {
		he := th.HandleTokenResponse(ctx, req, tokenResponse)
		if he == nil || errors.Is(he, ErrUnknownRequest) {
			continue
		} else if he != nil {
			return nil, he
		}
	}

	if tokenResponse.AccessToken == "" || tokenResponse.TokenType == "" {
		return nil, ErrInvalidRequest.
			WithHint("An internal server occurred while trying to complete the request.").
			WithDebug("Access token or token type not set")
	}

	return tokenResponse, nil
}

func (o *OAuth2) WriteTokenError(ctx context.Context, rw http.ResponseWriter, req *TokenRequest, err error) {
	o.writeError(ctx, rw, err)
}

func (o *OAuth2) WriteTokenResponse(ctx context.Context, rw http.ResponseWriter, req *TokenRequest, resp *TokenResponse) {
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	jsonPayload, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonPayload)
}
