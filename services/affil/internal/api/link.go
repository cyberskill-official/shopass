package api

import (
	"context"
	"encoding/json"
	"net/http"

	"shopass/services/affil/internal/auth"
	"shopass/services/affil/internal/affil"
)

type LinkRequest struct {
	ProductID     int64 `json:"product_id"`
	UserInitiated bool  `json:"user_initiated"`
}

type LinkResponse struct {
	DeepLink   string `json:"deep_link"`
	TargetURL  string `json:"target_url"`
	Disclosure string `json:"disclosure"`
}

type ProductService interface {
	Get(ctx context.Context, productID int64) (Product, bool)
}

type Product interface {
	ID() int64
	PlatformID() int16
	TargetURL() string
}

type NetworkService interface {
	TemplateFor(platformID int16) (network string, tmpl affil.NetworkTemplate, ok bool)
}

type AffiliateRepo interface {
	RecordClick(ctx context.Context, c affil.AffiliateClick) (int64, error)
}

type MetricsClient interface {
	LinkRejected(reason string)
	LinkCreated(platformID int16, network string)
}

type Handler struct {
	products ProductService
	networks NetworkService
	repo     AffiliateRepo
	metrics  MetricsClient
}

func NewHandler(products ProductService, networks NetworkService, repo AffiliateRepo, metrics MetricsClient) *Handler {
	return &Handler{
		products: products,
		networks: networks,
		repo:     repo,
		metrics:  metrics,
	}
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (h *Handler) HandleCreateLink(w http.ResponseWriter, req *http.Request) {
	userID := auth.UserID(req.Context())
	var body LinkRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	if !body.UserInitiated {
		h.metrics.LinkRejected("not_user_initiated")
		writeErr(w, 400, "link requires explicit user action")
		return
	}
	tp, ok := h.products.Get(req.Context(), body.ProductID)
	if !ok {
		h.metrics.LinkRejected("product_not_found")
		writeErr(w, 404, "product not found")
		return
	}
	network, tmpl, ok := h.networks.TemplateFor(tp.PlatformID())
	if !ok {
		h.metrics.LinkRejected("network_unavailable")
		writeErr(w, 503, "affiliate network unavailable")
		return
	}
	subID := affil.NewSubID()
	prodID := tp.ID()
	if _, err := h.repo.RecordClick(req.Context(), affil.AffiliateClick{
		UserID:     userID,
		PlatformID: tp.PlatformID(),
		ProductID:  &prodID,
		SubID:      subID,
		Network:    network,
	}); err != nil {
		writeErr(w, 500, "internal error")
		return
	}
	deep := affil.BuildDeepLink(tmpl, tp, subID)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(LinkResponse{
		DeepLink:   deep,
		TargetURL:  tp.TargetURL(),
		Disclosure: affil.Disclosure(),
	})
	h.metrics.LinkCreated(tp.PlatformID(), network)
}
