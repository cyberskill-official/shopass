package track

import (
	"testing"
)

func ptr(v int64) *int64 { return &v }

func TestValidate_PriceBelow(t *testing.T) {
	if err := ValidateRule("price_below", ptr(int64(89_000)), []string{"push"}); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if err := ValidateRule("price_below", nil, []string{"push"}); err == nil {
		t.Errorf("Expected error for missing threshold")
	}
	if err := ValidateRule("price_below", ptr(int64(0)), []string{"push"}); err == nil {
		t.Errorf("Expected error for <= 0")
	}
}

func TestValidate_DropPct(t *testing.T) {
	if err := ValidateRule("drop_pct", ptr(int64(20)), []string{"push"}); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if err := ValidateRule("drop_pct", ptr(int64(500)), []string{"push"}); err == nil {
		t.Errorf("Expected error for > 99")
	}
}

func TestValidate_SignalRules_NoThreshold(t *testing.T) {
	if err := ValidateRule("real_sale", nil, []string{"push"}); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
	if err := ValidateRule("real_sale", ptr(int64(10)), []string{"push"}); err == nil {
		t.Errorf("Expected error for non-nil threshold")
	}
	if err := ValidateRule("bottom_predicted", nil, []string{"email"}); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestValidate_Channel(t *testing.T) {
	if err := ValidateRule("real_sale", nil, []string{"telegram"}); err == nil {
		t.Errorf("Expected error for unknown channel")
	}
	if err := ValidateRule("real_sale", nil, []string{}); err == nil {
		t.Errorf("Expected error for empty channel")
	}
}
