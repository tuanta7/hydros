package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/jwk"
	"github.com/tuanta7/hydros/internal/session"
	"go.uber.org/zap"
)

func (h *OAuthHandler) HandleTokenRequest(c *gin.Context) {
	ctx := c.Request.Context()
	s := session.NewSession("")
	tokenRequest, err := h.oauth2.NewTokenRequest(ctx, c.Request, s)
	if err != nil {
		h.logger.Error("error validating token request",
			zap.Error(err),
			zap.String("method", "oauth2.NewTokenRequest"),
		)
		h.oauth2.WriteTokenError(ctx, c.Writer, tokenRequest, err)
		return
	}

	if tokenRequest.GrantType.ExactOne(string(core.GrantTypeClientCredentials)) {
		s.Claims.Subject = tokenRequest.Client.GetID()

		// TODO: Do we need to let client to set which type of token it wants?
		if h.cfg.GetAccessTokenFormat() == "jwt" {
			key, err := h.jwkUC.GetOrCreateJWKFn(jwk.AccessTokenSet)(ctx)
			if err != nil {
				h.logger.Error("error getting jwk",
					zap.Error(err),
					zap.String("method", "jwk.GetOrCreateJWKFn"),
				)
				h.oauth2.WriteTokenError(ctx, c.Writer, tokenRequest, err)
				return
			}

			s.KeyID = key.(*jose.JSONWebKey).KeyID
		}
	}

	tokenRequest.GrantedScope = tokenRequest.GrantedScope.Append(tokenRequest.RequestedScope...)
	tokenRequest.GrantedAudience = tokenRequest.GrantedAudience.Append(tokenRequest.RequestedAudience...)

	s.ClientID = tokenRequest.Client.GetID()
	s.IDTokenSession.Claims.Issuer = h.cfg.GetAccessTokenIssuer()
	s.IDTokenSession.Claims.IssuedAt = jwt.NewNumericDate(x.NowUTC())

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
