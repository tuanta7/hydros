package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/tuanta7/hydros/core/x"
)

type IntrospectionRequest struct {
	Token         string    `json:"token" form:"token"`
	TokenTypeHint TokenType `json:"token_type_hint" form:"token_type_hint"`
	Scope         Arguments `json:"scope" form:"scope"`
}

type IntrospectionResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"aud,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	JTI       string `json:"jti,omitempty"`
}

func (o *OAuth2) IntrospectToken(ctx context.Context, req *http.Request, session Session) (*IntrospectionResponse, error) {
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

	_, clientErr := o.AuthenticateClient(ctx, req, form)
	if clientErr != nil {
		return nil, clientErr
	}

	request := &IntrospectionRequest{
		Token:         form.Get("token"),
		TokenTypeHint: TokenType(form.Get("token_type_hint")),
		Scope:         strings.Fields(form.Get("scope")),
	}

	if clientToken := x.AccessTokenFromRequest(req); clientToken != "" {
		if request.Token == clientToken {
			return nil, ErrRequestUnauthorized.WithHint("Bearer and introspection token are identical.")
		}
		// TODO: read fosite/hydra implementation, idk what to do here
	} else {
	}

	handled := false
	tokenType := TokenType("")
	tr := NewTokenRequest(session)
	for _, ih := range o.introspectionHandler {
		tt, ie := ih.IntrospectToken(ctx, request, tr)
		if ie == nil {
			handled = true
			tokenType = tt
		} else if errors.Is(ie, ErrUnknownRequest) {
			// this handler does not handle this token type, try the next one
			continue
		} else if ie != nil {
			return nil, ie
		}
	}

	if !handled {
		return nil, ErrRequestUnauthorized.WithHint("Unable to find a suitable introspection strategy for the token, thus it is invalid.")
	}

	accessTokenType := ""
	if tokenType == AccessToken {
		accessTokenType = BearerToken
	}

	return &IntrospectionResponse{
		Active:    true,
		Scope:     strings.Join(tr.RequestedScope, " "),
		ClientID:  tr.Client.GetID(),
		TokenType: accessTokenType,
		Subject:   tr.Session.GetSubject(),
		Audience:  strings.Join(tr.RequestedAudience, " "),
	}, nil
}

func (o *OAuth2) WriteIntrospectionError(ctx context.Context, rw http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	if !errors.Is(err, ErrInactiveToken) {
		o.writeError(ctx, rw, err)
		return
	}

	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")
	rw.WriteHeader(http.StatusInternalServerError)
	_, _ = rw.Write([]byte(`{"active":"false"}`))
}

func (o *OAuth2) WriteIntrospectionResponse(ctx context.Context, rw http.ResponseWriter, resp *IntrospectionResponse) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")

	if !resp.Active {
		http.Error(rw, `{"active":"false"}`, http.StatusInternalServerError)
	}

	jsonPayload, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonPayload)
}
