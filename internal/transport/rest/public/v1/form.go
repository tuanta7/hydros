package v1

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/internal/login"
	"github.com/tuanta7/hydros/pkg/helper"
	"github.com/tuanta7/hydros/pkg/urlx"
)

// FormHandler is used to handle login and consent pages. It mimics the behavior of the
// public APIs.
type FormHandler struct {
	cfg    *config.Config
	idp    []login.IdentityProvider
	flowUC *flow.UseCase
}

func NewFormHandler(cfg *config.Config, uc *flow.UseCase, idp ...login.IdentityProvider) *FormHandler {
	return &FormHandler{
		cfg:    cfg,
		flowUC: uc,
		idp:    idp,
	}
}

func (h *FormHandler) LoginPage(c *gin.Context) {
	f, returned := h.getLoginFlow(c)
	if returned {
		return
	}

	if !f.LoginSkip {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"LoginChallenge":   c.Query("login_challenge"),
			"CSRFToken":        f.LoginCSRF,
			"Error":            c.Query("error"),
			"ErrorDescription": c.Query("error_description"),
			"Hint":             c.Query("hint"),
		})
		return
	}

	h.acceptLogin(c, &flow.HandledLoginRequest{
		Subject: f.Subject,
	})
}

func (h *FormHandler) Login(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	email := c.PostForm("email")
	for _, i := range h.idp {
		err = i.Login(c.Request.Context(), &login.Credentials{
			Username: email,
			Password: c.PostForm("password"),
		})

		if err == nil {
			// If we use another standalone login provider, that IDP must call the /login/accept API endpoint internally to
			// accept the login and redirect back to the authorization endpoint with the login verifier.
			rememberForDays, _ := strconv.ParseInt(c.PostForm("remember_for"), 10, 64)
			h.acceptLogin(c, &flow.HandledLoginRequest{
				Subject:     c.PostForm("email"),
				Remember:    c.PostForm("remember") == "on",
				RememberFor: int(rememberForDays) * 24 * 60 * 60, // to seconds
			})
			return
		}
	}

	h.writeFormError(c, core.ErrRequestUnauthorized.WithHint("Invalid username or password"))
}

func (h *FormHandler) getLoginFlow(c *gin.Context) (*flow.Flow, bool) {
	ctx := c.Request.Context()

	challenge := c.Query("login_challenge")
	if challenge == "" {
		c.JSON(http.StatusBadRequest, core.ErrInvalidRequest.WithHint("Query parameter 'login_challenge' is not defined but should have been."))
		return nil, true
	}

	f, err := h.flowUC.GetLoginRequest(ctx, challenge)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return nil, true
	}

	if f.LoginWasHandled {
		// redirect to the original authorization request
		c.Redirect(http.StatusGone, f.RequestURL)
		return nil, true
	}

	if f.RequestedScope == nil {
		f.RequestedScope = []string{}
	}

	if f.RequestedScope == nil {
		f.RequestedScope = []string{}
	}

	f.Client = client.SanitizeClient(f.Client)
	return f, false
}

func (h *FormHandler) writeFormError(c *gin.Context, err *core.RFC6749Error) {
	formURL := c.Request.URL

	formURL = urlx.AppendQuery(formURL, url.Values{
		"login_challenge":   []string{c.PostForm("login_challenge")},
		"consent_challenge": []string{c.PostForm("consent_challenge")},
		"error":             []string{err.Error()},
		"error_description": []string{err.DescriptionField},
	})

	if h.cfg.IsDebugging() {
		formURL = urlx.AppendQuery(formURL, url.Values{
			"debug": []string{err.DebugField},
			"hint":  []string{err.HintField},
		})
	}

	c.Redirect(http.StatusSeeOther, formURL.String())
}

