package machinawallet

import (
	"errors"
	"testing"
)

func TestErrorIs(t *testing.T) {
	t.Parallel()

	e := &Error{Code: ErrCodeNotFound, Message: "missing"}
	if !errors.Is(e, ErrNotFound) {
		t.Errorf("errors.Is(e, ErrNotFound) failed")
	}
	if errors.Is(e, ErrAuth) {
		t.Errorf("unexpected match with ErrAuth")
	}
}

func TestPolicyDeniedErrorIs(t *testing.T) {
	t.Parallel()

	e := &PolicyDeniedError{
		Err:      &Error{Code: ErrCodePolicyDenied, Message: "blocked"},
		PolicyID: "pol_1",
	}
	if !errors.Is(e, ErrPolicyDenied) {
		t.Errorf("errors.Is(e, ErrPolicyDenied) failed")
	}
}

func TestApprovalRequiredErrorIs(t *testing.T) {
	t.Parallel()

	e := &ApprovalRequiredError{
		Err:        &Error{Code: ErrCodeApprovalRequired, Message: "wait"},
		ApprovalID: "app_1",
	}
	if !errors.Is(e, ErrApprovalRequired) {
		t.Errorf("errors.Is(e, ErrApprovalRequired) failed")
	}
}

func TestMapStatus(t *testing.T) {
	t.Parallel()

	cases := map[int]string{
		401: ErrCodeAuth,
		403: ErrCodeAuth,
		404: ErrCodeNotFound,
		412: ErrCodePolicyDenied,
		422: ErrCodeValidation,
		429: ErrCodeRateLimit,
		500: ErrCodeServer,
		503: ErrCodeServer,
		418: ErrCodeValidation,
		200: ErrCodeUnknown,
	}
	for status, want := range cases {
		if got := mapStatus(status); got != want {
			t.Errorf("mapStatus(%d) = %q, want %q", status, got, want)
		}
	}
}

func TestMapResponseError(t *testing.T) {
	t.Parallel()

	body := []byte(`{"error":{"code":"policy_denied","message":"blocked","policy_id":"pol_1","policy_name":"Daily limit"}}`)
	err := mapResponseError(412, body)
	var pde *PolicyDeniedError
	if !errors.As(err, &pde) {
		t.Fatalf("expected PolicyDeniedError, got %T", err)
	}
	if pde.PolicyID != "pol_1" || pde.PolicyName != "Daily limit" {
		t.Errorf("policy fields = %+v", pde)
	}

	body = []byte(`{"error":{"code":"approval_required","message":"wait","approval_id":"app_42"}}`)
	err = mapResponseError(202, body)
	var ar *ApprovalRequiredError
	if !errors.As(err, &ar) {
		t.Fatalf("expected ApprovalRequiredError, got %T", err)
	}
	if ar.ApprovalID != "app_42" {
		t.Errorf("ApprovalID = %q", ar.ApprovalID)
	}

	err = mapResponseError(500, []byte(`malformed json`))
	var sdkErr *Error
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if sdkErr.Code != ErrCodeServer {
		t.Errorf("code = %q", sdkErr.Code)
	}
}
