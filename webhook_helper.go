package machinawallet

import (
	"time"

	"github.com/machina-xyz/machina-wallet-sdk-go/webhooks"
)

// WebhookHelper is a stateful verifier + parser for incoming webhook
// deliveries. Construct one per signing secret and reuse it across requests;
// the helper is safe for concurrent use.
//
// Usage in an HTTP handler:
//
//	helper := machinawallet.NewWebhookHelper(secret)
//	event, err := helper.VerifyAndParse(rawBody, r.Header.Get("X-Machina-Signature"))
//	if err != nil {
//	    http.Error(w, err.Error(), http.StatusBadRequest)
//	    return
//	}
//	switch ev := event.(type) {
//	case webhooks.WalletCreatedEvent:
//	    handleWalletCreated(ev.Data)
//	}
type WebhookHelper struct {
	secret    string
	tolerance time.Duration
}

// WebhookHelperOption is a functional option for NewWebhookHelper.
type WebhookHelperOption func(*WebhookHelper)

// WithWebhookTolerance overrides the default signature-timestamp tolerance.
// A typical receiver picks a value between 1 and 5 minutes.
func WithWebhookTolerance(d time.Duration) WebhookHelperOption {
	return func(h *WebhookHelper) {
		if d > 0 {
			h.tolerance = d
		}
	}
}

// NewWebhookHelper constructs a WebhookHelper for the given signing secret.
// The default tolerance is DefaultWebhookTolerance (5 minutes); override
// with WithWebhookTolerance.
func NewWebhookHelper(secret string, opts ...WebhookHelperOption) *WebhookHelper {
	h := &WebhookHelper{secret: secret, tolerance: DefaultWebhookTolerance}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// VerifyAndParse verifies the signature header against the raw body and
// returns the strongly-typed event. Returns an authentication error when the
// signature is invalid and a validation error when the body cannot be
// decoded.
func (h *WebhookHelper) VerifyAndParse(rawBody []byte, signatureHeader string) (webhooks.Event, error) {
	if err := h.Verify(rawBody, signatureHeader); err != nil {
		return nil, err
	}
	return h.UnsafeUnwrap(rawBody)
}

// Verify checks the signature header against the raw body without parsing
// the payload. Use this when the caller needs to forward the raw body to
// another worker instead of decoding it locally.
func (h *WebhookHelper) Verify(rawBody []byte, signatureHeader string) error {
	return VerifySignature(rawBody, signatureHeader, h.secret, h.tolerance)
}

// UnsafeUnwrap decodes the raw body into the strongest typed event the SDK
// knows without verifying the signature. Only use this after Verify has
// returned nil, or when the body originates from a trusted internal source.
func (h *WebhookHelper) UnsafeUnwrap(rawBody []byte) (webhooks.Event, error) {
	if len(rawBody) == 0 {
		return nil, &Error{Code: ErrCodeValidation, Message: "webhook body is empty"}
	}
	ev, err := webhooks.Parse(rawBody)
	if err != nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "webhook payload decode: " + err.Error(), Cause: err}
	}
	return ev, nil
}
