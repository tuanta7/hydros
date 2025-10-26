package oauth

import (
	"context"
	"time"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/storage"
	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/pkg/timex"
)

type ClientCredentialsGrantConfig interface {
	core.AccessTokenLifetimeProvider
}

type ClientCredentialsGrantHandler struct {
	config              ClientCredentialsGrantConfig
	scopeStrategy       strategy.ScopeStrategy
	audienceStrategy    strategy.AudienceStrategy
	accessTokenStrategy strategy.TokenStrategy
	accessTokenStorage  storage.AccessTokenStorage
}

func NewClientCredentialsGrantHandler(
	config ClientCredentialsGrantConfig,
	accessTokenStrategy strategy.TokenStrategy,
	storage storage.AccessTokenStorage,
) *ClientCredentialsGrantHandler {
	return &ClientCredentialsGrantHandler{
		config:              config,
		accessTokenStorage:  storage,
		accessTokenStrategy: accessTokenStrategy,
		scopeStrategy:       strategy.ExactScopeStrategy,
		audienceStrategy:    strategy.ExactAudienceStrategy,
	}
}

func (h *ClientCredentialsGrantHandler) HandleTokenRequest(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
) error {
	if !req.GrantType.ExactOne("client_credentials") {
		return core.ErrUnknownRequest
	}

	client := req.Client
	if client == nil {
		return core.ErrUnauthorizedClient
	}

	if !client.GetGrantTypes().IncludeOne("client_credentials") {
		return core.ErrUnauthorizedClient
	}

	for _, scope := range req.RequestedScope {
		if !h.scopeStrategy(client.GetScopes(), scope) {
			return core.ErrInvalidScope
		}
	}

	err := h.audienceStrategy(client.GetAudience(), req.RequestedAudience)
	if err != nil {
		return err
	}

	if client.IsPublic() {
		return core.ErrInvalidGrant
	}

	accessTokenLifetime := h.config.GetAccessTokenLifetime()
	req.Session.SetExpiresAt(core.AccessToken, timex.NowUTC().Add(accessTokenLifetime))

	err = h.issueToken(ctx, req, res, accessTokenLifetime)
	if err != nil {
		return err
	}

	return nil
}

func (h *ClientCredentialsGrantHandler) issueToken(
	ctx context.Context,
	req *core.TokenRequest,
	res *core.TokenResponse,
	accessTokenLifetime time.Duration,
) error {
	token, signature, err := h.accessTokenStrategy.Generate(req)
	if err != nil {
		return err
	}

	err = h.accessTokenStorage.CreateAccessTokenSession(ctx, signature, req)
	if err != nil {
		return err
	}

	res.AccessToken = token
	res.TokenType = core.BearerToken
	res.ExpiresIn = int64(accessTokenLifetime.Seconds())
	res.Scope = req.GrantedScope
	return nil
}
