package fraud

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	internal "shopass/services/trust/internal/fraud"
)

type Reason = internal.Reason
type Assessment = internal.Assessment
type Config = internal.Config
type EventCounter = internal.EventCounter
type ClusterSizer = internal.ClusterSizer
type SignalStore = internal.SignalStore
type RewardHolder = internal.RewardHolder
type Engine = internal.Engine
type PGEventCounter = internal.PGEventCounter
type PGClusterSizer = internal.PGClusterSizer
type PGSignalStore = internal.PGSignalStore
type PGAccountLinkStore = internal.PGAccountLinkStore
type PGRewardHolder = internal.PGRewardHolder
type PGDeviceEdges = internal.PGDeviceEdges
type DeviceService = internal.DeviceService

type PGDB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func DefaultConfig() Config {
	return internal.DefaultConfig()
}

func NewEngine(cfg Config, counter EventCounter, edges ClusterSizer, store SignalStore, holder RewardHolder) *Engine {
	return internal.NewEngine(cfg, counter, edges, store, holder)
}

func NewPGEventCounter(db PGDB) *PGEventCounter {
	return internal.NewPGEventCounter(db)
}

func NewPGClusterSizer(db PGDB) *PGClusterSizer {
	return internal.NewPGClusterSizer(db)
}

func NewPGSignalStore(db PGDB) *PGSignalStore {
	return internal.NewPGSignalStore(db)
}

func NewPGAccountLinkStore(db PGDB) *PGAccountLinkStore {
	return internal.NewPGAccountLinkStore(db)
}

func NewPGRewardHolder(db PGDB, log *slog.Logger) *PGRewardHolder {
	return internal.NewPGRewardHolder(db, log)
}

func NewPGDeviceEdges(db PGDB) *PGDeviceEdges {
	return internal.NewPGDeviceEdges(db)
}
