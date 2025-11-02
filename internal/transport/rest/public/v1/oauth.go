package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/domain"
	"github.com/tuanta7/hydros/internal/usecase/jwk"
	"github.com/tuanta7/hydros/pkg/zapx"
	"go.uber.org/zap"
)

type OAuthHandler struct {
	cfg    *config.Config
	oauth2 core.OAuth2Provider
	jwkUC  *jwk.UseCase
	logger *zapx.Logger
}

func NewOAuthHandler(cfg *config.Config, oauth2 core.OAuth2Provider, jwkUC *jwk.UseCase, logger *zapx.Logger) *OAuthHandler {
	return &OAuthHandler{
		cfg:    cfg,
		oauth2: oauth2,
		jwkUC:  jwkUC,
		logger: logger,
	}
}

func (h *OAuthHandler) HandleAuthorizeRequest(c *gin.Context) {
	ctx := c.Request.Context()
	authorizeRequest, err := h.oauth2.NewAuthorizeRequest(ctx, c.Request)
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Writer, authorizeRequest, err)
		return
	}

	authorizeResponse, err := h.oauth2.NewAuthorizeResponse(ctx, authorizeRequest, nil)
	if err != nil {
		h.oauth2.WriteAuthorizeError(ctx, c.Writer, authorizeRequest, err)
		return
	}

	h.oauth2.WriteAuthorizeResponse(ctx, c.Writer, authorizeRequest, authorizeResponse)
}

func (h *OAuthHandler) HandleTokenRequest(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.NewSession("")
	tokenRequest, err := h.oauth2.NewTokenRequest(ctx, c.Request, session)
	if err != nil {
		h.logger.Error("error validating token request",
			zap.Error(err),
			zap.String("method", "oauth2.NewTokenRequest"),
		)
		h.oauth2.WriteTokenError(ctx, c.Writer, tokenRequest, err)
		return
	}

	if tokenRequest.GrantType.ExactOne(string(core.GrantTypeClientCredentials)) {
		session.Subject = tokenRequest.Client.GetID()

		// TODO: Do we need to let client to set which type of token it wants?
		if h.cfg.GetAccessTokenFormat() == "jwt" {
			key, err := h.jwkUC.GetOrCreateJWKFn(domain.AccessTokenSet)(ctx)
			if err != nil {
				h.logger.Error("error getting jwk",
					zap.Error(err),
					zap.String("method", "jwk.GetOrCreateJWKFn"),
				)
				h.oauth2.WriteTokenError(ctx, c.Writer, tokenRequest, err)
				return
			}

			session.KeyID = key.(*jose.JSONWebKey).KeyID
		}
	}

	tokenRequest.GrantedScope = tokenRequest.GrantedScope.Append(tokenRequest.Scope...)
	tokenRequest.GrantedAudience = tokenRequest.GrantedAudience.Append(tokenRequest.Audience...)

	session.ClientID = tokenRequest.Client.GetID()
	session.IDTokenSession.Claims.Issuer = h.cfg.GetAccessTokenIssuer()
	session.IDTokenSession.Claims.IssuedAt = x.NowUTC()

	// TODO: Implement rfc8693 token exchange

	tokenResponse, err := h.oauth2.NewTokenResponse(ctx, tokenRequest)
	if err != nil {
		h.logger.Error("error populating token response",
			zap.Error(err),
			zap.String("method", "oauth2.NewTokenResponse"),
		)
		h.oauth2.WriteTokenError(ctx, c.Writer, tokenRequest, err)
		return
	}

	h.oauth2.WriteTokenResponse(ctx, c.Writer, tokenRequest, tokenResponse)
}

func (h *OAuthHandler) HandleIntrospectionRequest(c *gin.Context) {
	ctx := c.Request.Context()
	session := domain.NewSession("")

	resp, err := h.oauth2.IntrospectToken(ctx, c.Request, session)
	if err != nil {
		h.logger.Error("error while introspecting token",
			zap.Error(err),
			zap.String("method", "oauth2.IntrospectToken"),
		)
		h.oauth2.WriteIntrospectionError(ctx, c.Writer, err)
		return
	}

	h.oauth2.WriteIntrospectionResponse(ctx, c.Writer, resp)
}
