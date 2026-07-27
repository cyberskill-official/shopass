package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"shopass/services/b2b/internal/report"
)

// SubLookup resolves the caller's B2B subscription from the authenticated org id.
// MUST load entitlement from DB — never trust client-supplied tier (§1 #11).
type SubLookup interface {
	GetByID(ctx context.Context, id int64) (report.Subscription, error)
}

type ReportHandler struct {
	Subs    SubLookup
	Builder *report.Builder
	Metrics *report.Metrics
	Now     func() time.Time
}

func (h *ReportHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/b2b/reports", h.handleReport)
	mux.HandleFunc("GET /v1/b2b/reports/export", h.handleExport)
}

func (h *ReportHandler) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}

func (h *ReportHandler) handleReport(w http.ResponseWriter, r *http.Request) {
	sub, scope, ok := h.authorize(w, r)
	if !ok {
		return
	}
	rep, err := h.Builder.Build(r.Context(), scope)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "build_failed", err.Error(), nil)
		return
	}
	if h.Metrics != nil {
		h.Metrics.Served(sub.Tier)
	}
	writeJSON(w, http.StatusOK, rep)
}

func (h *ReportHandler) handleExport(w http.ResponseWriter, r *http.Request) {
	sub, scope, ok := h.authorize(w, r)
	if !ok {
		return
	}
	e := report.EntitlementFrom(sub)
	if err := report.AssertExport(e); err != nil {
		if h.Metrics != nil {
			h.Metrics.Denied("no_export")
		}
		writeErr(w, http.StatusForbidden, "export_forbidden", "Gói hiện tại không cho phép export. Nâng cấp để tải CSV/JSON.", nil)
		return
	}
	rep, err := h.Builder.Build(r.Context(), scope)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "build_failed", err.Error(), nil)
		return
	}
	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" || format == "json" {
		b, err := report.ExportJSON(rep)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "export_failed", err.Error(), nil)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
		return
	}
	if format == "csv" {
		csv, err := report.ExportCSV(rep)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "export_failed", err.Error(), nil)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csv))
		return
	}
	writeErr(w, http.StatusBadRequest, "bad_format", "format must be json or csv", nil)
}

func (h *ReportHandler) authorize(w http.ResponseWriter, r *http.Request) (report.Subscription, report.ReportScope, bool) {
	orgID, err := strconv.ParseInt(r.Header.Get("X-B2B-Org-Id"), 10, 64)
	if err != nil || orgID <= 0 {
		writeErr(w, http.StatusBadRequest, "bad_org", "missing X-B2B-Org-Id", nil)
		return report.Subscription{}, report.ReportScope{}, false
	}
	sub, err := h.Subs.GetByID(r.Context(), orgID)
	if err != nil {
		if h.Metrics != nil {
			h.Metrics.Denied("inactive")
		}
		writeErr(w, http.StatusPaymentRequired, "payment_required", "subscription not found or inactive", nil)
		return report.Subscription{}, report.ReportScope{}, false
	}
	if err := report.AssertActive(sub, h.now()); err != nil {
		if h.Metrics != nil {
			h.Metrics.Denied("inactive")
		}
		writeErr(w, http.StatusPaymentRequired, "payment_required", "subscription inactive or expired", nil)
		return report.Subscription{}, report.ReportScope{}, false
	}
	scope, err := parseScope(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_params", err.Error(), nil)
		return report.Subscription{}, report.ReportScope{}, false
	}
	if err := report.CheckScope(report.EntitlementFrom(sub), scope); err != nil {
		var se report.ErrScopeExceeded
		if errors.As(err, &se) {
			if h.Metrics != nil {
				h.Metrics.Denied("scope_exceeded")
			}
			writeErr(w, http.StatusForbidden, "scope_exceeded",
				"Gói "+sub.Tier+" vượt giới hạn. Nâng cấp để xem thêm.",
				map[string]any{"field": se.Field, "limit": se.Limit})
			return report.Subscription{}, report.ReportScope{}, false
		}
		writeErr(w, http.StatusForbidden, "forbidden", err.Error(), nil)
		return report.Subscription{}, report.ReportScope{}, false
	}
	return sub, scope, true
}

func parseScope(r *http.Request) (report.ReportScope, error) {
	q := r.URL.Query()
	cats, err := parseInt64List(q.Get("categories"))
	if err != nil || len(cats) == 0 {
		return report.ReportScope{}, errors.New("categories required")
	}
	pfs, err := parseInt16List(q.Get("platforms"))
	if err != nil || len(pfs) == 0 {
		return report.ReportScope{}, errors.New("platforms required")
	}
	from, err := time.Parse("2006-01-02", q.Get("from"))
	if err != nil {
		return report.ReportScope{}, errors.New("from must be YYYY-MM-DD")
	}
	to, err := time.Parse("2006-01-02", q.Get("to"))
	if err != nil {
		return report.ReportScope{}, errors.New("to must be YYYY-MM-DD")
	}
	if !to.After(from) {
		return report.ReportScope{}, errors.New("to must be after from")
	}
	return report.ReportScope{CategoryIDs: cats, PlatformIDs: pfs, From: from.UTC(), To: to.UTC()}, nil
}

func parseInt64List(s string) ([]int64, error) {
	if strings.TrimSpace(s) == "" {
		return nil, errors.New("empty")
	}
	parts := strings.Split(s, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func parseInt16List(s string) ([]int16, error) {
	vals, err := parseInt64List(s)
	if err != nil {
		return nil, err
	}
	out := make([]int16, len(vals))
	for i, v := range vals {
		out[i] = int16(v)
	}
	return out, nil
}
