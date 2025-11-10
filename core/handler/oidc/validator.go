package oidc

import (
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/tuanta7/hydros/core"

	"github.com/tuanta7/hydros/core/strategy"
	"github.com/tuanta7/hydros/core/x"
)

var defaultPrompts = []string{"login", "none", "consent", "select_account"}

type OpenIDConnectPromptConfigurator interface {
	core.AllowedPromptsProvider
	core.RedirectSecureCheckerProvider
}

func validatePrompt(
	cfg OpenIDConnectPromptConfigurator,
	ar *core.AuthorizeRequest,
	idTokenStrategy strategy.OpenIDConnectTokenStrategy,
) error {
	if ar == nil {
		return errors.New("authorize request cannot be nil")
	}

	prompts := x.SplitSpace(ar.Form.Get("prompt"))
	if ar.Client.IsPublic() {
		uriChecker := cfg.GetRedirectSecureChecker()
		if slices.Contains(prompts, "none") {
			if !uriChecker(ar.RedirectURI) {
				return core.ErrConsentRequired.WithHint("OAuth 2.0 Client is marked public and redirect uri is not considered secure (https missing), but \"prompt=none\" was requested.")
			}
		}
	}

	availablePrompts := cfg.GetAllowedPrompts()
	if len(availablePrompts) == 0 {
		availablePrompts = defaultPrompts
	}

	for _, p := range prompts {
		if !slices.Contains(availablePrompts, p) {
			return core.ErrInvalidRequest.WithHint("Used unknown value '%s' for prompt parameter", p)
		}
	}

	if slices.Contains(prompts, "none") && len(prompts) > 1 {
		return core.ErrInvalidRequest.WithHint("Parameter 'prompt' was set to 'none', but contains other values as well which is not allowed.")
	}

	oidcSession, ok := ar.Session.(OpenIDConnectSession)
	if !ok {
		return core.ErrServerError.WithDebug("Failed to validate OpenID Connect request because session is not of type OpenIDConnectSession.")
	}

	claims := oidcSession.IDTokenClaims()
	if claims.Subject == "" {
		return core.ErrServerError.WithDebug("Failed to validate OpenID Connect request because session subject is empty.")
	}

	// add 5 seconds to account for clock skew
	if claims.AuthTime.After(x.NowUTC().Add(5 * time.Second)) {
		return core.ErrServerError.WithDebug("Failed to validate OpenID Connect request because authentication time is in the future.")
	}

	if slices.Contains(prompts, "login") && claims.AuthTime.Before(claims.RequestedAt) {
		return core.ErrLoginRequired.WithHint("Failed to validate OpenID Connect request because prompt was set to 'login' but auth_time ('%s') happened before the authorization request ('%s') was registered, indicating that the user was not re-authenticated which is forbidden.", claims.AuthTime, claims.RequestedAt)
	}

	if slices.Contains(prompts, "none") {
		if claims.AuthTime.IsZero() {
			return core.ErrServerError.WithDebug("Failed to validate OpenID Connect request because because authentication time is missing from session.")
		}
		if !claims.AuthTime.Equal(claims.RequestedAt) && claims.AuthTime.After(claims.RequestedAt) {
			return core.ErrLoginRequired.WithHint("Failed to validate OpenID Connect request because prompt was set to 'none' but auth_time ('%s') happened after the authorization request ('%s') was registered, indicating that the user was logged in during this request which is not allowed.", claims.AuthTime, claims.RequestedAt)
		}
	}

	maxAge, err := strconv.ParseInt(ar.Form.Get("max_age"), 10, 64)
	if err != nil {
		maxAge = 0
	}

	if maxAge > 0 {
		if claims.AuthTime.IsZero() {
			return core.ErrServerError.WithDebug("Failed to validate OpenID Connect request because authentication time claim is required when max_age is set.")
		} else if claims.RequestedAt.IsZero() {
			return core.ErrServerError.WithDebug("Failed to validate OpenID Connect request because requested time claim is required when max_age is set.")
		} else if claims.AuthTime.Add(time.Duration(maxAge) * time.Second).Before(claims.RequestedAt) {
			return core.ErrLoginRequired.WithDebug("Failed to validate OpenID Connect request because authentication time does not satisfy max_age time.")
		}
	}

	// TODO: support token hint
	return nil
}
