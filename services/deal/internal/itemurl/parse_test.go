package itemurl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse_Shopee(t *testing.T) {
	got, ok := Parse("https://shopee.vn/Tai-nghe-i.88123.20114455667?sp_atk=x")
	require.True(t, ok)
	require.Equal(t, "shopee", got.Platform)
	require.Equal(t, "20114455667:88123", got.PlatformItemID)
}

func TestParse_RejectsNonShopeeHost(t *testing.T) {
	_, ok := Parse("https://example.com/x-i.88123.20114455667")
	require.False(t, ok)
}

func TestParse_Lazada(t *testing.T) {
	got, ok := Parse("https://www.lazada.vn/products/abc-pro-i7788.html")
	require.True(t, ok)
	require.Equal(t, "lazada", got.Platform)
	require.Equal(t, "7788", got.PlatformItemID)
}

func TestParse_TikTok(t *testing.T) {
	got, ok := Parse("https://www.tiktok.com/view/product/990011")
	require.True(t, ok)
	require.Equal(t, "tiktok", got.Platform)
	require.Equal(t, "990011", got.PlatformItemID)
}
