package domain

import "github.com/tuanta7/hydros/pkg/dbtype"

// Flow represents the flow information associated with an OAuth2/IDToken Connect session.
// It contains information about the login and consent steps that were taken.
type Flow struct {
	ID                string             `json:"id" db:"id"`
	RequestedScope    dbtype.StringArray `json:"rs,omitempty" db:"requested_scope" `
	RequestedAudience dbtype.StringArray `json:"ra,omitempty" db:"requested_at_audience"`
}
