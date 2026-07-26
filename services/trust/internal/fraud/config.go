package fraud

// Config holds tunable thresholds (DEC-TRUST-17c) — not hard-coded in logic paths.
type Config struct {
	HoldThreshold int // risk_score >= this → HOLD rewards (not seize)

	VelocityWindowMinutes int
	VelocityRedeemMax     int
	VelocityWeight        int

	GraphClusterMinSize int
	GraphWeight         int

	RuleWeight int
}

func DefaultConfig() Config {
	return Config{
		HoldThreshold:         70,
		VelocityWindowMinutes: 60,
		VelocityRedeemMax:     10,
		VelocityWeight:        40,
		GraphClusterMinSize:   5,
		GraphWeight:           35,
		RuleWeight:            25,
	}
}
