package engine

import (
	"context"

	"shopass/services/track/internal/track"
)

type HandoffService interface {
	CreateAndEnqueue(ctx context.Context, r track.AlertRule, payload map[string]any) (int64, error)
}
