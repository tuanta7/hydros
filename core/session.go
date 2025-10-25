package core

import "time"

type Session struct {
	Username  string                  `json:"username"`
	Subject   string                  `json:"subject"`
	ExpiresAt map[TokenType]time.Time `json:"expires_at"`
	Extra     map[string]any          `json:"extra"`
}

func (s *Session) Clone() *Session {
	expiresAt := make(map[TokenType]time.Time)
	for k, v := range s.ExpiresAt {
		expiresAt[k] = v
	}

	extra := make(map[string]any)
	for k, v := range s.Extra {
		extra[k] = v
	}

	return &Session{
		Username:  s.Username,
		Subject:   s.Subject,
		ExpiresAt: expiresAt,
		Extra:     extra,
	}
}
