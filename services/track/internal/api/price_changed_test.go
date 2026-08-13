package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"shopass/services/track/internal/engine"
)

type stubEvaluator struct {
	calls []engine.Snapshot
	err   error
}

func (s *stubEvaluator) EvaluateForProduct(_ context.Context, snap engine.Snapshot) error {
	s.calls = append(s.calls, snap)
	return s.err
}

func TestPriceChangedRequiresServiceToken(t *testing.T) {
	eval := &stubEvaluator{}
	mux := http.NewServeMux()
	NewPriceChangedHandler(eval, "private-test-token").RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/price-changed", bytes.NewBufferString(`{"product_id":100,"price":199000}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want %d", rr.Code, http.StatusUnauthorized)
	}
	if len(eval.calls) != 0 {
		t.Fatalf("unexpected evaluate calls: %d", len(eval.calls))
	}
}

func TestPriceChangedEvaluatesWhenAuthorized(t *testing.T) {
	eval := &stubEvaluator{}
	mux := http.NewServeMux()
	NewPriceChangedHandler(eval, "private-test-token").RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/price-changed", bytes.NewBufferString(`{"product_id":100,"price":199000}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", "private-test-token")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d want %d", rr.Code, http.StatusNoContent)
	}
	if len(eval.calls) != 1 {
		t.Fatalf("evaluate calls=%d want 1", len(eval.calls))
	}
	if eval.calls[0].ProductID != 100 || eval.calls[0].Price != 199000 {
		t.Fatalf("snapshot=%+v", eval.calls[0])
	}
}

func TestPriceChangedUnavailableWithoutToken(t *testing.T) {
	eval := &stubEvaluator{}
	mux := http.NewServeMux()
	NewPriceChangedHandler(eval, "").RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/price-changed", bytes.NewBufferString(`{"product_id":100,"price":199000}`))
	req.Header.Set("X-Service-Token", "anything")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want %d", rr.Code, http.StatusServiceUnavailable)
	}
	if len(eval.calls) != 0 {
		t.Fatalf("unexpected evaluate calls: %d", len(eval.calls))
	}
}

func TestPriceChangedRejectsTrailingJSON(t *testing.T) {
	eval := &stubEvaluator{}
	mux := http.NewServeMux()
	NewPriceChangedHandler(eval, "private-test-token").RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/internal/v1/price-changed", bytes.NewBufferString(`{"product_id":100,"price":199000}{"x":1}`))
	req.Header.Set("X-Service-Token", "private-test-token")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want %d", rr.Code, http.StatusBadRequest)
	}
	if len(eval.calls) != 0 {
		t.Fatalf("unexpected evaluate calls: %d", len(eval.calls))
	}
}
