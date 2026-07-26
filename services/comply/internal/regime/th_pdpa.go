package regime

// THPDPA is Thailand Personal Data Protection Act — localization only until counsel SLAs.
type THPDPA struct{}

func (THPDPA) Code() string { return "TH_PDPA" }

func (THPDPA) Profile() RegimeProfile {
	p := baseline()
	p.Code = "TH_PDPA"
	p.ConsentLanguages = []string{"th", "en"}
	p.Notes = []string{
		"Thailand Personal Data Protection Act (PDPA)",
		"Localization declared (consent languages th/en)",
		"Numeric SLAs inherit VN baseline until counsel confirms PDPA-specific deltas",
	}
	return p
}
