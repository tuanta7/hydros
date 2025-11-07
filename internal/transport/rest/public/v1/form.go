package v1

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/config"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/client"
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

func NewFormHandler(cfg *config.Config, uc *flow.UseCase) *FormHandler {
	return &FormHandler{
		cfg:    cfg,
		flowUC: uc,
	}
}

func (h *FormHandler) LoginPage(c *gin.Context) {
	f, returned := h.getLoginFlow(c)
	if returned {
		return
	}

	if !f.LoginSkip {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"LoginChallenge": c.Query("login_challenge"),
			"CSRFToken":      f.LoginCSRF,
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
	password := c.PostForm("password")

	// TODO: implement login strategies
	if email != "admin@example.com" || password != "password" {
		h.writeFormError(c, core.ErrRequestUnauthorized.WithHint("Invalid username or password"))
		return
	}

	// If we use another standalone login provider, that IDP must call the /login/accept API endpoint internally to
	// accept the login and redirect back to the authorization endpoint with the login verifier.
	rememberForDays, _ := strconv.ParseInt(c.PostForm("remember_for"), 10, 64)
	h.acceptLogin(c, &flow.HandledLoginRequest{
		Subject:     email,
		Remember:    c.PostForm("remember") == "on",
		RememberFor: int(rememberForDays) * 24 * 60 * 60, // to seconds
	})
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
		c.JSON(http.StatusGone, gin.H{
			"redirect_to": f.RequestURL, // authorize request
		})
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
	req := c.Request.URL
	urlx.AppendQuery(req, url.Values{
		"error":       []string{err.Error()},
		"description": []string{err.DescriptionField},
	})

	if h.cfg.IsDebugging() {
		urlx.AppendQuery(req, url.Values{
			"debug": []string{err.DebugField},
			"hint":  []string{err.HintField},
		})
	}

	c.Redirect(http.StatusSeeOther, req.String())
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
	c.HTML(http.StatusOK, "consent.html", nil)
}
