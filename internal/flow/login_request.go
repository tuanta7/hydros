package flow

import (
	"encoding/json"

	"github.com/tuanta7/hydros/pkg/dbtype"
)

type LoginRequest struct{}

type HandledLoginRequest struct {
	Subject                   string             `json:"subject"`
	Remember                  bool               `json:"remember"`
	RememberFor               int                `json:"remember_for"`
	ExtendSessionLifespan     bool               `json:"extend_session_lifespan"`
	ACR                       string             `json:"acr"`
	AMR                       dbtype.StringArray `json:"amr"`
	IdentityProviderSessionID string             `json:"identity_provider_session_id,omitempty"`
	ForceSubjectIdentifier    string             `json:"force_subject_identifier"`
	Context                   json.RawMessage    `json:"context"`
}
