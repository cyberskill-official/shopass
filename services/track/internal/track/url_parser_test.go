package track

import (
	"testing"
)

func TestParse_Shopee(t *testing.T) {
	it, ok := ParseItemURL("shopee", "https://shopee.vn/Tai-nghe-i.88123.20114455667?sp_atk=x")
	if !ok {
		t.Fatalf("Expected true")
	}
	if it.ShopID != "88123" {
		t.Errorf("Expected 88123, got %s", it.ShopID)
	}
	if it.PlatformItemID != "20114455667" {
		t.Errorf("Expected 20114455667, got %s", it.PlatformItemID)
	}
}

func TestParse_Lazada(t *testing.T) {
	it, ok := ParseItemURL("lazada", "https://www.lazada.vn/products/abc-pro-i7788.html")
	if !ok {
		t.Fatalf("Expected true")
	}
	if it.PlatformItemID != "7788" {
		t.Errorf("Expected 7788, got %s", it.PlatformItemID)
	}
}

func TestParse_TikTok(t *testing.T) {
	it, ok := ParseItemURL("tiktok", "https://www.tiktok.com/view/product/990011")
	if !ok {
		t.Fatalf("Expected true")
	}
	if it.PlatformItemID != "990011" {
		t.Errorf("Expected 990011, got %s", it.PlatformItemID)
	}
}

func TestParse_BadURL(t *testing.T) {
	_, ok := ParseItemURL("shopee", "https://shopee.vn/no-item-id-here")
	if ok {
		t.Errorf("Expected false")
	}
}
