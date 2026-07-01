package affil

import (
	"testing"
)

type mockProduct struct {
	targetURL string
}

func (m mockProduct) TargetURL() string {
	return m.targetURL
}

func TestDeepLink_EmbedsSubID(t *testing.T) {
	tmpl := NetworkTemplate{
		BaseURL:     "https://go.involve.asia/aff",
		TargetParam: "url",
		SubIDParam:  "sub_id",
	}
	p := mockProduct{targetURL: "https://shopee.vn/product/88123/20114455667"}
	link := BuildDeepLink(tmpl, p, "sd_abc123")
	if link != "https://go.involve.asia/aff?sub_id=sd_abc123&url=https%3A%2F%2Fshopee.vn%2Fproduct%2F88123%2F20114455667" && link != "https://go.involve.asia/aff?url=https%3A%2F%2Fshopee.vn%2Fproduct%2F88123%2F20114455667&sub_id=sd_abc123" {
		t.Errorf("link does not match expected, got %s", link)
	}
}
