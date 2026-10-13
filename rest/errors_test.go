package rest

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAPIErrorFormatting(t *testing.T) {
	err := &APIError{
		Code:       ErrCodeUnknownChannel,
		Message:    "Unknown Channel",
		HTTPStatus: http.StatusNotFound,
	}
	expected := "discord rest: http 404: [10003] Unknown Channel"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestAPIErrorHelpers(t *testing.T) {
	notFound := &APIError{HTTPStatus: http.StatusNotFound}
	if !IsNotFound(notFound) {
		t.Fatalf("expected IsNotFound true")
	}
	if IsRateLimited(notFound) {
		t.Fatalf("expected IsRateLimited false")
	}

	rateLimited := &APIError{HTTPStatus: http.StatusTooManyRequests}
	if !IsRateLimited(rateLimited) {
		t.Fatalf("expected IsRateLimited true")
	}

	unauth := &APIError{HTTPStatus: http.StatusUnauthorized}
	if !IsUnauthorized(unauth) {
		t.Fatalf("expected IsUnauthorized true")
	}

	forbidden := &APIError{HTTPStatus: http.StatusForbidden, Code: ErrCodeMissingPermissions}
	if !IsForbidden(forbidden) {
		t.Fatalf("expected IsForbidden true")
	}
	if !HasErrorCode(forbidden, ErrCodeMissingPermissions) {
		t.Fatalf("expected HasErrorCode true")
	}
	if HasErrorCode(forbidden, ErrCodeUnknownChannel) {
		t.Fatalf("expected HasErrorCode false")
	}
	if HasErrorCode(nil, ErrCodeMissingPermissions) {
		t.Fatalf("expected HasErrorCode false on nil error")
	}
}

func TestAPIErrorUnmarshal(t *testing.T) {
	raw := `{"code":50001,"message":"Missing Access"}`
	var apiErr APIError
	if err := json.Unmarshal([]byte(raw), &apiErr); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if apiErr.Code != ErrCodeMissingAccess {
		t.Fatalf("expected code %d, got %d", ErrCodeMissingAccess, apiErr.Code)
	}
	if apiErr.Message != "Missing Access" {
		t.Fatalf("expected message Missing Access, got %s", apiErr.Message)
	}
}

