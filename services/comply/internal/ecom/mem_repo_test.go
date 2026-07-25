package ecom

import (
	"context"
	"fmt"
)

var validObStatus = map[string]bool{
	"not_started": true, "submitted": true, "approved": true, "done": true, "n_a": true,
}

// memRepo is an in-memory store implementation for unit tests. It mirrors the
// seed data in migration 0008_ecommerce_obligation.sql.
type memRepo struct {
	txCounts   map[int]int64
	thresholds map[string]int64
	obs        []EcommerceObligation
}

func newMemRepo() *memRepo {
	return &memRepo{
		txCounts:   map[int]int64{},
		thresholds: map[string]int64{"foreign_platform_yearly_tx": 100000},
		obs: []EcommerceObligation{
			{ID: 1, ObligationKey: "moit_registration", DescriptionVi: "Dang ky/thong bao website TMDT voi Bo Cong Thuong", SourceLaw: "ND_52_2013", Status: "not_started", Version: 1},
			{ID: 2, ObligationKey: "affiliate_disclosure", DescriptionVi: "Cong bo quan he affiliate (du thao Luat TMDT 2025 - cho luat chot)", SourceLaw: "DRAFT_2025", Status: "not_started", Version: 1},
			{ID: 3, ObligationKey: "livestream_disclosure", DescriptionVi: "Cong bo noi dung livestream thuong mai (du thao 2025 - cho luat chot)", SourceLaw: "DRAFT_2025", Status: "not_started", Version: 1},
		},
	}
}

func (m *memRepo) txCount(ctx context.Context, year int) (int64, error) { return m.txCounts[year], nil }
func (m *memRepo) threshold(ctx context.Context, key string) (int64, error) {
	return m.thresholds[key], nil
}

func (m *memRepo) Obligations(ctx context.Context) ([]EcommerceObligation, error) {
	out := make([]EcommerceObligation, len(m.obs))
	copy(out, m.obs)
	return out, nil
}

func (m *memRepo) MarkObligation(ctx context.Context, key, status string) error {
	if !validObStatus[status] {
		return fmt.Errorf("invalid status: %s", status)
	}
	for i := range m.obs {
		if m.obs[i].ObligationKey == key {
			m.obs[i].Status = status
			return nil
		}
	}
	return fmt.Errorf("obligation not found: %s", key)
}
