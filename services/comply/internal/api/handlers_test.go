package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/comply/internal/breach"
	"shopass/services/comply/internal/consent"
	"shopass/services/comply/internal/dsar"
)

type mockConsentService struct {
	history []consent.ConsentRecord
	grants  []consent.ConsentRecord
}

func (m *mockConsentService) Grant(ctx context.Context, userID int64, p consent.Purpose, src string, meta consent.ReqMeta) error {
	if p == consent.Purpose("bad") {
		return consent.ErrUnknownPurpose
	}
	m.grants = append(m.grants, consent.ConsentRecord{UserID: userID, PurposeKey: string(p), Granted: true, Source: src})
	return nil
}

func (m *mockConsentService) Withdraw(ctx context.Context, userID int64, p consent.Purpose, src string, meta consent.ReqMeta) error {
	if p == consent.Purpose("bad") {
		return consent.ErrUnknownPurpose
	}
	m.grants = append(m.grants, consent.ConsentRecord{UserID: userID, PurposeKey: string(p), Granted: false, Source: src})
	return nil
}

func (m *mockConsentService) HistoryAll(ctx context.Context, userID int64) ([]consent.ConsentRecord, error) {
	var out []consent.ConsentRecord
	for _, rec := range m.history {
		if rec.UserID == userID {
			out = append(out, rec)
		}
	}
	return out, nil
}

type mockDSARService struct {
	created   []string
	completed []int64
	erased    bool
}

func (m *mockDSARService) CreateRequest(ctx context.Context, userID int64, kind string) (int64, error) {
	m.created = append(m.created, kind)
	return int64(len(m.created)), nil
}

func (m *mockDSARService) MarkCompleted(ctx context.Context, dsarID int64) error {
	m.completed = append(m.completed, dsarID)
	return nil
}

func (m *mockDSARService) Export(ctx context.Context, userID int64) (dsar.ExportBundle, error) {
	return dsar.ExportBundle{
		Account:         dsar.AccountView{UserID: userID, Email: "u@example.test", Locale: "vi-VN"},
		TrackedProducts: []dsar.ProductView{{ID: 10, Platform: "shopee", Name: "A"}},
		GeneratedAt:     time.Unix(1, 0),
	}, nil
}

func (m *mockDSARService) Erase(ctx context.Context, userID int64) (dsar.EraseResult, error) {
	m.erased = true
	return dsar.EraseResult{WishlistDeleted: 2, PaymentsAnonymized: 1, ConsentLogRetained: true, Status: "completed"}, nil
}

type mockBreachService struct {
	opened []breach.BreachInput
}

func (m *mockBreachService) Open(ctx context.Context, in breach.BreachInput) (int64, error) {
	m.opened = append(m.opened, in)
	return int64(len(m.opened)), nil
}

func (m *mockBreachService) Advance(ctx context.Context, id int64, to breach.Status) error {
	if to == "notified_subjects" {
		return breach.ErrInvalidTransition
	}
	return nil
}

func (m *mockBreachService) Close(ctx context.Context, id int64) error {
	if id == 9 {
		return breach.ErrSubjectsNotNotified
	}
	return nil
}

func (m *mockBreachService) Overdue(ctx context.Context) ([]breach.BreachIncident, error) {
	return []breach.BreachIncident{{ID: 7, Summary: "log leak", Severity: "high", Status: "detected"}}, nil
}

