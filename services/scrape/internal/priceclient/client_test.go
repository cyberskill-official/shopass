package priceclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"shopass/services/scrape/internal/adapters/shopee"
	"shopass/services/scrape/internal/orchestrator"
)

func TestClient_PostsAndParsesWritten(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/price/snapshots", r.URL.Path)
		require.Equal(t, "test-service-token", r.Header.Get("X-Service-Token"))
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"written":true}`))
	}))
	defer srv.Close()

	lp := int64(250000)
	written, err := New(srv.URL, "test-service-token", 5*time.Second).InsertSnapshot(context.Background(),
		orchestrator.PriceSnapshot{ProductID: 100, TS: time.Now(), Price: 199000, ListPrice: &lp, FlashSale: true})
	require.NoError(t, err)
	require.True(t, written)
	require.EqualValues(t, 199000, got["price"])
	require.EqualValues(t, 100, got["product_id"])
	require.EqualValues(t, 250000, got["list_price"])
}

func TestClient_SkipReturnsFalse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"written":false}`))
	}))
	defer srv.Close()
	written, err := New(srv.URL, "test-service-token", 5*time.Second).InsertSnapshot(context.Background(),
		orchestrator.PriceSnapshot{ProductID: 1, Price: 1000, TS: time.Now()})
	require.NoError(t, err)
	require.False(t, written)
}

// --- end-to-end: Shopee JSON -> adapter -> pool -> priceclient -> price server ---

type stubRT struct{ body, ct string }

func (s stubRT) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{s.ct}},
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

type stubHTTP struct{ rt http.RoundTripper }

func (c stubHTTP) Do(r *http.Request) (*http.Response, error) { return c.rt.RoundTrip(r) }

type noopQueue struct{}

func (noopQueue) Enqueue(context.Context, orchestrator.ScrapeJob) error { return nil }
func (noopQueue) Claim(context.Context, int16) (orchestrator.ScrapeJob, bool, error) {
	return orchestrator.ScrapeJob{}, false, nil
}
func (noopQueue) Ack(context.Context, int64) error                               { return nil }
func (noopQueue) Retry(context.Context, orchestrator.ScrapeJob, time.Time) error { return nil }
func (noopQueue) Fail(context.Context, orchestrator.ScrapeJob) error             { return nil }
func (noopQueue) Reclaim(context.Context, int16, time.Duration) (orchestrator.ScrapeJob, bool, error) {
	return orchestrator.ScrapeJob{}, false, nil
}

func TestE2E_ShopeeToPriceIngest(t *testing.T) {
	var postedPrice, postedList int64 = -1, -1
	price := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]any
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &m)
		postedPrice = int64(m["price"].(float64))
		postedList = int64(m["list_price"].(float64))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"written":true}`))
	}))
	defer price.Close()

	// realistic Shopee pdp JSON: price is micro-VND (199000 * 100000)
	body := `{"error":0,"data":{"item":{"price":19900000000,"price_before_discount":25000000000,"stock":10,"historical_sold":5}}}`
	adapter := shopee.NewShopeeAdapter("https://shopee.test", stubHTTP{stubRT{body: body, ct: "application/json"}}, nil)

	pool := orchestrator.NewPool(
		orchestrator.Config{MaxConcurrency: map[int16]int{1: 1}, MaxAttempts: 3, BackoffBaseMs: 10},
		New(price.URL, "test-service-token", 5*time.Second),
		noopQueue{},
	)
	pool.RegisterAdapter(adapter)

	result, err := pool.ProcessJob(context.Background(), orchestrator.ScrapeJob{
		ProductID: 100, PlatformID: 1, PlatformItemID: "555:777", Tier: orchestrator.TierHot,
	})
	require.NoError(t, err)
	require.Equal(t, orchestrator.JobSucceeded, result.Outcome)
	require.EqualValues(t, 199000, postedPrice) // micro-VND parsed to BIGINT VND
	require.EqualValues(t, 250000, postedList)
}
