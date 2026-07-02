package optimizer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr(v int64) *int64 {
	return &v
}

func TestDiscountPercent_IntegerDivision(t *testing.T) {
	v := Voucher{DiscountType: DiscountPercent, DiscountValue: 15}
	require.Equal(t, int64(45_000), discountAmount(v, 300_000)) // 15% của 300k = 45k (chia nguyên)
}

func TestApplyCap(t *testing.T) {
	require.Equal(t, int64(50_000), applyCap(80_000, ptr(50_000))) // kẹp
	require.Equal(t, int64(30_000), applyCap(30_000, ptr(50_000))) // dưới cap giữ nguyên
	require.Equal(t, int64(80_000), applyCap(80_000, nil))         // không cap
}
