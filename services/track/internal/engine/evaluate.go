package engine

import (
	"context"
	"fmt"
	"log"

	"shopass/services/track/internal/track"
)

type Snapshot struct {
	ProductID int64
	Price     int64 // VND
	ListPrice *int64
}

// DealVerdict matches the type expected from FR-DEAL-001
type DealVerdict string

const (
	SaleXin  DealVerdict = "SALE_XIN"
	SaleAo   DealVerdict = "SALE_AO"
	TamDuoc  DealVerdict = "TAM_DUOC"
	Unknown  DealVerdict = "UNKNOWN"
)

type DealService interface {
	DetectFakeSale(ctx context.Context, productID int64, price int64, listPrice *int64) (DealVerdict, error)
}

type PriceService interface {
	Median7d(ctx context.Context, productID int64) (int64, error)
}

type RuleRepo interface {
	ActiveByProduct(ctx context.Context, productID int64) ([]track.AlertRule, error)
}

type Engine struct {
	rules   RuleRepo
	price   PriceService
	deal    DealService
	state   StateRepo
	handoff HandoffService
}

func NewEngine(rules RuleRepo, price PriceService, deal DealService, state StateRepo, handoff HandoffService) *Engine {
	return &Engine{
		rules:   rules,
		price:   price,
		deal:    deal,
		state:   state,
		handoff: handoff,
	}
}

// EvaluateForProduct đánh giá mọi luật active của một SKU sau khi giá đổi (DEC-TRACK-30).
func (e *Engine) EvaluateForProduct(ctx context.Context, snap Snapshot) error {
	rules, err := e.rules.ActiveByProduct(ctx, snap.ProductID)
	if err != nil {
		return err
	}
	for _, r := range rules {
		met, payload, err := e.conditionMet(ctx, r, snap)
		if err != nil {
			log.Printf("WARN eval rule lỗi rule_id=%d err=%v", r.ID, err) // §1 #11
			continue
		}
		if err := e.fireIfRisingEdge(ctx, r, met, payload); err != nil {
			log.Printf("WARN fire lỗi rule_id=%d err=%v", r.ID, err)
		}
	}
	return nil
}

func (e *Engine) conditionMet(ctx context.Context, r track.AlertRule, s Snapshot) (bool, map[string]any, error) {
	switch r.RuleType {
	case "price_below":
		met := s.Price <= *r.Threshold
		return met, map[string]any{"price": s.Price, "threshold": *r.Threshold}, nil
	case "drop_pct":
		ref, err := e.price.Median7d(ctx, s.ProductID)
		if err != nil {
			return false, nil, err
		}
		limit := ref * (100 - *r.Threshold) / 100
		return s.Price <= limit, map[string]any{"price": s.Price, "ref_price": ref}, nil
	case "real_sale":
		verdict, err := e.deal.DetectFakeSale(ctx, s.ProductID, s.Price, s.ListPrice)
		if err != nil {
			return false, nil, err
		}
		return verdict == SaleXin, map[string]any{"price": s.Price, "verdict": verdict}, nil
	case "bottom_predicted":
		return false, nil, nil // DEC-TRACK-31: do FR-DEAL-006 lo
	default:
		return false, nil, fmt.Errorf("rule_type lạ: %s", r.RuleType)
	}
}
