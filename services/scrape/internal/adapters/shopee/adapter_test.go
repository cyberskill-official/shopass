package shopee

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"shopass/services/scrape/internal/orchestrator"
)

// Compile-time assertion: ShopeeAdapter satisfies PlatformAdapter (§4 AC1)
var _ orchestrator.PlatformAdapter = (*ShopeeAdapter)(nil)

// --- test helpers ---

// stubTransport returns an http.RoundTripper that always returns the given body/ct.
type stubTransport struct {
	body string
	ct   string
	code int
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: s.code,
		Header:     http.Header{"Content-Type": []string{s.ct}},
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

type stubHTTPClient struct {
	transport *stubTransport
	lastReq   *http.Request
}

func (c *stubHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.lastReq = req
	return c.transport.RoundTrip(req)
}

func newStubClient(body, ct string) *stubHTTPClient {
	return &stubHTTPClient{
		transport: &stubTransport{body: body, ct: ct, code: 200},
	}
}

type farmSpy struct {
	called bool
	snap   orchestrator.PriceSnapshot
	err    error
}

func (f *farmSpy) RenderPrice(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error) {
	f.called = true
	return f.snap, f.err
}

func testJob(productID int64) orchestrator.ScrapeJob {
	return orchestrator.ScrapeJob{
		ProductID:      productID,
		PlatformID:     shopeePlatformID,
		PlatformItemID: "12345:67890",
	}
}

// --- tests ---

// §4 AC6: HTML challenge → adapter falls back to Playwright farm
func TestFetch_HTMLChallenge_FallsBackToFarm(t *testing.T) {
	client := newStubClient("<html>verify</html>", "text/html")
	farm := &farmSpy{snap: orchestrator.PriceSnapshot{ProductID: 90112, Price: 1}}
	a := NewShopeeAdapter("https://shopee.vn", client, farm)

	snap, err := a.Fetch(context.Background(), testJob(90112))
	require.NoError(t, err)
	require.True(t, farm.called, "should have fallen back to Playwright farm")
	require.Equal(t, int64(1), snap.Price)
}

// §4 AC8: adapter does NOT send user cookie
func TestFetch_NoUserCookieSent(t *testing.T) {
	validJSON := `{"error":0,"data":{"item":{"price":5000000000,"stock":1,"historical_sold":0}}}`
	client := newStubClient(validJSON, "application/json")
	a := NewShopeeAdapter("https://shopee.vn", client, nil)

	_, _ = a.Fetch(context.Background(), testJob(1))
	require.NotNil(t, client.lastReq)
	require.Empty(t, client.lastReq.Header.Get("Cookie"), "backend must not send user cookies (DEC-SCRAPE-09)")
}

// §4 AC2: pdpURL builds correct path
func TestPdpURL(t *testing.T) {
	u := pdpURL("https://shopee.vn", 12345, 67890)
	require.Equal(t, "https://shopee.vn/api/v4/pdp/get_pc?item_id=12345&shop_id=67890&detail_level=0", u)
}

// §4 AC5: error != 0 → ErrItemGone (no fallback, item is dead)
func TestFetch_ItemGone_NoFallback(t *testing.T) {
	gone := `{"error":4,"data":{"item":{}}}`
	client := newStubClient(gone, "application/json")
	farm := &farmSpy{}
	a := NewShopeeAdapter("https://shopee.vn", client, farm)

	_, err := a.Fetch(context.Background(), testJob(1))
	require.ErrorIs(t, err, ErrItemGone)
	require.False(t, farm.called, "should NOT fallback for dead items")
}

// §4 AC7: both endpoint and farm fail → error returned (no panic)
func TestFetch_BothFail_ReturnsError(t *testing.T) {
	client := newStubClient("<html>waf</html>", "text/html")
	a := NewShopeeAdapter("https://shopee.vn", client, nil) // farm is nil

	_, err := a.Fetch(context.Background(), testJob(1))
	require.Error(t, err, "should error when no farm available")
}

// Valid JSON → correct snapshot
func TestFetch_ValidJSON_Snapshot(t *testing.T) {
	validJSON := `{"error":0,"data":{"item":{"price":8900000000,"price_before_discount":14900000000,"stock":37,"historical_sold":1240,"flash_sale":{"status":1}}}}`
	client := newStubClient(validJSON, "application/json")
	a := NewShopeeAdapter("https://shopee.vn", client, nil)

	snap, err := a.Fetch(context.Background(), testJob(90112))
	require.NoError(t, err)
	require.Equal(t, int64(89_000), snap.Price)
	require.Equal(t, int64(149_000), *snap.ListPrice)
	require.True(t, snap.FlashSale)
	require.Equal(t, int64(90112), snap.ProductID)
}

// splitRef tests
func TestSplitRef(t *testing.T) {
	item, shop, err := splitRef("12345:67890")
	require.NoError(t, err)
	require.Equal(t, int64(12345), item)
	require.Equal(t, int64(67890), shop)

	_, _, err = splitRef("invalid")
	require.Error(t, err)

	_, _, err = splitRef("abc:def")
	require.Error(t, err)
}
