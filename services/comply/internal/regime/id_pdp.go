package regime

// IDPDP is Indonesia Personal Data Protection Law — localization only until counsel SLAs.
type IDPDP struct{}

func (IDPDP) Code() string { return "ID_PDP" }

func (IDPDP) Profile() RegimeProfile {
	p := baseline()
	p.Code = "ID_PDP"
	p.ConsentLanguages = []string{"id", "en"}
	p.Notes = []string{
		"UU PDP (Undang-Undang Perlindungan Data Pribadi) Indonesia",
		"Localization declared (consent languages id/en)",
		"Numeric SLAs inherit VN baseline until counsel confirms PDP-specific deltas",
	}
	return p
}
