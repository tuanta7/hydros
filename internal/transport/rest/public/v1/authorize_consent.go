package v1

import (
	"context"
	"net/http"
	"strings"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/errors"
	"github.com/tuanta7/hydros/internal/flow"
)

func (h *OAuthHandler) handleConsent(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	req *core.AuthorizeRequest,
	flow *flow.Flow,
) (*flow.Flow, error) {
	consentVerifier := strings.TrimSpace(req.Form.Get("consent_verifier"))
	if consentVerifier == "" {
		return nil, h.requestConsent(ctx, w, r, req, flow)
	}

	return h.verifyConsent(ctx, consentVerifier)
}

func (h *OAuthHandler) requestConsent(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
	flow *flow.Flow,
) error {
	return errors.ErrAbortOAuth2Request
}

func (h *OAuthHandler) verifyConsent(ctx context.Context, verifier string) (*flow.Flow, error) {
	f, err := h.flowUC.VerifyAndInvalidateLoginRequest(ctx, verifier)
	if err != nil {
		return nil, err
	}

	return f, nil
}
