package gating

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SQLPlanCatalog maps plan_catalog.id → tier with a small in-process cache.
type SQLPlanCatalog struct {
	pool  *pgxpool.Pool
	mu    sync.RWMutex
	cache map[int16]string
}

func NewSQLPlanCatalog(pool *pgxpool.Pool) *SQLPlanCatalog {
	return &SQLPlanCatalog{pool: pool, cache: make(map[int16]string)}
}

func (c *SQLPlanCatalog) TierOf(planID int16) string {
	c.mu.RLock()
	if t, ok := c.cache[planID]; ok {
		c.mu.RUnlock()
		return t
	}
	c.mu.RUnlock()

	var tier string
	err := c.pool.QueryRow(context.Background(),
		`SELECT tier FROM plan_catalog WHERE id = $1`, planID).Scan(&tier)
	if err != nil || tier == "" {
		return "free"
	}
	c.mu.Lock()
	c.cache[planID] = tier
	c.mu.Unlock()
	return tier
}
