package flow

import (
	"encoding/json"
	stderr "errors"
	"fmt"
	"time"

	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/errors"
	"github.com/tuanta7/hydros/pkg/dbtype"
)

func (f *Flow) HandleLoginRequest(h *HandledLoginRequest) error {
	if f.LoginWasHandled {
		return errors.ErrConflict.WithHint("The login request was already used and can no longer be changed.")
	}

	if f.State != FlowStateLoginInitialized && f.State != FlowStateLoginAuthenticated && f.State != FlowStateLoginError {
		return fmt.Errorf("invalid flow state: expected %d/%d/%d, got %d", FlowStateLoginInitialized, FlowStateLoginAuthenticated, FlowStateLoginError, f.State)
	}

	if f.Subject != "" && h.Subject != "" && f.Subject != h.Subject {
		return fmt.Errorf("flow Subject %s does not match the HandledLoginRequest Subject %s", f.Subject, h.Subject)
	}

	if h.Error != nil {
		f.State = FlowStateLoginError
	} else {
		f.State = FlowStateLoginAuthenticated
	}

	if f.Context != nil {
		f.Context = h.Context
	}

	f.AMR = h.AMR
	f.ACR = h.ACR
	f.LoginError = h.Error
	f.Subject = h.Subject
	f.IdentityProviderSessionID = dbtype.NullString(h.IdentityProviderSessionID)
	f.LoginRememberFor = h.RememberFor
	f.LoginRemember = h.Remember

	if f.LoginSkip {
		// if the user skipped the login, it means that the login session is remembered
		f.LoginRemember = true
	} else {
		// if the user did not skip the login, we can set the authenticated_at time to the current time
		f.LoginAuthenticatedAt = dbtype.NullTime(x.NowUTC().Truncate(time.Second))
	}

	return nil
}

func (f *Flow) InvalidateLoginRequest() error {
	if f.State != FlowStateLoginAuthenticated && f.State != FlowStateLoginError {
		return fmt.Errorf("invalid flow state: expected %d or %d, got %d", FlowStateLoginAuthenticated, FlowStateLoginError, f.State)
	}

	if f.LoginWasHandled {
		return stderr.New("login verifier has already been handled")
	}

	f.LoginWasHandled = true
	f.State = FlowStateLoginHandled
	return nil
}

type HandledLoginRequest struct {
	ACR                       string              `json:"acr" form:"acr"`
	AMR                       dbtype.StringArray  `json:"amr" form:"amr"`
	Remember                  bool                `json:"remember" form:"remember"`
	RememberFor               int                 `json:"remember_for" form:"remember_for"`
	ExtendSessionLifespan     bool                `json:"extend_session_lifespan" form:"extend_session_lifespan"`
	Subject                   string              `json:"subject" form:"subject"`
	Error                     *RequestDeniedError `json:"-"` // populated by the reject handler
	IdentityProviderSessionID string              `json:"identity_provider_session_id,omitempty" form:"identity_provider_session_id"`
	Context                   json.RawMessage     `json:"context" form:"context"`
}
