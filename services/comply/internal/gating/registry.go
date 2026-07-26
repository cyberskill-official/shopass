package gating

import (
	"context"
	"sync"
)

type RuleSource interface {
	Load(ctx context.Context) ([]Rule, error)
}

// Registry is a deny-by-default per-country gate matrix (COMPLY-006).
type Registry struct {
	mu    sync.RWMutex
	rules map[string]Rule // key: country|gate
	src   RuleSource
}

func NewRegistry(src RuleSource) *Registry {
	return &Registry{rules: map[string]Rule{}, src: src}
}

func key(country, gate string) string { return country + "|" + gate }

func (r *Registry) Reload(ctx context.Context) error {
	if r.src == nil {
		return nil
	}
	rows, err := r.src.Load(ctx)
	if err != nil {
		return err
	}
	next := make(map[string]Rule, len(rows))
	for _, row := range rows {
		next[key(row.Country, row.GateKey)] = row
	}
	r.mu.Lock()
	r.rules = next
	r.mu.Unlock()
	return nil
}

func (r *Registry) Allow(country, gate string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.rules[key(country, gate)]
	if !ok {
		return false // deny-by-default
	}
	return rule.Allowed
}

func (r *Registry) Value(country, gate string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, ok := r.rules[key(country, gate)]
	if !ok || !rule.Allowed {
		return "", false
	}
	return rule.Value, true
}
