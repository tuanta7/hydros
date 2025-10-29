package domain

import (
	"github.com/mohae/deepcopy"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/handler/oidc"
	"github.com/tuanta7/hydros/core/token/jwt"
)

// Session is used for methods that handle business logic related to sessions.
type Session struct {
	*oidc.IDTokenSession `json:"id_token"`
	Extra                map[string]any `json:"extra"`
	KeyID                string         `json:"kid"`
	ClientID             string         `json:"client_id"`
	Challenge            string         `json:"challenge"`
	Flow                 *Flow          `json:"-"`

	//ExcludeNotBeforeClaim bool `json:"exclude_not_before_claim"`
	//AllowedTopLevelClaims []string `json:"allowed_top_level_claims"`
	//MirrorTopLevelClaims bool `json:"mirror_top_level_claims"`
}

func (s *Session) Clone() core.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(core.Session)
}

func NewSession(subject string) *Session {
	return &Session{
		IDTokenSession: &oidc.IDTokenSession{
			Claims:  &jwt.IDTokenClaims{},
			Headers: &jwt.Headers{},
			Subject: subject,
		},
	}
}
