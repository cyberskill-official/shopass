package fraud

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// DeviceLinker upserts salted device hashes and creates undirected device edges.
type DeviceEdgeWriter interface {
	UpsertFingerprint(ctx context.Context, deviceHash string, userID int64) error
	UsersSharingHash(ctx context.Context, deviceHash string) ([]int64, error)
	UpsertDeviceEdge(ctx context.Context, a, b int64) error
}

type DeviceService struct {
	ServerSalt []byte
	Edges      DeviceEdgeWriter
	// SharedAccountMin: when >= N users share a hash, emit edges (configurable).
	SharedAccountMin int
}

func (d DeviceService) NormalizeClientHash(clientHash string) string {
	mac := hmac.New(sha256.New, d.ServerSalt)
	mac.Write([]byte(clientHash))
	return hex.EncodeToString(mac.Sum(nil))
}

// RegisterHash stores the salted hash and links multi-account edges. Never auto-bans.
func (d DeviceService) RegisterHash(ctx context.Context, clientHash string, userID int64) error {
	if d.SharedAccountMin <= 0 {
		d.SharedAccountMin = 2
	}
	h := d.NormalizeClientHash(clientHash)
	if err := d.Edges.UpsertFingerprint(ctx, h, userID); err != nil {
		return err
	}
	users, err := d.Edges.UsersSharingHash(ctx, h)
	if err != nil || len(users) < d.SharedAccountMin {
		return err
	}
	for _, other := range users {
		if other == userID {
			continue
		}
		a, b := userID, other
		if a > b {
			a, b = b, a
		}
		if err := d.Edges.UpsertDeviceEdge(ctx, a, b); err != nil {
			return err
		}
	}
	return nil
}
