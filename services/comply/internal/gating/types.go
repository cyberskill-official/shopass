package gating

import "fmt"

const (
	GateVoucherStacking      = "voucher_stacking"
	GateAffiliateChannel     = "affiliate_channel"
	GateDataProtectionRegime = "data_protection_regime"
)

var knownCountries = map[string]struct{}{
	"VN": {},
	"ID": {},
	"TH": {},
	"PH": {},
	"MY": {},
	"SG": {},
	"TW": {},
}

var knownGateKeys = map[string]struct{}{
	GateVoucherStacking:      {},
	GateAffiliateChannel:     {},
	GateDataProtectionRegime: {},
}

type Rule struct {
	Country string
	GateKey string
	Allowed bool
	Value   string
	Version int
}

func ValidateRule(rule Rule) error {
	if _, ok := knownCountries[rule.Country]; !ok {
		return fmt.Errorf("unknown country %q", rule.Country)
	}
	if _, ok := knownGateKeys[rule.GateKey]; !ok {
		return fmt.Errorf("unknown gate %q", rule.GateKey)
	}
	return nil
}
