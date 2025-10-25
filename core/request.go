package core

import (
	"net/url"
	"time"
)

type Request struct {
	ID                string     `json:"id"`
	RequestedAt       time.Time  `json:"requested_at"`
	Client            Client     `json:"client"`
	RequestedScope    Arguments  `json:"scopes"`
	GrantedScope      Arguments  `json:"granted_scopes"`
	RequestedAudience Arguments  `json:"requested_audience"`
	GrantedAudience   Arguments  `json:"granted_audience"`
	Form              url.Values `json:"form"`
	Session           Session    `json:"session"`
}
