package cashback

import (
	"context"
	"sync"
	"time"
)

// MemStore is an in-memory Store for unit tests.
type MemStore struct {
	mu      sync.Mutex
	byConv  map[int64]Entry
	payouts []PayoutRecord
	nextID  int64
}

type PayoutRecord struct {
	UserID     int64
	Amount     int64
	GatewayRef string
}

func NewMemStore() *MemStore {
	return &MemStore{byConv: map[int64]Entry{}, nextID: 1}
}

func (m *MemStore) InsertPending(ctx context.Context, e Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byConv[e.ConversionID]; ok {
		return nil
	}
	e.ID = m.nextID
	m.nextID++
	if e.Status == "" {
		e.Status = StatusPending
	}
	m.byConv[e.ConversionID] = e
	return nil
}

func (m *MemStore) GetByConversion(ctx context.Context, conversionID int64) (Entry, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.byConv[conversionID]
	return e, ok, nil
}

func (m *MemStore) ListDuePending(ctx context.Context, now time.Time) ([]Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Entry
	for _, e := range m.byConv {
		if e.Status == StatusPending && !e.AvailableAt.After(now) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *MemStore) MarkAvailable(ctx context.Context, conversionID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.byConv[conversionID]
	if !ok || e.Status != StatusPending {
		return nil
	}
	e.Status = StatusAvailable
	m.byConv[conversionID] = e
	return nil
}

func (m *MemStore) MarkClawedBack(ctx context.Context, conversionID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.byConv[conversionID]
	if !ok {
		return nil
	}
	if e.Status == StatusPending || e.Status == StatusAvailable {
		e.Status = StatusClawedBack
		m.byConv[conversionID] = e
	}
	return nil
}

func (m *MemStore) SumAvailable(ctx context.Context, userID int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var sum int64
	for _, e := range m.byConv {
		if e.UserID == userID && e.Status == StatusAvailable {
			sum += e.UserShare
		}
	}
	return sum, nil
}

func (m *MemStore) ListAvailable(ctx context.Context, userID int64) ([]Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Entry
	for _, e := range m.byConv {
		if e.UserID == userID && e.Status == StatusAvailable {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *MemStore) MarkPaid(ctx context.Context, conversionIDs []int64, paidAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range conversionIDs {
		e, ok := m.byConv[id]
		if !ok || e.Status != StatusAvailable {
			continue
		}
		e.Status = StatusPaid
		t := paidAt
		e.PaidAt = &t
		m.byConv[id] = e
	}
	return nil
}

func (m *MemStore) CreatePayoutRequest(ctx context.Context, userID, amount int64, gatewayRef string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.payouts = append(m.payouts, PayoutRecord{UserID: userID, Amount: amount, GatewayRef: gatewayRef})
	return int64(len(m.payouts)), nil
}

func (m *MemStore) Summary(ctx context.Context, userID int64) (UserSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out UserSummary
	out.Note = DisclosureNote
	for _, e := range m.byConv {
		if e.UserID != userID {
			continue
		}
		switch e.Status {
		case StatusPending:
			out.PendingCount++
			out.PendingAmount += e.UserShare
			if out.NextAvailableAt == nil || e.AvailableAt.Before(*out.NextAvailableAt) {
				t := e.AvailableAt
				out.NextAvailableAt = &t
			}
		case StatusAvailable:
			out.AvailableCount++
			out.AvailableAmount += e.UserShare
		case StatusPaid:
			out.PaidTotal += e.UserShare
		}
	}
	return out, nil
}

func (m *MemStore) Payouts() []PayoutRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]PayoutRecord, len(m.payouts))
	copy(out, m.payouts)
	return out
}

var _ Store = (*MemStore)(nil)
