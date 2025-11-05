package client

import "github.com/tuanta7/hydros/core"

func SanitizedClientFromRequest(ar *core.AuthorizeRequest) *Client {
	cl := ar.Client.(*Client)
	cc := &Client{}
	*cc = *cl // copy
	cc.Secret = ""

	return cc
}
