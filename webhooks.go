package machinawallet

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

// DefaultWebhookTolerance is the default acceptable clock skew for webhook
// timestamps. Receivers should pass an explicit tolerance for tighter checks.
const DefaultWebhookTolerance = 5 * time.Minute

// VerifySignature validates a MACHINA webhook signature using a
// Stripe-compatible header format:
//
//	t=<unix_seconds>,v1=<hex_hmac_sha256>
//
// The signed payload is "<timestamp>." || body. When tolerance is zero, the
// default of 5 minutes is used. Returns nil on success.
func VerifySignature(payload []byte, signatureHeader, secret string, tolerance time.Duration) error {
	return verifySignatureAt(payload, signatureHeader, secret, tolerance, time.Now())
}

// verifySignatureAt is the testable form that accepts an explicit clock.
func verifySignatureAt(payload []byte, signatureHeader, secret string, tolerance time.Duration, now time.Time) error {
	if signatureHeader == "" {
		return &Error{Code: ErrCodeValidation, Message: "signature header is empty"}
	}
	if secret == "" {
		return &Error{Code: ErrCodeValidation, Message: "webhook secret is empty"}
	}
	if tolerance <= 0 {
		tolerance = DefaultWebhookTolerance
	}

	ts, sig, err := parseWebhookHeader(signatureHeader)
	if err != nil {
		return err
	}

	delta := now.Sub(time.Unix(ts, 0))
	if delta < 0 {
		delta = -delta
	}
	if delta > tolerance {
		return &Error{Code: ErrCodeValidation, Message: "webhook timestamp outside tolerance"}
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return &Error{Code: ErrCodeAuth, Message: "webhook signature mismatch"}
	}
	return nil
}

func parseWebhookHeader(h string) (int64, string, error) {
	var ts int64
	var sig string
	parts := strings.Split(h, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		switch key {
		case "t":
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				continue
			}
			ts = n
		case "v1":
			sig = value
		}
	}
	if ts == 0 || sig == "" {
		return 0, "", &Error{Code: ErrCodeValidation, Message: "malformed signature header — expected t=<unix>,v1=<hex>"}
	}
	return ts, sig, nil
}

// ParseEvent decodes a webhook payload into a CloudEvent envelope.
func ParseEvent(payload []byte) (*CloudEvent, error) {
	if len(payload) == 0 {
		return nil, &Error{Code: ErrCodeValidation, Message: "webhook payload is empty"}
	}
	var ev CloudEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return nil, &Error{Code: ErrCodeValidation, Message: "webhook body is not valid JSON: " + err.Error(), Cause: err}
	}
	if ev.ID == "" || ev.Source == "" || ev.Type == "" {
		return nil, &Error{Code: ErrCodeValidation, Message: "webhook body is missing required CloudEvent fields", Cause: errors.New("missing id/source/type")}
	}
	return &ev, nil
}
