package core

import (
	"net/url"
	"time"

	"github.com/tuanta7/hydros/core/x"
)

type Request struct {
	ID              string     `json:"id" form:"-"`
	RequestedAt     time.Time  `json:"requested_at" form:"-"`
	Scope           Arguments  `json:"scope" form:"scope"`
	GrantedScope    Arguments  `json:"granted_scope" form:"-"`
	Audience        Arguments  `json:"audience" form:"audience"`
	GrantedAudience Arguments  `json:"granted_audience" form:"-"`
	Form            url.Values `json:"form" form:"-"`
	Client          Client     `json:"client" form:"-"`
	Session         Session    `json:"session" form:"-"`
}

func NewRequest() *Request {
	return &Request{
		RequestedAt:     x.NowUTC(),
		Scope:           Arguments{},
		Audience:        Arguments{},
		GrantedAudience: Arguments{},
		GrantedScope:    Arguments{},
		Form:            url.Values{},
	}
}

// Merge merges back target request values into the current one.
func (r *Request) Merge(target *Request) {
	r.ID = target.ID
	r.RequestedAt = target.RequestedAt
	r.Scope = r.Scope.Append(target.Scope...)
	r.GrantedScope = r.GrantedScope.Append(target.GrantedScope...)
	r.Audience = r.Audience.Append(target.Audience...)
	r.GrantedAudience = r.GrantedAudience.Append(target.GrantedAudience...)
	r.Client = target.Client
	r.Session = target.Session

	for k, v := range target.Form {
		r.Form[k] = v
	}
}

// Sanitize removes sensitive data from the request for storage.
func (r *Request) Sanitize(allowedParameters ...string) *Request {
	allowed := map[string]bool{
		"grant_type":    true,
		"response_type": true,
		"scope":         true,
		"client_id":     true,
	}

	for _, k := range allowedParameters {
		allowed[k] = true
	}

	sr := &Request{
		ID:              r.ID,
		RequestedAt:     r.RequestedAt,
		Scope:           r.Scope,
		GrantedScope:    r.GrantedScope,
		Audience:        r.Audience,
		GrantedAudience: r.GrantedAudience,
		Form:            url.Values{},
		Client:          r.Client,
		Session:         r.Session,
	}

	for k, v := range r.Form {
		if allowed[k] {
			sr.Form[k] = v
		}
	}

	return sr
}
