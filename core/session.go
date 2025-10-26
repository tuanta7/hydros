package core

import "time"

type Session interface {
	SetExpiresAt(key TokenType, exp time.Time)
	GetExpiresAt(key TokenType) time.Time
	GetUsername() string
	GetSubject() string
	Clone() Session
}

type JWTProfileSession interface {
	SetSubject(subject string)
}

type ExtraClaimsSession interface {
	GetExtraClaims() map[string]any
}

type DefaultSession struct {
	Username  string                  `json:"username"`
	Subject   string                  `json:"subject"`
	ExpiresAt map[TokenType]time.Time `json:"expires_at"`
	Extra     map[string]any          `json:"extra"`
}

func (s *DefaultSession) SetExpiresAt(key TokenType, exp time.Time) {
	if s.ExpiresAt == nil {
		s.ExpiresAt = make(map[TokenType]time.Time)
	}
	s.ExpiresAt[key] = exp
}

func (s *DefaultSession) GetExpiresAt(key TokenType) time.Time {
	if s.ExpiresAt == nil {
		s.ExpiresAt = make(map[TokenType]time.Time)
	}

	return s.ExpiresAt[key]
}

func (s *DefaultSession) GetUsername() string {
	if s == nil {
		return ""
	}
	return s.Username
}

func (s *DefaultSession) SetSubject(subject string) {
	s.Subject = subject
}

func (s *DefaultSession) GetSubject() string {
	if s == nil {
		return ""
	}

	return s.Subject
}

func (s *DefaultSession) Clone() Session {
	expiresAt := make(map[TokenType]time.Time)
	for k, v := range s.ExpiresAt {
		expiresAt[k] = v
	}

	extra := make(map[string]any)
	for k, v := range s.Extra {
		extra[k] = v
	}

	return &DefaultSession{
		Username:  s.Username,
		Subject:   s.Subject,
		ExpiresAt: expiresAt,
		Extra:     extra,
	}
}

func (s *DefaultSession) GetExtraClaims() map[string]any {
	if s == nil {
		return nil
	}

	if s.Extra == nil {
		s.Extra = make(map[string]any)
	}

	return s.Extra
}
