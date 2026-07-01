package region

type Channel string

const (
	ChannelWeb       Channel = "web"
	ChannelExtension Channel = "extension"
	ChannelCoupon    Channel = "coupon"
	ChannelApp       Channel = "app"
)

const defaultResidency = "ap-southeast-1"

type CountryPolicy struct {
	Country                  string    `yaml:"country"` // ISO-3166 alpha-2
	VoucherStackingAllowed   bool      `yaml:"voucherStackingAllowed"`
	AffiliateChannelsAllowed []Channel `yaml:"affiliateChannelsAllowed"`
	DataResidencyRegion      string    `yaml:"dataResidencyRegion"`
}

// restrictivePolicy là mặc định an toàn cho nước chưa cấu hình (DEC-INFRA-23).
func restrictivePolicy(country string) CountryPolicy {
	return CountryPolicy{
		Country:                  country,
		VoucherStackingAllowed:   false,
		AffiliateChannelsAllowed: nil, // rỗng = không bật kênh nào
		DataResidencyRegion:      defaultResidency,
	}
}
