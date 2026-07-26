package regime

// baseline is the VN PDPL numeric SLA set reused by ID/TH until counsel deltas land.
func baseline() RegimeProfile {
	return RegimeProfile{
		BreachWindowHours: 72,
		DPIAFilingDays:    60,
		DSARDays:          30,
	}
}

// DefaultAdapters registers the three open SEA regimes (codes match country_rule seed).
func DefaultAdapters() []RegimeAdapter {
	return []RegimeAdapter{VNPDPL{}, IDPDP{}, THPDPA{}}
}
