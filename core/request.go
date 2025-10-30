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
