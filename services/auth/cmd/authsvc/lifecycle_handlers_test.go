package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"shopass/services/auth/internal/auth"
)

type fakeLifecycle struct {
	requestIDs []string
	confirmErr error
	deleteErr  error
	deletedID  int64
}

func (f *fakeLifecycle) RequestReset(_ context.Context, identifier string) error {
	f.requestIDs = append(f.requestIDs, identifier)
	return nil
}

func (f *fakeLifecycle) ConfirmReset(_ context.Context, _, _ string) error {
	return f.confirmErr
}

func (f *fakeLifecycle) DeleteAccount(_ context.Context, userID int64) error {
	f.deletedID = userID
	return f.deleteErr
}

func TestLifecycleHTTP_ResetRequest_Uniform(t *testing.T) {
	fake := &fakeLifecycle{}
	h := &handlers{log: slog.Default(), lifecycle: fake}

	body := `{"identifier":"anyone@x.com"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset-request", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.requestPasswordReset(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var out map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["message"] != resetAckMessage {
		t.Fatalf("message=%q", out["message"])
	}
	if len(fake.requestIDs) != 1 || fake.requestIDs[0] != "anyone@x.com" {
		t.Fatalf("calls=%v", fake.requestIDs)
	}
}

func TestLifecycleHTTP_ConfirmInvalidToken(t *testing.T) {
	fake := &fakeLifecycle{confirmErr: auth.ErrInvalidResetToken}
	h := &handlers{log: slog.Default(), lifecycle: fake}
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset-confirm",
		strings.NewReader(`{"token":"bad","new_password":"newp@ss12345"}`))
	rr := httptest.NewRecorder()
	h.confirmPasswordReset(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestLifecycleHTTP_DeleteRequiresUserHeader(t *testing.T) {
	fake := &fakeLifecycle{}
	h := &handlers{log: slog.Default(), lifecycle: fake}
	req := httptest.NewRequest(http.MethodDelete, "/v1/account", nil)
	rr := httptest.NewRecorder()
	h.deleteAccount(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestLifecycleHTTP_DeleteOK(t *testing.T) {
	fake := &fakeLifecycle{}
	h := &handlers{log: slog.Default(), lifecycle: fake}
	req := httptest.NewRequest(http.MethodDelete, "/v1/account", nil)
	req.Header.Set("X-User-Id", "77")
	rr := httptest.NewRecorder()
	h.deleteAccount(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	if fake.deletedID != 77 {
		t.Fatalf("deletedID=%d", fake.deletedID)
	}
	var out map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["status"] != "deleted" {
		t.Fatalf("out=%v", out)
	}
	grace, _ := out["grace_until"].(string)
	if _, err := time.Parse(time.RFC3339, grace); err != nil {
		t.Fatalf("grace_until: %v", err)
	}
}

func TestLifecycleHTTP_ConfirmWeakPassword(t *testing.T) {
	fake := &fakeLifecycle{confirmErr: auth.ErrWeakPassword}
	h := &handlers{log: slog.Default(), lifecycle: fake}
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset-confirm",
		strings.NewReader(`{"token":"tok","new_password":"short"}`))
	rr := httptest.NewRecorder()
	h.confirmPasswordReset(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestLifecycleHTTP_Unavailable(t *testing.T) {
	h := &handlers{log: slog.Default(), lifecycle: nil}
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset-request",
		strings.NewReader(`{"email":"a@b.co"}`))
	rr := httptest.NewRecorder()
	h.requestPasswordReset(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", rr.Code)
	}
}
