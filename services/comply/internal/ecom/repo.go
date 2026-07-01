package ecom

import (
	"context"
	"fmt"
)

type Repo struct {
	// mock internal structure for testing since we don't have DB
	txCounts   map[int]int64
	thresholds map[string]int64
	obs        []EcommerceObligation
}

func NewRepo() *Repo {
	return &Repo{
		txCounts:   make(map[int]int64),
		thresholds: map[string]int64{"foreign_platform_yearly_tx": 100000},
		obs: []EcommerceObligation{
			{ObligationKey: "moit_registration", DescriptionVi: "Dang ky/thong bao website TMĐT voi Bo Cong Thuong", SourceLaw: "ND_52_2013", Status: "not_started"},
			{ObligationKey: "affiliate_disclosure", DescriptionVi: "Cong bo quan he affiliate (du thao Luat TMĐT 2025 - cho luat chot)", SourceLaw: "DRAFT_2025", Status: "not_started"},
			{ObligationKey: "livestream_disclosure", DescriptionVi: "Cong bo noi dung livestream thuong mai (du thao 2025 - cho luat chot)", SourceLaw: "DRAFT_2025", Status: "not_started"},
		},
	}
}

func (r *Repo) txCount(ctx context.Context, year int) (int64, error) {
	return r.txCounts[year], nil
}

func (r *Repo) threshold(ctx context.Context, key string) (int64, error) {
	if val, ok := r.thresholds[key]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("threshold not found")
}

func (r *Repo) Obligations(ctx context.Context) ([]EcommerceObligation, error) {
	return r.obs, nil
}

func (r *Repo) MarkObligation(ctx context.Context, key string, status string) error {
	if status != "not_started" && status != "submitted" && status != "approved" && status != "done" && status != "n_a" {
		return fmt.Errorf("invalid status") // CHECK constraint equivalent
	}

	for i, o := range r.obs {
		if o.ObligationKey == key {
			r.obs[i].Status = status
			return nil
		}
	}
	return fmt.Errorf("obligation not found")
}