func TestConsentGrantRequiresGatewayIdentity(t *testing.T) {
	mux := http.NewServeMux()
	cs := &mockConsentService{}
	RegisterRoutes(mux, cs, &mockDSARService{}, &mockBreachService{}, "test-operator-token")

	req := httptest.NewRequest(http.MethodPost, "/v1/consent/grant", strings.NewReader(`{"purpose":"cart_read","source":"web"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestConsentGrantAndHistory(t *testing.T) {
	mux := http.NewServeMux()
	cs := &mockConsentService{
		history: []consent.ConsentRecord{
			{UserID: 42, PurposeKey: "cart_read", Granted: true},
			{UserID: 99, PurposeKey: "cart_read", Granted: true},
		},
	}
	RegisterRoutes(mux, cs, &mockDSARService{}, &mockBreachService{}, "test-operator-token")

	req := httptest.NewRequest(http.MethodPost, "/v1/consent/grant", strings.NewReader(`{"purpose":"cart_read","source":"extension"}`))
	req.Header.Set("X-User-Id", "42")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, cs.grants, 1)
	require.Equal(t, int64(42), cs.grants[0].UserID)

	req = httptest.NewRequest(http.MethodGet, "/v1/consent/history", nil)
	req.Header.Set("X-User-Id", "42")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "cart_read")
	require.NotContains(t, rec.Body.String(), "99")
}

func TestDSARPortabilityCompletesWithExport(t *testing.T) {
	mux := http.NewServeMux()
	ds := &mockDSARService{}
	RegisterRoutes(mux, &mockConsentService{}, ds, &mockBreachService{}, "test-operator-token")

	req := httptest.NewRequest(http.MethodPost, "/v1/dsar", strings.NewReader(`{"kind":"portability"}`))
	req.Header.Set("X-User-Id", "42")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, []string{"portability"}, ds.created)
	require.Equal(t, []int64{1}, ds.completed)
	require.Contains(t, rec.Body.String(), "tracked_products")
}

func TestDSAREraseExecutesErasePath(t *testing.T) {
	mux := http.NewServeMux()
	ds := &mockDSARService{}
	RegisterRoutes(mux, &mockConsentService{}, ds, &mockBreachService{}, "test-operator-token")

	req := httptest.NewRequest(http.MethodPost, "/v1/dsar", strings.NewReader(`{"kind":"erase"}`))
	req.Header.Set("X-User-Id", "42")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, ds.erased)
	require.Contains(t, rec.Body.String(), "consent_log_retained")
}

func TestBreachRoutes(t *testing.T) {
	mux := http.NewServeMux()
	bs := &mockBreachService{}
	RegisterRoutes(mux, &mockConsentService{}, &mockDSARService{}, bs, "test-operator-token")

	req := httptest.NewRequest(http.MethodPost, "/v1/comply/breach/open", strings.NewReader(`{"summary":"log leak","severity":"high"}`))
	req.Header.Set("X-Operator-Token", "test-operator-token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Len(t, bs.opened, 1)

	req = httptest.NewRequest(http.MethodPost, "/v1/comply/breach/1/advance", strings.NewReader(`{"status":"notified_subjects"}`))
	req.Header.Set("X-Operator-Token", "test-operator-token")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	req = httptest.NewRequest(http.MethodPost, "/v1/comply/breach/9/close", nil)
	req.Header.Set("X-Operator-Token", "test-operator-token")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusConflict, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/v1/comply/breach/overdue", nil)
	req.Header.Set("X-Operator-Token", "test-operator-token")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "log leak")
}

func TestBreachRoutesRejectMissingOperatorToken(t *testing.T) {
	mux := http.NewServeMux()
	bs := &mockBreachService{}
	RegisterRoutes(mux, &mockConsentService{}, &mockDSARService{}, bs, "test-operator-token")

	req := httptest.NewRequest(http.MethodPost, "/v1/comply/breach/open", strings.NewReader(`{"summary":"log leak","severity":"high"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Empty(t, bs.opened)
}

func TestBreachRoutesFailClosedWithoutConfiguredToken(t *testing.T) {
	mux := http.NewServeMux()
	bs := &mockBreachService{}
	RegisterRoutes(mux, &mockConsentService{}, &mockDSARService{}, bs, "")

	req := httptest.NewRequest(http.MethodPost, "/v1/comply/breach/open", strings.NewReader(`{"summary":"log leak","severity":"high"}`))
	req.Header.Set("X-Operator-Token", "anything")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Empty(t, bs.opened)
}
