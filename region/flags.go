package region

// Lookup luôn cần country; không có overload thiếu country (DEC-INFRA-24, §1 #7).
func (r *Registry) Lookup(flag string, country string) bool {
	if country == "" {
		return false // thiếu country -> an toàn = false
	}
	return r.flagFor(flag, country) // override runtime > mặc định file
}

func (r *Registry) flagFor(flag, country string) bool {
	if f, ok := r.flags[flag]; ok {
		if val, ok := f[country]; ok {
			return val
		}
	}
	return false
}

// SetFlagForTest is a helper for testing
func (r *Registry) SetFlagForTest(flag string, values map[string]bool) {
	if r.flags == nil {
		r.flags = map[string]map[string]bool{}
	}
	r.flags[flag] = values
}
