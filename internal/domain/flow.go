package domain

import (
	"database/sql"

	"github.com/tuanta7/hydros/pkg/dbtype"
)

// Flow represents the flow information associated with an OAuth2/IDToken Connect session.
// It contains information about the login and consent steps that were taken.
type Flow struct {
	ID                string             `json:"id" db:"id"`
	RequestedScope    dbtype.StringArray `json:"rs,omitempty" db:"requested_scope" `
	RequestedAudience dbtype.StringArray `json:"ra,omitempty" db:"requested_audience"`

	LoginVerifier string `json:"lv,omitempty" db:"login_verifier"`
	LoginCSRF     string `db:"login_csrf" json:"lc,omitempty"`

	ConsentChallenge sql.NullString `db:"consent_challenge" json:"cc,omitempty"`
	ConsentSkip      bool           `db:"consent_skip" json:"cs,omitempty"`
	ConsentVerifier  sql.NullString `db:"consent_verifier" json:"cv,omitempty"`
	ConsentCSRF      sql.NullString `db:"consent_csrf" json:"cr,omitempty"`
}