func (h *FormHandler) acceptLogin(c *gin.Context, handledLoginRequest *flow.HandledLoginRequest) {
	ctx := c.Request.Context()

	challenge := helper.StringCoalesce(
		c.Query("login_challenge"),    // skip login
		c.PostForm("login_challenge"), // form value
	)
	if challenge == "" {
		h.writeFormError(c, core.ErrInvalidRequest.WithHint("'login_challenge' is not defined but should have been."))
		return
	}

	if handledLoginRequest.Subject == "" {
		h.writeFormError(c, core.ErrInvalidRequest.WithHint("Field 'subject' must not be empty."))
		return
	}

	f, err := h.flowUC.GetLoginRequest(ctx, challenge)
	if err != nil {
		h.writeFormError(c, core.ErrInvalidRequest.WithWrap(err))
		return
	} else if f.Subject != "" && f.Subject != handledLoginRequest.Subject {
		// if the flow already has a subject from a remembered login, we redirect the user back to the original
		// authorization request with "prompt=login" to force the user to log in again.
		redirectTo, err := urlx.AppendQueryString(f.RequestURL, url.Values{"prompt": []string{"login"}})
		if err != nil {
			h.writeFormError(c, core.ErrorToRFC6749Error(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"redirect_to": redirectTo,
		})
		return
	}

	err = f.HandleLoginRequest(handledLoginRequest)
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	verifier, err := h.flowUC.EncodeFlow(ctx, f, flow.AsLoginVerifier)
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	redirectTo, err := urlx.AppendQueryString(f.RequestURL, url.Values{"login_verifier": []string{verifier}})
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	c.Redirect(http.StatusSeeOther, redirectTo)
}

func (h *FormHandler) ConsentPage(c *gin.Context) {
	f, returned := h.getConsentFlow(c)
	if returned {
		return
	}

	if !f.ConsentSkip {
		c.HTML(http.StatusOK, "consent.html", gin.H{
			"ConsentChallenge": c.Query("consent_challenge"),
			"CSRFToken":        f.ConsentCSRF,
			"ClientName":       f.Client.Name,
			"Scopes":           f.RequestedScope,
			"Audiences":        f.RequestedAudience,
		})
		return
	}

	h.acceptConsent(c, &flow.HandledConsentRequest{
		GrantedScope:    f.GrantedScope,
		GrantedAudience: f.GrantedAudience,
		Remember:        f.ConsentRemember,
	})
}

func (h *FormHandler) Consent(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	actions := c.PostForm("action")
	if actions == "deny" {
		h.rejectConsent(c, &flow.RequestDeniedError{
			Error:            "scope_grant_denied",
			ErrorDescription: "The resource owner denied the request",
		})
		return
	}

	rememberForDays, _ := strconv.ParseInt(c.PostForm("remember_for"), 10, 64)
	h.acceptConsent(c, &flow.HandledConsentRequest{
		GrantedScope:    c.PostFormArray("scopes"),
		GrantedAudience: c.PostFormArray("audiences"),
		Remember:        c.PostForm("remember") == "on",
		RememberFor:     int(rememberForDays) * 24 * 60 * 60, // to seconds
	})
}

func (h *FormHandler) getConsentFlow(c *gin.Context) (*flow.Flow, bool) {
	ctx := c.Request.Context()

	challenge := c.Query("consent_challenge")
	if challenge == "" {
		c.JSON(http.StatusBadRequest, core.ErrInvalidRequest.WithHint("Query parameter 'consent_challenge' is not defined but should have been."))
		return nil, true
	}

	f, err := h.flowUC.GetConsentRequest(ctx, challenge)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return nil, true
	}

	if f.ConsentWasHandled {
		// redirect to the original authorization request
		c.Redirect(http.StatusGone, f.RequestURL)
		return nil, true
	}

	if f.RequestedScope == nil {
		f.RequestedScope = []string{}
	}

	if f.RequestedScope == nil {
		f.RequestedScope = []string{}
	}

	f.Client = client.SanitizeClient(f.Client)
	return f, false
}

func (h *FormHandler) acceptConsent(c *gin.Context, handledConsentRequest *flow.HandledConsentRequest) {
	ctx := c.Request.Context()

	challenge := helper.StringCoalesce(
		c.Query("consent_challenge"),    // skip consent
		c.PostForm("consent_challenge"), // form value
	)
	if challenge == "" {
		h.writeFormError(c, core.ErrInvalidRequest.WithHint("'consent_challenge' is not defined but should have been."))
		return
	}

	f, err := h.flowUC.GetConsentRequest(ctx, challenge)
	if err != nil {
		h.writeFormError(c, core.ErrInvalidRequest.WithWrap(err))
		return
	}

	err = f.HandleConsentRequest(handledConsentRequest)
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	verifier, err := h.flowUC.EncodeFlow(ctx, f, flow.AsConsentVerifier)
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	redirectTo, err := urlx.AppendQueryString(f.RequestURL, url.Values{"consent_verifier": []string{verifier}})
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	c.Redirect(http.StatusSeeOther, redirectTo)
}

func (h *FormHandler) rejectConsent(c *gin.Context, deniedErr *flow.RequestDeniedError) {
	ctx := c.Request.Context()

	challenge := helper.StringCoalesce(
		c.Query("consent_challenge"),    // skip consent
		c.PostForm("consent_challenge"), // form value
	)
	if challenge == "" {
		h.writeFormError(c, core.ErrInvalidRequest.WithHint("'consent_challenge' is not defined but should have been."))
		return
	}

	f, err := h.flowUC.GetConsentRequest(ctx, challenge)
	if err != nil {
		h.writeFormError(c, core.ErrInvalidRequest.WithWrap(err))
		return
	}

	deniedErr.Valid = true
	deniedErr.SetDefaults(flow.ConsentRequestDeniedErrorName)

	err = f.HandleConsentRequest(&flow.HandledConsentRequest{
		Error: deniedErr,
	})
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	verifier, err := h.flowUC.EncodeFlow(ctx, f, flow.AsConsentVerifier)
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	redirectTo, err := urlx.AppendQueryString(f.RequestURL, url.Values{"consent_verifier": []string{verifier}})
	if err != nil {
		h.writeFormError(c, core.ErrorToRFC6749Error(err))
		return
	}

	c.Redirect(http.StatusSeeOther, redirectTo)
}
