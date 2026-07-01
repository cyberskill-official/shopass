package shopee

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	return data
}

// §5 test 1: parse fixture JSON → PriceSnapshot correct
func TestParse_ValidPDP(t *testing.T) {
	raw := readFixture(t, "pdp_get_pc.json")
	var resp pdpResponse
	require.NoError(t, json.Unmarshal(raw, &resp))
	snap, err := resp.toSnapshot(90112, t0)
	require.NoError(t, err)
	require.Equal(t, int64(89_000), snap.Price)        // 8900000000 / 100000
	require.Equal(t, int64(149_000), *snap.ListPrice)   // 14900000000 / 100000
	require.True(t, snap.FlashSale)
	require.Equal(t, int32(37), *snap.Stock)
	require.Equal(t, int32(1240), *snap.Sold)
	require.Equal(t, int64(90112), snap.ProductID)
	require.Equal(t, t0, snap.TS)
}

// §5 test 2: integer division precision — no float error
func TestParse_IntegerDivision_NoFloatError(t *testing.T) {
	var resp pdpResponse
	resp.Data.Item.Price = 333_333_00000 // 333_333 VND exact
	snap, err := resp.toSnapshot(1, t0)
	require.NoError(t, err)
	require.Equal(t, int64(333_333), snap.Price)
}

// §5 test 3: item gone
func TestParse_ItemGone(t *testing.T) {
	resp := pdpResponse{Error: 4}
	_, err := resp.toSnapshot(1, t0)
	require.ErrorIs(t, err, ErrItemGone)
}

// No flash sale → FlashSale == false
func TestParse_NoFlashSale(t *testing.T) {
	var resp pdpResponse
	resp.Data.Item.Price = 100_000_00000
	snap, err := resp.toSnapshot(1, t0)
	require.NoError(t, err)
	require.False(t, snap.FlashSale)
}

// No list_price → ListPrice == nil
func TestParse_NoListPrice(t *testing.T) {
	var resp pdpResponse
	resp.Data.Item.Price = 50_000_00000
	snap, err := resp.toSnapshot(1, t0)
	require.NoError(t, err)
	require.Nil(t, snap.ListPrice)
}
