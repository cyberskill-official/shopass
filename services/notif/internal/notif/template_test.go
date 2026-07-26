package notif

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender_PriceBelow(t *testing.T) {
	r, err := Render("price_below", map[string]any{"price": int64(79_000)})
	require.NoError(t, err)
	require.Contains(t, r.Body, "79.000 VND")
}

func TestRender_MissingVar_Errors(t *testing.T) {
	_, err := Render("price_below", map[string]any{})
	require.Error(t, err)
}

func TestRender_EscapesPayload(t *testing.T) {
	r, _ := Render("price_below", map[string]any{
		"price":        int64(79_000),
		"product_name": "<script>x</script>",
	})
	require.NotContains(t, r.Body, "<script>")
	require.Contains(t, r.Body, "&lt;script&gt;x&lt;/script&gt;")
}

func TestRender_BottomPredicted_FromProbability(t *testing.T) {
	r, err := Render("bottom_predicted", map[string]any{"p_bottom_14d": 0.82})
	require.NoError(t, err)
	require.Contains(t, r.Title, "đáy")
	require.Contains(t, r.Body, "82%")
}
