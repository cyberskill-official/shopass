package regime

// VNPDPL is Vietnam Luật Bảo vệ dữ liệu cá nhân (Luật 91/2025) + NĐ 356/2025.
type VNPDPL struct{}

func (VNPDPL) Code() string { return "VN_PDPL" }

func (VNPDPL) Profile() RegimeProfile {
	p := baseline()
	p.Code = "VN_PDPL"
	p.ConsentLanguages = []string{"vi", "en"}
	p.Notes = []string{
		"Luật 91/2025 Bảo vệ dữ liệu cá nhân (PDPL)",
		"Nghị định 356/2025 hướng dẫn thi hành",
		"Baseline SEA adapter; DSAR 30 ngày, breach 72 giờ, DPIA 60 ngày",
	}
	return p
}
