package machinawallet

import (
	"errors"
	"fmt"
)

// Error codes returned by the MACHINA wallet management API.
const (
	ErrCodeAuth             = "auth_error"
	ErrCodeValidation       = "validation_error"
	ErrCodePolicyDenied     = "policy_denied"
	ErrCodeApprovalRequired = "approval_required"
	ErrCodeNetwork          = "network_error"
	ErrCodeRateLimit        = "rate_limit_error"
	ErrCodeNotFound         = "not_found"
	ErrCodeServer           = "server_error"
	ErrCodeUnknown          = "unknown_error"
)

// Error is the base SDK error type. All non-IO errors returned by the SDK can
// be unwrapped to or compared against the sentinel Err* variables.
type Error struct {
	// Code is a stable machine-readable error code.
	Code string
	// Message is a human-readable description.
	Message string
	// StatusCode is the HTTP status code returned by the server, when known.
	StatusCode int
	// Cause is the underlying error, when wrapping IO or parse errors.
	Cause error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return fmt.Sprintf("machinawallet: %s", e.Code)
	}
	return fmt.Sprintf("machinawallet: %s: %s", e.Code, e.Message)
}

// Unwrap exposes the underlying cause for errors.Is and errors.As.
func (e *Error) Unwrap() error { return e.Cause }

// Is reports whether the target is the same error code. Sentinel errors below
// only set Code, so comparing by Code is the intended behavior.
func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return e == nil && target == nil
	}
	var t *Error
	if !errors.As(target, &t) {
		return false
	}
	return e.Code == t.Code
}

// Sentinel errors. Use errors.Is to test for these.
var (
	ErrAuth             = &Error{Code: ErrCodeAuth, Message: "authentication failed"}
	ErrValidation       = &Error{Code: ErrCodeValidation, Message: "request validation failed"}
	ErrPolicyDenied     = &Error{Code: ErrCodePolicyDenied, Message: "policy denied"}
	ErrApprovalRequired = &Error{Code: ErrCodeApprovalRequired, Message: "approval required"}
	ErrNetwork          = &Error{Code: ErrCodeNetwork, Message: "network error"}
	ErrRateLimit        = &Error{Code: ErrCodeRateLimit, Message: "rate limited"}
	ErrNotFound         = &Error{Code: ErrCodeNotFound, Message: "not found"}
	ErrServer           = &Error{Code: ErrCodeServer, Message: "server error"}
)

// PolicyDeniedError is returned when a spend policy denies an action. It
// always unwraps to ErrPolicyDenied.
type PolicyDeniedError struct {
	Err        *Error
	PolicyID   string
	PolicyName string
}

// Error implements the error interface.
func (e *PolicyDeniedError) Error() string {
	if e == nil || e.Err == nil {
		return "machinawallet: policy_denied"
	}
	return e.Err.Error()
}

// Is reports identity with ErrPolicyDenied.
func (e *PolicyDeniedError) Is(target error) bool {
	if target == ErrPolicyDenied {
		return true
	}
	if e == nil || e.Err == nil {
		return false
	}
	return e.Err.Is(target)
}

// Unwrap returns the embedded Error.
func (e *PolicyDeniedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// ApprovalRequiredError is returned when a transaction has been queued and
// requires human approval before broadcasting. It always unwraps to
// ErrApprovalRequired.
type ApprovalRequiredError struct {
	Err        *Error
	ApprovalID string
}

// Error implements the error interface.
func (e *ApprovalRequiredError) Error() string {
	if e == nil || e.Err == nil {
		return "machinawallet: approval_required"
	}
	return e.Err.Error()
}

// Is reports identity with ErrApprovalRequired.
func (e *ApprovalRequiredError) Is(target error) bool {
	if target == ErrApprovalRequired {
		return true
	}
	if e == nil || e.Err == nil {
		return false
	}
	return e.Err.Is(target)
}

// Unwrap returns the embedded Error.
func (e *ApprovalRequiredError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// errorEnvelope is the JSON shape used by the wallet management API to report
// errors. It is intentionally permissive — extra fields are ignored.
type errorEnvelope struct {
	Error struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		PolicyID   string `json:"policy_id,omitempty"`
		PolicyName string `json:"policy_name,omitempty"`
		ApprovalID string `json:"approval_id,omitempty"`
	} `json:"error"`
}

// mapStatus returns a sensible code for an HTTP status when the server did not
// supply a structured error body.
func mapStatus(status int) string {
	switch {
	case status == 401 || status == 403:
		return ErrCodeAuth
	case status == 404:
		return ErrCodeNotFound
	case status == 412:
		return ErrCodePolicyDenied
	case status == 422:
		return ErrCodeValidation
	case status == 429:
		return ErrCodeRateLimit
	case status >= 500:
		return ErrCodeServer
	case status >= 400:
		return ErrCodeValidation
	default:
		return ErrCodeUnknown
	}
}
