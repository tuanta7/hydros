package core

import (
	"net/http"
	"net/url"
)

type ResponseModeType string

const (
	ResponseModeDefault  ResponseModeType = ""
	ResponseModeFormPost ResponseModeType = "form_post"
	ResponseModeQuery    ResponseModeType = "query"
	ResponseModeFragment ResponseModeType = "fragment"
)

type AuthorizeRequest struct {
	State         string           `json:"state"`
	RedirectURI   *url.URL         `json:"redirectUri"`
	ResponseTypes Arguments        `json:"responseTypes"`
	ResponseMode  ResponseModeType `json:"responseMode"`
	Request
}

type AuthorizeResponse struct {
	Header     http.Header
	Parameters url.Values
	code       string
}
