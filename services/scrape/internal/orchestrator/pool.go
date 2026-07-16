package orchestrator

import (
	"context"
	"time"
)

type Config struct {
	MaxConcurrency map[int16]int
	MaxAttempts    int
	BackoffBaseMs  int
}

type Pool struct {
	cfg      Config
	adapters map[int16]PlatformAdapter
	price    PriceRepo
	queue    Queue
	resched  Rescheduler
	inflight map[int16]chan struct{}
}

// SetRescheduler enables persisted tier-based rescheduling on successful scrapes.
func (p *Pool) SetRescheduler(r Rescheduler) { p.resched = r }

func NewPool(cfg Config, price PriceRepo, q Queue) *Pool {
	p := &Pool{
		cfg:      cfg,
		adapters: make(map[int16]PlatformAdapter),
		price:    price,
		queue:    q,
		inflight: make(map[int16]chan struct{}),
	}
	for pid, cap := range cfg.MaxConcurrency {
		p.inflight[pid] = make(chan struct{}, cap)
	}
	return p
}

func (p *Pool) RegisterAdapter(a PlatformAdapter) {
	p.adapters[a.PlatformID()] = a
}

func (p *Pool) ProcessJob(ctx context.Context, job ScrapeJob) error {
	sem := p.inflight[job.PlatformID]
	if sem != nil {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	err := p.runOne(ctx, job)
	if err == nil {
		return p.queue.Ack(ctx, job.ProductID)
	}
	return err
}

func (p *Pool) runOne(ctx context.Context, job ScrapeJob) error {
	a := p.adapters[job.PlatformID]
	if a == nil {
		return p.scheduleRetry(ctx, job, nil) // or error no adapter
	}
	snap, err := a.Fetch(ctx, job)
	if err != nil {
		return p.scheduleRetry(ctx, job, err) // tăng attempts + backoff
	}
	written, err := p.price.InsertSnapshot(ctx, snap) // delta-only (TASK-PRICE-002)
	if err != nil {
		return p.scheduleRetry(ctx, job, err)
	}
	nextTier := ReTier(job.Tier, written, snap.FlashSale)
	return p.commit(ctx, job.ProductID, nextTier, NextRunAt(nextTier, time.Now()))
}

func (p *Pool) scheduleRetry(ctx context.Context, job ScrapeJob, err error) error {
	// In a real implementation this would write to DB/Queue to delay next execution
	return nil
}

func (p *Pool) commit(ctx context.Context, productID int64, tier Tier, nextRun time.Time) error {
	if p.resched != nil {
		return p.resched.Reschedule(ctx, productID, tier, nextRun)
	}
	return nil
}
