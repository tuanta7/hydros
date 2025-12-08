package oauth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/core/x"
)

type ClientCredentialsGrantConfigurator interface {
	core.AccessTokenLifetimeProvider
	strategy.ScopeStrategyProvider
	strategy.AudienceStrategyProvider
}

type ClientCredentialsGrantHandler struct {
	config              ClientCredentialsGrantConfigurator
	accessTokenStrategy strategy.AccessTokenStrategy
	accessTokenStorage  storage.AccessTokenStorage
}

// NewClientCredentialsGrantHandler returns a new handler with default matching strategies
func NewClientCredentialsGrantHandler(
	config ClientCredentialsGrantConfigurator,
	accessTokenStrategy strategy.TokenStrategy,
	storage storage.AccessTokenStorage,
) *ClientCredentialsGrantHandler {
	return &ClientCredentialsGrantHandler{
		config:              config,
		accessTokenStorage:  storage,
		accessTokenStrategy: accessTokenStrategy,
	}
}

func (h *ClientCredentialsGrantHandler) HandleTokenRequest(ctx context.Context, req *core.TokenRequest) error {
	if !req.GrantType.ExactOne("client_credentials") {
		return core.ErrUnknownRequest
	}

	client := req.Client
	if client == nil {
		// should never happen because this client must be authenticated to get here
		return core.ErrInvalidClient.WithHint("The requested OAuth 2.0 Client does not exist.")
	}

	if !client.GetGrantTypes().IncludeOne("client_credentials") {
		return core.ErrUnauthorizedClient
	}

	scopeStrategy := h.config.GetScopeStrategy()
	if err := scopeStrategy(client.GetScopes(), req.RequestedScope); err != nil {
		return err
	}

	audienceStrategy := h.config.GetAudienceStrategy()
	err := audienceStrategy(client.GetAudience(), req.RequestedAudience)
	if err != nil {
		return err
	}

	if client.IsPublic() {
		return core.ErrInvalidGrant
	}

	if req.Session == nil {
		return errors.New("session cannot be nil")
	}

	accessTokenLifetime := h.config.GetAccessTokenLifetime()
	req.Session.SetExpiresAt(core.AccessToken, x.NowUTC().Add(accessTokenLifetime))
	return nil
}

func (h *ClientCredentialsGrantHandler) HandleTokenResponse(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
) error {
	if !req.GrantType.ExactOne("client_credentials") {
		return core.ErrUnknownRequest
	}

	if !req.Client.GetGrantTypes().IncludeOne("client_credentials") {
		return core.ErrUnauthorizedClient
	}

	token, signature, err := h.accessTokenStrategy.GenerateAccessToken(ctx, &req.Request)
	if err != nil {
		return err
	}

	err = h.accessTokenStorage.CreateAccessTokenSession(ctx, signature, &req.Request)
	if err != nil {
		return err
	}

	accessTokenLifetime := h.config.GetAccessTokenLifetime()
	if req.Session.GetExpiresAt(core.AccessToken).IsZero() {
		res.ExpiresIn = time.Duration(accessTokenLifetime.Seconds())
	} else {
		res.ExpiresIn = x.SecondsFromNow(req.Session.GetExpiresAt(core.AccessToken))
	}

	res.AccessToken = token
	res.TokenType = core.BearerToken
	res.Scope = strings.Join(req.GrantedScope, " ")
	return nil
}
