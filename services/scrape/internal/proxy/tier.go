package proxy

type Tier string

const (
	TierEnterprise Tier = "enterprise" // Bright Data, Oxylabs (~8,5-12 USD/GB)
	TierMid        Tier = "mid"        // Decodo, SOAX, NetNut (~3-6 USD/GB)
	TierBudget     Tier = "budget"     // IPRoyal (~1,75 USD/GB)
)

type TargetDifficulty int

const (
	DiffShopeeJSON TargetDifficulty = iota // dễ: internal JSON
	DiffAkamai                             // Lazada
	DiffByteDance                          // TikTok attestation
	DiffUnknown
)

// SelectTier chọn tier proxy theo độ khó target (DEC-SCRAPE-17).
func SelectTier(d TargetDifficulty) Tier {
	switch d {
	case DiffAkamai, DiffByteDance:
		return TierEnterprise // độ tin cậy cao nhất cho WAF mạnh
	case DiffShopeeJSON:
		return TierBudget // JSON dễ -> tiết kiệm
	default:
		return TierMid
	}
}
