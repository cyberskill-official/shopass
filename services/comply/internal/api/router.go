package api

import (
	"context"
	"net/http"

	"shopass/services/comply/internal/breach"
	"shopass/services/comply/internal/consent"
	"shopass/services/comply/internal/dsar"
)

type ConsentService interface {
	Grant(ctx context.Context, userID int64, p consent.Purpose, src string, m consent.ReqMeta) error
	Withdraw(ctx context.Context, userID int64, p consent.Purpose, src string, m consent.ReqMeta) error
	HistoryAll(ctx context.Context, userID int64) ([]consent.ConsentRecord, error)
}

type DSARService interface {
	CreateRequest(ctx context.Context, userID int64, kind string) (int64, error)
	MarkCompleted(ctx context.Context, dsarID int64) error
	Export(ctx context.Context, userID int64) (dsar.ExportBundle, error)
	Erase(ctx context.Context, userID int64) (dsar.EraseResult, error)
}

type BreachService interface {
	Open(ctx context.Context, in breach.BreachInput) (int64, error)
	Advance(ctx context.Context, id int64, to breach.Status) error
	Close(ctx context.Context, id int64) error
	Overdue(ctx context.Context) ([]breach.BreachIncident, error)
}

func RegisterRoutes(mux *http.ServeMux, consentSvc ConsentService, dsarSvc DSARService, breachSvc BreachService) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	NewConsentHandler(consentSvc).RegisterRoutes(mux)
	NewDSARHandler(dsarSvc).RegisterRoutes(mux)
	NewBreachHandler(breachSvc).RegisterRoutes(mux)
}
