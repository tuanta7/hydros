package session

import (
	"database/sql"

	"github.com/mohae/deepcopy"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/core/signer/jwt"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/pkg/dbtype"
)

// Session is used for methods that handle business logic related to sessions.
type Session struct {
	*oidc.IDTokenSession `json:"id_token"`
	Extra                map[string]any `json:"extra"`
	KeyID                string         `json:"kid"`
	ClientID             string         `json:"client_id"`
	Challenge            string         `json:"challenge"`
	Flow                 *flow.Flow     `json:"-"`

	//ExcludeNotBeforeClaim bool `json:"exclude_not_before_claim"`
	//AllowedTopLevelClaims []string `json:"allowed_top_level_claims"`
	//MirrorTopLevelClaims bool `json:"mirror_top_level_claims"`
}

func NewSession(subject string) *Session {
	return &Session{
		IDTokenSession: &oidc.IDTokenSession{
			Claims:  &jwt.IDTokenClaims{},
			Headers: &jwt.Headers{},
			Subject: subject,
		},
		Challenge: "",
	}
}

func (s *Session) Clone() core.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(core.Session)
}

type LoginSession struct {
	ID                        string          `db:"id"`
	AuthenticatedAt           dbtype.NullTime `db:"authenticated_at"`
	Subject                   string          `db:"subject"`
	IdentityProviderSessionID sql.NullString  `db:"identity_provider_session_id"`
	Remember                  bool            `db:"remember"`
}

func (s *LoginSession) ColumnMap() map[string]any {
	return map[string]any{
		"id":                           s.ID,
		"authenticated_at":             s.AuthenticatedAt,
		"subject":                      s.Subject,
		"identity_provider_session_id": s.IdentityProviderSessionID,
		"remember":                     s.Remember,
	}
}
