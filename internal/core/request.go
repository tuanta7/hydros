package core

import (
	"net/url"
	"time"
)

type Request struct {
	ID                string     `json:"id"`
	RequestedAt       time.Time  `json:"requestedAt"`
	Client            Client     `json:"client"`
	RequestedScope    Arguments  `json:"scopes"`
	GrantedScope      Arguments  `json:"grantedScopes"`
	RequestedAudience Arguments  `json:"requestedAudience"`
	GrantedAudience   Arguments  `json:"grantedAudience"`
	Form              url.Values `json:"form"`
}
