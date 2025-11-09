package flow

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/pkg/dbtype"
)

func (f *Flow) HandleConsentRequest(h *HandledConsentRequest) error {
	if f.State != StateConsentInitialized && f.State != StateConsentGranted && f.State != StateConsentError {
		return fmt.Errorf("invalid flow state: expected %d/%d/%d, got %d", StateConsentInitialized, StateConsentGranted, StateConsentError, f.State)
	}

	if h.Error != nil {
		f.State = StateConsentError
	} else {
		f.State = StateConsentGranted
	}

	f.GrantedScope = h.GrantedScope
	f.GrantedAudience = h.GrantedAudience
	f.ConsentRemember = h.Remember
	f.ConsentRememberFor = &h.RememberFor
	f.ConsentGrantedAt = dbtype.NullTime(x.NowUTC())
	f.ConsentError = h.Error
	if h.Context != nil {
		f.Context = h.Context
	}

	return nil
}

func (f *Flow) InvalidateConsentRequest() error {
	if f.ConsentWasHandled {
		return errors.New("consent verifier has already been used")
	}
	if f.State != StateConsentGranted && f.State != StateConsentError {
		return fmt.Errorf("unexpected flow state: expected %d or %d, got %d", StateConsentGranted, StateConsentError, f.State)
	}

	f.ConsentWasHandled = true
	f.State = StateConsentHandled
	return nil
}

type HandledConsentRequest struct {
	GrantedScope    dbtype.StringArray  `json:"scope" form:"scope"`
	GrantedAudience dbtype.StringArray  `json:"audience" form:"audience"`
	Remember        bool                `json:"remember" form:"remember"`
	RememberFor     int                 `json:"remember_for" form:"remember_for"`
	Error           *RequestDeniedError `json:"-"`
	Context         json.RawMessage     `json:"context"`
}
