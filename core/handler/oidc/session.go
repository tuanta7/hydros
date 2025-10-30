package oidc

import (
	"time"

	"github.com/mohae/deepcopy"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/signer/jwt"
)

type Session interface {
	IDTokenClaims() *jwt.IDTokenClaims
	IDTokenHeaders() *jwt.Headers
	core.Session
}

type IDTokenSession struct {
	Claims    *jwt.IDTokenClaims           `json:"id_token_claims"`
	Headers   *jwt.Headers                 `json:"headers"`
	Username  string                       `json:"username"`
	Subject   string                       `json:"subject"`
	ExpiresAt map[core.TokenType]time.Time `json:"expires_at"`
}

func (s *IDTokenSession) IDTokenClaims() *jwt.IDTokenClaims {
	if s.Claims == nil {
		s.Claims = &jwt.IDTokenClaims{}
	}
	return s.Claims
}

func (s *IDTokenSession) IDTokenHeaders() *jwt.Headers {
	if s.Headers == nil {
		s.Headers = &jwt.Headers{}
	}
	return s.Headers
}

func (s *IDTokenSession) SetExpiresAt(key core.TokenType, exp time.Time) {
	if s.ExpiresAt == nil {
		s.ExpiresAt = make(map[core.TokenType]time.Time)
	}
	s.ExpiresAt[key] = exp
}

func (s *IDTokenSession) GetExpiresAt(key core.TokenType) time.Time {
	if s.ExpiresAt == nil {
		s.ExpiresAt = make(map[core.TokenType]time.Time)
	}

	if _, ok := s.ExpiresAt[key]; !ok {
		return time.Time{}
	}
	return s.ExpiresAt[key]
}

func (s *IDTokenSession) GetUsername() string {
	if s == nil {
		return ""
	}
	return s.Username
}

func (s *IDTokenSession) GetSubject() string {
	if s == nil {
		return ""
	}

	return s.Subject
}

func (s *IDTokenSession) SetSubject(subject string) {
	s.Subject = subject
}

func (s *IDTokenSession) Clone() core.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(core.Session)
}
