package fraud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type memDev struct {
	fp    map[string]map[int64]bool
	edges map[[2]int64]bool
}

func (m *memDev) UpsertFingerprint(ctx context.Context, deviceHash string, userID int64) error {
	if m.fp == nil {
		m.fp = map[string]map[int64]bool{}
	}
	if m.fp[deviceHash] == nil {
		m.fp[deviceHash] = map[int64]bool{}
	}
	m.fp[deviceHash][userID] = true
	return nil
}
func (m *memDev) UsersSharingHash(ctx context.Context, deviceHash string) ([]int64, error) {
	var out []int64
	for u := range m.fp[deviceHash] {
		out = append(out, u)
	}
	return out, nil
}
func (m *memDev) UpsertDeviceEdge(ctx context.Context, a, b int64) error {
	if m.edges == nil {
		m.edges = map[[2]int64]bool{}
	}
	m.edges[[2]int64{a, b}] = true
	return nil
}

func TestDevice_SoloNoEdge(t *testing.T) {
	m := &memDev{}
	svc := DeviceService{ServerSalt: []byte("salt"), Edges: m, SharedAccountMin: 2}
	require.NoError(t, svc.RegisterHash(context.Background(), "client-a", 1))
	require.Empty(t, m.edges)
}

func TestDevice_SharedCreatesEdge(t *testing.T) {
	m := &memDev{}
	svc := DeviceService{ServerSalt: []byte("salt"), Edges: m, SharedAccountMin: 2}
	require.NoError(t, svc.RegisterHash(context.Background(), "same", 1))
	require.NoError(t, svc.RegisterHash(context.Background(), "same", 2))
	require.True(t, m.edges[[2]int64{1, 2}])
}
