package canon

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize_FoldsDiacritics(t *testing.T) {
	got := Normalize("Điện Thoại Thông Minh")
	require.Equal(t, "dien thoai thong minh", got)
}

func TestNormalize_StripsMarketingNoise(t *testing.T) {
	// \U0001F525 là một emoji thật (lửa) nhúng vào chuỗi để chứng minh Normalize bóc emoji;
	// nguồn giữ ASCII thuần, runtime vẫn nhận đúng rune emoji.
	got := Normalize("[CHÍNH HÃNG] Tai nghe Sony WH-1000XM5 Freeship Giảm Sốc \U0001F525 - Shop ABC")
	require.Equal(t, "tai nghe sony wh 1000xm5", got)

	// TASK-PRICE-005 AC #1
	got2 := Normalize("[CHÍNH HÃNG] Điện Thoại iPhone 15 Freeship ")
	require.Equal(t, "dien thoai iphone 15", got2)
}

func TestCanonicalKey_Deterministic(t *testing.T) {
	a := CanonicalKey("apple", "iphone 15", map[string]string{"capacity": "128gb", "color": "blue"})
	b := CanonicalKey("apple", "iphone 15", map[string]string{"color": "blue", "capacity": "128gb"})
	require.Equal(t, a, b) // thứ tự map không đổi key
}

func TestExtract(t *testing.T) {
	// 2. Extract của title điện thoại trả Brand="apple" (qua từ điển), Model="iphone 15", Salient chứa color/capacity nếu có.
	attrs := Extract("dien thoai iphone 15 128gb")
	require.Equal(t, "apple", attrs.Brand)
	require.Equal(t, "iphone 15", attrs.Model)
	require.Equal(t, "128gb", attrs.Salient["capacity"])
}
