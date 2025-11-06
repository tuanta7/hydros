package v1

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/pkg/dbtype"
	"github.com/tuanta7/hydros/pkg/urlx"
)

type FlowHandler struct {
	flowUC *flow.UseCase
}

func NewFlowHandler(uc *flow.UseCase) *FlowHandler {
	return &FlowHandler{
		flowUC: uc,
	}
}

func (h *FlowHandler) GetLoginFlow(c *gin.Context) {
	ctx := c.Request.Context()
	challenge := c.Query("login_challenge")
	if challenge == "" {
		c.JSON(http.StatusBadRequest, core.ErrInvalidRequest.WithHint("Query parameter 'login_challenge' is not defined but should have been."))
		return
	}

	f, err := h.flowUC.GetLoginRequest(ctx, challenge)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if f.LoginWasHandled {
		c.JSON(http.StatusGone, gin.H{
			"redirect_to": f.RequestURL, // authorize request
		})
		return
	}

	if f.RequestedScope == nil {
		f.RequestedScope = []string{}
	}

	if f.RequestedScope == nil {
		f.RequestedScope = []string{}
	}

	f.Client = client.SanitizeClient(f.Client)
	c.JSON(http.StatusOK, gin.H{
		"flow": f,
	})
}

func (h *FlowHandler) AcceptLogin(c *gin.Context) {
	ctx := c.Request.Context()

	challenge := c.Query("login_challenge")
	if challenge == "" {
		c.JSON(http.StatusBadRequest, core.ErrInvalidRequest.WithHint("Query parameter 'login_challenge' is not defined but should have been."))
		return
	}

	var handledLoginRequest *flow.HandledLoginRequest
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&handledLoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, core.ErrInvalidRequest.WithHint("Unable to decode body: %s", err).WithWrap(err))
		return
	}

	if handledLoginRequest.Subject == "" {
		c.JSON(http.StatusBadRequest, core.ErrInvalidRequest.WithHint("Field 'subject' must not be empty."))
		return
	}

	f, err := h.flowUC.GetLoginRequest(ctx, challenge)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	} else if f.Subject != "" && f.Subject != handledLoginRequest.Subject {
		// if the flow already has a subject from a remembered login, we redirect the user back to the original
		// authorization request with "prompt=login" to force the user to log in again.
		redirectTo, err := urlx.AppendQueryString(f.RequestURL, url.Values{"prompt": []string{"login"}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, core.ErrServerError.WithWrap(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"redirect_to": redirectTo,
		})
		return
	}

	if f.LoginSkip {
		f.LoginRemember = true
	} else {
		f.LoginAuthenticatedAt = dbtype.NullTime(x.NowUTC().Truncate(time.Second))
	}

	f.Subject = handledLoginRequest.Subject
	f.LoginWasHandled = true
	f.AMR = handledLoginRequest.AMR
	f.ACR = handledLoginRequest.ACR
	f.LoginRememberFor = handledLoginRequest.RememberFor

	verifier, err := h.flowUC.EncodeFlow(ctx, f, flow.AsLoginVerifier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, core.ErrServerError.WithWrap(err))
		return
	}

	redirectTo, err := urlx.AppendQueryString(f.RequestURL, url.Values{"login_verifier": []string{verifier}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, core.ErrServerError.WithWrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"redirect_to": redirectTo,
	})

	return

}

func (h *FlowHandler) RejectLogin(c *gin.Context) {
	// ctx := c.Request.Context()
	challenge := c.Query("login_challenge")
	if challenge == "" {
		c.JSON(http.StatusBadRequest, core.ErrInvalidRequest.WithHint("Query parameter 'login_challenge' is not defined but should have been."))
		return
	}
}
