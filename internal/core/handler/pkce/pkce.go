package pkce

import "github.com/tuanta7/hydros/internal/core"

type ProofKeyForCodeExchangeHandler struct {
}

func (h *ProofKeyForCodeExchangeHandler) HandleAuthorizeRequest(req *core.AuthorizeRequest) (*core.AuthorizeResponse, error) {
	return nil, nil
}
