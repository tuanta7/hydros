package domain

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/pkg/dbtype"
)

// Flow represents the flow information associated with an OAuth2/IDToken Connect session.
// It contains information about the login and consent steps that were taken.
type Flow struct {
	LoginChallenge             string              `db:"login_challenge" json:"i"` // PK
	LoginSkip                  bool                `db:"login_skip" json:"ls,omitempty"`
	LoginVerifier              string              `db:"login_verifier" json:"lv,omitempty" `
	LoginCSRF                  string              `db:"login_csrf" json:"lc,omitempty"`
	LoginInitializedAt         sql.NullTime        `db:"login_initialized_at" json:"li,omitempty"`
	LoginRemember              bool                `db:"login_remember" json:"lr,omitempty"`
	LoginRememberFor           int                 `db:"login_remember_for" json:"lf,omitempty"`
	LoginExtendSessionLifetime bool                `db:"login_extend_session_lifetime" json:"ll,omitempty"`
	LoginWasUsed               bool                `db:"login_was_used" json:"lu,omitempty"`
	LoginError                 *RequestDeniedError `db:"login_error" json:"le,omitempty"`
	LoginAuthenticatedAt       sql.NullTime        `db:"login_authenticated_at" json:"la,omitempty"`

	ConsentChallenge   sql.NullString      `db:"consent_challenge" json:"cc,omitempty"`
	ConsentSkip        bool                `db:"consent_skip" json:"cs,omitempty"`
	ConsentVerifier    sql.NullString      `db:"consent_verifier" json:"cv,omitempty"`
	ConsentCSRF        sql.NullString      `db:"consent_csrf" json:"cr,omitempty"`
	ConsentRemember    bool                `db:"consent_remember" json:"ce,omitempty"`
	ConsentRememberFor *int                `db:"consent_remember_for" json:"cf"`
	ConsentWasHandled  bool                `db:"consent_was_used" json:"cw,omitempty"`
	ConsentError       *RequestDeniedError `db:"consent_error" json:"cx"`

	RequestedAt time.Time `db:"requested_at" json:"ia,omitempty"`
	RequestURL  string    `db:"request_url" json:"r,omitempty"`

	State             int16              `db:"state" json:"q,omitempty"`
	RequestedScope    dbtype.StringArray `db:"requested_scope" json:"rs,omitempty"`
	GrantedScope      dbtype.StringArray `db:"granted_scope" json:"gs,omitempty"`
	RequestedAudience dbtype.StringArray `db:"requested_audience" json:"ra,omitempty" `
	GrantedAudience   dbtype.StringArray `db:"granted_at_audience" json:"ga,omitempty"`
	Client            *Client            `db:"-" json:"c,omitempty"`
	ClientID          string             `db:"client_id" json:"ci,omitempty"`
	Subject           string             `db:"subject" json:"s,omitempty"`
	SessionID         sql.NullString     `db:"login_session_id" json:"si,omitempty"`

	ACR                       string              `db:"acr" json:"a,omitempty"`
	AMR                       sql.NullString      `db:"amr" json:"am,omitempty"`
	Context                   json.RawMessage     `db:"context" json:"ct"`
	ForceSubjectIdentifier    string              `db:"forced_subject_identifier" json:"fs,omitempty"`
	IdentityProviderSessionID sql.NullString      `db:"identity_provider_session_id" json:"is,omitempty"`
	SessionIDToken            dbtype.MapStringAny `db:"session_id_token" json:"st"`
	SessionAccessToken        dbtype.MapStringAny `db:"session_access_token" json:"sa"`
}

type RequestDeniedError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Hint             string `json:"error_hint"`
	Code             int    `json:"status_code"`
	Debug            string `json:"error_debug"`
	Valid            bool   `json:"valid"`
}

func (e *RequestDeniedError) IsError() bool {
	return e != nil && e.Valid
}

func (e *RequestDeniedError) ToRFCError() *core.RFC6749Error {
	if e.Error == "" {
		e.Error = "request_denied"
	}

	if e.Code == 0 {
		e.Code = core.ErrInvalidRequest.CodeField
	}

	return &core.RFC6749Error{
		ErrorField:       e.Error,
		DescriptionField: e.ErrorDescription,
		HintField:        e.Hint,
		CodeField:        e.Code,
		DebugField:       e.Debug,
	}
}

func (e *RequestDeniedError) Scan(value any) error {
	v := fmt.Sprintf("%s", value)
	if len(v) == 0 || v == "{}" {
		return nil
	}

	if err := json.Unmarshal([]byte(v), e); err != nil {
		return err
	}

	e.Valid = true
	return nil
}

func (e *RequestDeniedError) Value() (driver.Value, error) {
	if !e.IsError() {
		return "{}", nil
	}

	value, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return string(value), nil
}
