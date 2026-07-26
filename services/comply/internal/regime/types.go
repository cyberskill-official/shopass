package regime

import "errors"

var (
	ErrCountryNotOpen    = errors.New("regime: country not open")
	ErrUnsupportedRegime = errors.New("regime: unsupported")
)

// RegimeProfile describes a SEA data-protection regime for adapter selection.
// Numeric SLAs for ID/TH stay at VN baseline until counsel confirms deltas.
type RegimeProfile struct {
	Code              string   // VN_PDPL | ID_PDP | TH_PDPA
	BreachWindowHours int      // hours to notify after awareness
	DPIAFilingDays    int      // days to file DPIA when required
	DSARDays          int      // days to fulfill DSAR
	ConsentLanguages  []string // BCP-47 tags for consent copy
	Notes             []string // statute citations / counsel TBD
}

// RegimeAdapter is one country's data-protection profile.
type RegimeAdapter interface {
	Code() string
	Profile() RegimeProfile
}

// Adapter is an alias kept for existing call sites.
type Adapter = RegimeAdapter
