package machinawallet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// transport encapsulates request signing, retries, and JSON decoding.
type transport struct {
	cfg    Config
	signer Signer
	http   *http.Client
	// auth, when non-nil, attaches a second authorization artifact in
	// addition to the app-level signature. Set via Client.WithAuthorization.
	auth AuthorizationContext
}

// newTransport constructs a transport from a validated config.
func newTransport(cfg Config, signer Signer) *transport {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	} else if client.Timeout == 0 {
		client.Timeout = cfg.Timeout
	}
	return &transport{cfg: cfg, signer: signer, http: client}
}

// request executes an HTTP request with signing, retries, and JSON decoding.
// When out is non-nil the response body is decoded into it. The method
// performs up to cfg.MaxRetries retries for retryable error classes.
func (t *transport) request(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return &Error{Code: ErrCodeValidation, Message: "marshaling request body: " + err.Error(), Cause: err}
		}
	}

	fullPath := path
	if len(query) > 0 {
		fullPath = path + "?" + query.Encode()
	}
	fullURL := t.cfg.BaseURL + fullPath

	var lastErr error
	for attempt := 0; attempt <= t.cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return &Error{Code: ErrCodeNetwork, Message: "building request: " + err.Error(), Cause: err}
		}
		req.Header.Set("Accept", "application/json")
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("User-Agent", t.cfg.UserAgent)

		// Sign the canonical path (without query string). Server SDK signs the
		// same way for parity with TS and Python SDKs.
		sig, mode, err := t.signer.Sign(method, path, bodyBytes, time.Now())
		if err != nil {
			return &Error{Code: ErrCodeAuth, Message: "signing request: " + err.Error(), Cause: err}
		}
		req.Header.Set(HeaderService, t.cfg.AppID)
		req.Header.Set(HeaderTimestamp, strconv.FormatInt(time.Now().Unix(), 10))
		req.Header.Set(HeaderSignature, sig)
		req.Header.Set(HeaderMode, mode)

		// Layer the optional per-user authorization artifact on top of
		// the app-level signature when present.
		if t.auth != nil {
			if err := t.auth.apply(req, bodyBytes); err != nil {
				return &Error{Code: ErrCodeAuth, Message: "authorization context: " + err.Error(), Cause: err}
			}
		}

		resp, err := t.http.Do(req)
		if err != nil {
			lastErr = &Error{Code: ErrCodeNetwork, Message: err.Error(), Cause: err}
			if !shouldRetry(0, err) || attempt == t.cfg.MaxRetries {
				return lastErr
			}
			sleepBackoff(ctx, attempt, 0)
			continue
		}

		// Read body up front so we can both decode it and surface error details.
		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = &Error{Code: ErrCodeNetwork, Message: "reading response: " + readErr.Error(), StatusCode: resp.StatusCode, Cause: readErr}
			if !shouldRetry(resp.StatusCode, readErr) || attempt == t.cfg.MaxRetries {
				return lastErr
			}
			sleepBackoff(ctx, attempt, retryAfter(resp))
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil || len(respBody) == 0 {
				return nil
			}
			if err := json.Unmarshal(respBody, out); err != nil {
				return &Error{Code: ErrCodeValidation, Message: "decoding response: " + err.Error(), StatusCode: resp.StatusCode, Cause: err}
			}
			return nil
		}

		mapped := mapResponseError(resp.StatusCode, respBody)
		lastErr = mapped
		if !shouldRetry(resp.StatusCode, nil) || attempt == t.cfg.MaxRetries {
			return mapped
		}
		sleepBackoff(ctx, attempt, retryAfter(resp))
	}
	if lastErr == nil {
		lastErr = &Error{Code: ErrCodeNetwork, Message: "exhausted retries"}
	}
	return lastErr
}

// mapResponseError translates an HTTP error response into a typed SDK error.
func mapResponseError(status int, body []byte) error {
	env := errorEnvelope{}
	_ = json.Unmarshal(body, &env)
	code := env.Error.Code
	if code == "" {
		code = mapStatus(status)
	}
	msg := env.Error.Message
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}
	base := &Error{Code: code, Message: msg, StatusCode: status}

	switch code {
	case ErrCodePolicyDenied:
		return &PolicyDeniedError{Err: base, PolicyID: env.Error.PolicyID, PolicyName: env.Error.PolicyName}
	case ErrCodeApprovalRequired:
		return &ApprovalRequiredError{Err: base, ApprovalID: env.Error.ApprovalID}
	}
	return base
}

// shouldRetry returns true if the request can be safely retried.
func shouldRetry(status int, networkErr error) bool {
	if networkErr != nil {
		// Network errors are retryable except context cancellation.
		return !errors.Is(networkErr, context.Canceled) && !errors.Is(networkErr, context.DeadlineExceeded)
	}
	if status == http.StatusTooManyRequests {
		return true
	}
	if status >= 500 && status <= 599 {
		return true
	}
	return false
}

// retryAfter parses the Retry-After header into a duration. Returns 0 when
// the header is absent or unparseable.
func retryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	h := resp.Header.Get("Retry-After")
	if h == "" {
		return 0
	}
	if secs, err := strconv.Atoi(h); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(h); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

// sleepBackoff blocks for an exponential backoff with jitter, honoring an
// optional Retry-After hint and returning early if the context is canceled.
func sleepBackoff(ctx context.Context, attempt int, hint time.Duration) {
	d := hint
	if d <= 0 {
		base := 200 * time.Millisecond
		max := 5 * time.Second
		expo := time.Duration(math.Pow(2, float64(attempt))) * base
		if expo > max {
			expo = max
		}
		// Full jitter
		d = time.Duration(rand.Int63n(int64(expo) + 1))
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// formatChainQuery encodes a Chain as a query parameter value, omitting empty.
func setIfNonEmpty(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

// setIfNonZero encodes an int parameter when non-zero.
func setIfNonZero(q url.Values, key string, value int) {
	if value != 0 {
		q.Set(key, fmt.Sprintf("%d", value))
	}
}
