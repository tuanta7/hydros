package flow

import (
	"encoding/json"
	"time"

	"github.com/tuanta7/hydros/internal/client"
	"github.com/tuanta7/hydros/pkg/dbtype"
)

const (
	StateLoginInitialized   = int16(1)   // start the login flow
	StateLoginAuthenticated = int16(2)   // login done on the user side
	StateLoginError         = int16(128) // login done on the user side with an error
	StateLoginHandled       = int16(3)   // login done on the server side

	StateConsentInitialized = int16(4)   // start the consent flow
	StateConsentGranted     = int16(5)   // consent done on the user side
	StateConsentError       = int16(129) // consent done on the user side with an error
	StateConsentHandled     = int16(6)   // consent done on the server side
)

// Flow represents the flow information associated with an OAuth2/IDToken Connect session.
// It contains information about the login and consent steps that were taken.
type Flow struct {
	ID                         string              `db:"id" json:"i"`
	ACR                        string              `db:"acr" json:"a,omitempty"`  // not supported yet
	AMR                        dbtype.StringArray  `db:"amr" json:"am,omitempty"` // not supported yet
	LoginSkip                  bool                `db:"login_skip" json:"ls,omitempty"`
	LoginExtendSessionLifetime bool                `db:"login_extend_session_lifetime" json:"ll,omitempty"`
	LoginCSRF                  string              `db:"login_csrf" json:"lc,omitempty"`
	LoginRemember              bool                `db:"login_remember" json:"lr,omitempty"`
	LoginRememberFor           int                 `db:"login_remember_for" json:"lf,omitempty"`
	LoginAuthenticatedAt       dbtype.NullTime     `db:"login_authenticated_at" json:"la,omitempty"`
	LoginError                 *RequestDeniedError `db:"login_error" json:"le,omitempty"`
	LoginWasHandled            bool                `db:"login_was_handled" json:"lw,omitempty"` // prevent double-submits

	LoginSessionID            dbtype.NullString `db:"login_session_id" json:"si,omitempty"`
	Subject                   string            `db:"subject" json:"s,omitempty"`
	ForcedSubjectIdentifier   string            `db:"forced_subject_identifier" json:"fs,omitempty"`
	IdentityProviderSessionID dbtype.NullString `db:"identity_provider_session_id" json:"is,omitempty"`

	ConsentSkip        bool                `db:"consent_skip" json:"cs,omitempty"`
	ConsentCSRF        dbtype.NullString   `db:"consent_csrf" json:"cr,omitempty"`
	ConsentRemember    bool                `db:"consent_remember" json:"ce,omitempty"`
	ConsentRememberFor *int                `db:"consent_remember_for" json:"cf"`
	ConsentGrantedAt   dbtype.NullTime     `db:"consent_granted_at" json:"ch,omitempty"`
	ConsentError       *RequestDeniedError `db:"consent_error" json:"cx"`
	ConsentWasHandled  bool                `db:"consent_was_handled" json:"cw,omitempty"`

	RequestedAt       time.Time          `db:"requested_at" json:"ia,omitempty"`
	RequestURL        string             `db:"request_url" json:"r,omitempty"`
	RequestedScope    dbtype.StringArray `db:"requested_scope" json:"rs,omitempty"`
	GrantedScope      dbtype.StringArray `db:"granted_scope" json:"gs,omitempty"`
	RequestedAudience dbtype.StringArray `db:"requested_audience" json:"ra,omitempty" `
	GrantedAudience   dbtype.StringArray `db:"granted_audience" json:"ga,omitempty"`
	Client            *client.Client     `db:"-" json:"c,omitempty"`
	ClientID          string             `db:"client_id" json:"ci,omitempty"`
	Context           json.RawMessage    `db:"context" json:"ct"`      // is it used tho?
	OIDCContext       json.RawMessage    `db:"oidc_context" json:"oc"` // is it used tho?
	State             int16              `db:"state" json:"q,omitempty"`
}

func (f *Flow) ColumnMap() map[string]any {
	return map[string]any{
		"id":                            f.ID,
		"acr":                           f.ACR,
		"amr":                           f.AMR,
		"login_skip":                    f.LoginSkip,
		"login_extend_session_lifetime": f.LoginExtendSessionLifetime,
		"login_csrf":                    f.LoginCSRF,
		"login_remember":                f.LoginRemember,
		"login_remember_for":            f.LoginRememberFor,
		"login_authenticated_at":        f.LoginAuthenticatedAt,
		"login_error":                   f.LoginError,
		"login_was_handled":             f.LoginWasHandled,

		"login_session_id":             f.LoginSessionID,
		"subject":                      f.Subject,
		"forced_subject_identifier":    f.ForcedSubjectIdentifier,
		"identity_provider_session_id": f.IdentityProviderSessionID,

		"consent_skip":         f.ConsentSkip,
		"consent_csrf":         f.ConsentCSRF,
		"consent_remember":     f.ConsentRemember,
		"consent_remember_for": f.ConsentRememberFor,
		"consent_granted_at":   f.ConsentGrantedAt,
		"consent_error":        f.ConsentError,
		"consent_was_handled":  f.ConsentWasHandled,

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
