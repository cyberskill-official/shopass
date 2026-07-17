package priming

import (
	"context"
	"errors"
	"testing"
)

func TestNoopQueueMakesDeferredPrimingExplicit(t *testing.T) {
	err := NewNoopQueue().EnqueuePriming(context.Background(), 7)
	if !errors.Is(err, ErrDeferred) {
		t.Fatalf("error = %v, want ErrDeferred", err)
	}
}
