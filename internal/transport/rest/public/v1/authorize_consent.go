package v1

import (
	"context"
	stderr "errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/errors"
	"github.com/tuanta7/hydros/internal/flow"
	"github.com/tuanta7/hydros/internal/session"
	"github.com/tuanta7/hydros/pkg/dbtype"
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

	return h.verifyConsent(ctx, r, consentVerifier)
}

func (h *OAuthHandler) requestConsent(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
	f *flow.Flow,
) error {
	if ar.Prompt.IncludeAll("consent") {
		return h.forwardConsentRequest(ctx, w, r, ar, f, nil)
	}

	if ar.Client.IsPublic() && !x.IsURISecure(ar.RedirectURI) {
		// insecure redirect uri for public clients always requires consent
		return h.forwardConsentRequest(ctx, w, r, ar, f, nil)
	}

	consentSession, err := h.flowUC.FindGrantedAndRememberedConsentRequest(ctx, ar.Client.GetID(), f.Subject)
	if stderr.Is(err, errors.ErrNoPreviousConsentFound) {
		return h.forwardConsentRequest(ctx, w, r, ar, f, nil)
	} else if err != nil {
		return err
	}

	scopeStrategy := h.cfg.GetScopeStrategy()
	for _, s := range ar.RequestedScope {
		// if a new scope is required, forward the consent request without the consent session
		if !scopeStrategy(consentSession.GrantedScope, s) {
			return h.forwardConsentRequest(ctx, w, r, ar, f, nil)
		}
	}

	return h.forwardConsentRequest(ctx, w, r, ar, f, consentSession)
}

func (h *OAuthHandler) verifyConsent(ctx context.Context, r *http.Request, verifier string) (*flow.Flow, error) {
	f, err := h.flowUC.DecodeFlow(ctx, verifier, flow.AsConsentVerifier)
	if err != nil {
		return nil, err
	}

	err = f.InvalidateConsentRequest()
	if err != nil {
		return nil, core.ErrInvalidRequest.WithDebug(err.Error())
	}

	// persist login and consent request
	err = h.flowUC.SaveFlow(ctx, f)
	if err != nil {
		return nil, err
	}

	if f.RequestedAt.Add(h.cfg.GetConsentRequestMaxAge()).Before(x.NowUTC()) {
		return nil, core.ErrRequestUnauthorized.WithHint("The consent request has expired, please try again.")
	}

	if f.ConsentError.IsError() {
		f.ConsentError.SetDefaults(flow.ConsentRequestDeniedErrorName)
		return nil, f.ConsentError.ToRFCError()
	}

	if time.Time(f.LoginAuthenticatedAt).IsZero() {
		return nil, core.ErrServerError.WithHint("The authenticated time value was not set.")
	}

	err = session.ValidateCSRFSession(r, h.store, session.ConsentCSRFCookieKey, f.ConsentCSRF.String())
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (h *OAuthHandler) forwardConsentRequest(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	ar *core.AuthorizeRequest,
	f *flow.Flow,
	previousConsent *flow.Flow,
) error {
	skip := false
	if previousConsent != nil {
		skip = true
	}

	csrf := x.RandomUUID()

	f.ConsentSkip = skip
	f.ConsentCSRF = dbtype.NullString(csrf)
	f.State = flow.StateConsentInitialized

	err := session.CreateCSRFSession(w, r, h.cfg, h.store, session.ConsentCSRFCookieKey, csrf, h.cfg.GetConsentRequestMaxAge())
	if err != nil {
		return err
	}

	encodedFlow, err := h.flowUC.EncodeFlow(ctx, f, flow.AsConsentChallenge)
	if err != nil {
		return err
	}

	redirectTo := h.cfg.GetConsentPageURL()
	params := url.Values{}
	params.Set("consent_challenge", encodedFlow)
	redirectTo.RawQuery = params.Encode()

	http.Redirect(w, r, redirectTo.String(), http.StatusFound)
	return errors.ErrAbortOAuth2Request
}
