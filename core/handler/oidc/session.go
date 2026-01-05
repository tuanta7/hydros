package oidc

import (
	"time"

	"github.com/mohae/deepcopy"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/signer/jwt"
)

type OpenIDConnectSession interface {
	IDTokenClaims() *jwt.IDTokenClaims
	core.Session
}

type IDTokenSession struct {
	Claims    *jwt.IDTokenClaims           `json:"id_token_claims"`
	Username  string                       `json:"username"`
	ExpiresAt map[core.TokenType]time.Time `json:"expires_at"`
}

func (s *IDTokenSession) IDTokenClaims() *jwt.IDTokenClaims {
	if s.Claims == nil {
		s.Claims = &jwt.IDTokenClaims{}
	}
	return s.Claims
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

	return s.Claims.Subject
}

func (s *IDTokenSession) SetSubject(subject string) {
	s.Claims.Subject = subject
}

func (s *IDTokenSession) Clone() core.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(core.Session)
}
