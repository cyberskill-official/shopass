package gating

const (
	GateVoucherStacking       = "voucher_stacking"
	GateAffiliateChannel      = "affiliate_channel"
	GateDataProtectionRegime  = "data_protection_regime"
)

type Rule struct {
	Country string
	GateKey string
	Allowed bool
	Value   string
	Version int
}
