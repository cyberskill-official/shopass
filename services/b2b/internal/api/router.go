package api

import "net/http"

// RegisterRoutes mounts B2B HTTP surfaces.
func RegisterRoutes(mux *http.ServeMux, reportH *ReportHandler) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	if reportH != nil {
		reportH.Register(mux)
	}
}
