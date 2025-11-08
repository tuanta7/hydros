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
