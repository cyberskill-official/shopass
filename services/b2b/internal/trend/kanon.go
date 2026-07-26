package trend

const KAnonymityMin = 50

// Publishable returns false when the cell must be suppressed (k-anonymity).
func Publishable(skuCount int) bool {
	return skuCount >= KAnonymityMin
}
