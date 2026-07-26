package regime

type VNPDPL struct{}

func (VNPDPL) Code() string { return "VN_PDPL" }
func (VNPDPL) Profile() Profile {
	return Profile{Code: "VN_PDPL", BreachHours: 72, DSARDays: 30, Notes: []string{"baseline PDPL VN"}}
}

type IDPDP struct{}

func (IDPDP) Code() string { return "ID_PDP" }
func (IDPDP) Profile() Profile {
	return Profile{Code: "ID_PDP", BreachHours: 72, DSARDays: 14, Notes: []string{"UU PDP Indonesia deltas on VN baseline"}}
}

type THPDPA struct{}

func (THPDPA) Code() string { return "TH_PDPA" }
func (THPDPA) Profile() Profile {
	return Profile{Code: "TH_PDPA", BreachHours: 72, DSARDays: 30, Notes: []string{"Thailand PDPA deltas on VN baseline"}}
}
