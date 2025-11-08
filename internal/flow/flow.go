package flow

import (
	"encoding/json"
	"time"

	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/pkg/dbtype"
)

const (
	FlowStateLoginInitialized   = int16(1)
	FlowStateLoginAuthenticated = int16(2) // done on the user side
	FlowStateLoginHandled       = int16(3) // done on the server side
	FlowStateConsentInitialized = int16(4)
	FlowStateConsentGranted     = int16(5)
	FlowStateConsentHandled     = int16(6)
	FlowStateLoginError         = int16(128)
	FlowStateConsentError       = int16(129)
)

// Flow represents the flow information associated with an OAuth2/IDToken Connect session.
// It contains information about the login and consent steps that were taken.
type Flow struct {
	ID                         string              `db:"id" json:"i"`
	ACR                        string              `db:"acr" json:"a,omitempty"`
	AMR                        dbtype.StringArray  `db:"amr" json:"am,omitempty"`
	LoginSkip                  bool                `db:"login_skip" json:"ls,omitempty"`
	LoginCSRF                  string              `db:"login_csrf" json:"lc,omitempty"`
	LoginRemember              bool                `db:"login_remember" json:"lr,omitempty"`
	LoginRememberFor           int                 `db:"login_remember_for" json:"lf,omitempty"`
	LoginExtendSessionLifetime bool                `db:"login_extend_session_lifetime" json:"ll,omitempty"`
	LoginWasHandled            bool                `db:"login_was_handled" json:"lw,omitempty"`
	LoginError                 *RequestDeniedError `db:"login_error" json:"le,omitempty"`
	LoginAuthenticatedAt       dbtype.NullTime     `db:"login_authenticated_at" json:"la,omitempty"`
	LoginSessionID             dbtype.NullString   `db:"login_session_id" json:"si,omitempty"`
	Subject                    string              `db:"subject" json:"s,omitempty"`
	ForcedSubjectIdentifier    string              `db:"forced_subject_identifier" json:"fs,omitempty"`
	IdentityProviderSessionID  dbtype.NullString   `db:"identity_provider_session_id" json:"is,omitempty"`

	ConsentSkip        bool                `db:"consent_skip" json:"cs,omitempty"`
	ConsentCSRF        dbtype.NullString   `db:"consent_csrf" json:"cr,omitempty"`
	ConsentRemember    bool                `db:"consent_remember" json:"ce,omitempty"`
	ConsentRememberFor *int                `db:"consent_remember_for" json:"cf"`
	ConsentWasHandled  bool                `db:"consent_was_handled" json:"cw,omitempty"`
	ConsentError       *RequestDeniedError `db:"consent_error" json:"cx"`
	ConsentHandledAt   dbtype.NullTime     `db:"consent_handled_at" json:"ch,omitempty"`

	RequestedAt       time.Time          `db:"requested_at" json:"ia,omitempty"`
	RequestURL        string             `db:"request_url" json:"r,omitempty"`
	RequestedScope    dbtype.StringArray `db:"requested_scope" json:"rs,omitempty"`
	GrantedScope      dbtype.StringArray `db:"granted_scope" json:"gs,omitempty"`
	RequestedAudience dbtype.StringArray `db:"requested_audience" json:"ra,omitempty" `
	GrantedAudience   dbtype.StringArray `db:"granted_at_audience" json:"ga,omitempty"`
	Client            *client.Client     `db:"-" json:"c,omitempty"`
	ClientID          string             `db:"client_id" json:"ci,omitempty"`
	Context           json.RawMessage    `db:"context" json:"ct"`      // is it used tho?
	OIDCContext       json.RawMessage    `db:"oidc_context" json:"oc"` // is it used tho?
	State             int16              `db:"state" json:"q,omitempty"`
}

func (f *Flow) ColumnMap() map[string]any {
	return map[string]interface{}{
		"id":                            f.ID,
		"acr":                           f.ACR,
		"amr":                           f.AMR,
		"login_skip":                    f.LoginSkip,
		"login_csrf":                    f.LoginCSRF,
		"login_remember":                f.LoginRemember,
		"login_remember_for":            f.LoginRememberFor,
		"login_extend_session_lifetime": f.LoginExtendSessionLifetime,
		"login_was_handled":             f.LoginWasHandled,
		"login_error":                   f.LoginError,
		"login_authenticated_at":        f.LoginAuthenticatedAt,
		"login_session_id":              f.LoginSessionID,
		"subject":                       f.Subject,
		"forced_subject_identifier":     f.ForcedSubjectIdentifier,
		"identity_provider_session_id":  f.IdentityProviderSessionID,

		"consent_skip":         f.ConsentSkip,
		"consent_csrf":         f.ConsentCSRF,
		"consent_remember":     f.ConsentRemember,
		"consent_remember_for": f.ConsentRememberFor,
		"consent_was_handled":  f.ConsentWasHandled,
		"consent_error":        f.ConsentError,
		"consent_handled_at":   f.ConsentHandledAt,

		"requested_at":       f.RequestedAt,
		"request_url":        f.RequestURL,
		"requested_scope":    f.RequestedScope,
		"granted_scope":      f.GrantedScope,
		"requested_audience": f.RequestedAudience,
		"granted_audience":   f.GrantedAudience,
		"client_id":          f.ClientID,
		"context":            f.Context,
		"oidc_context":       f.OIDCContext,
		"state":              f.State,
	}
}
