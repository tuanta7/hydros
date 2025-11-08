package flow

import (
	"errors"
	"fmt"
)

func (f *Flow) InvalidateConsentRequest() error {
	if f.ConsentWasHandled {
		return errors.New("consent verifier has already been used")
	}
	if f.State != FlowStateConsentGranted && f.State != FlowStateConsentError {
		return fmt.Errorf("unexpected flow state: expected %d or %d, got %d", FlowStateConsentGranted, FlowStateConsentError, f.State)
	}

	f.ConsentWasHandled = true
	f.State = FlowStateConsentHandled
	return nil
}
