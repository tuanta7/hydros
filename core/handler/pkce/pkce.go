package pkce

import (
	"context"

	"github.com/tuanta7/hydros/core"
)

type ProofKeyForCodeExchangeHandler struct {
}

func (h *ProofKeyForCodeExchangeHandler) HandleAuthorizeRequest(
	ctx context.Context,
	req *core.AuthorizeRequest,
	resp *core.AuthorizeResponse,
) error {
	if !req.ResponseTypes.IncludeAll("code") {
		return nil
	}

	//challenge := req.CodeChallenge
	//challengeMethod := req.CodeChallenge

	return nil
}
