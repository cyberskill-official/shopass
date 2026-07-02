package affil

import "fmt"

type Route struct {
	Path                 string
	CreatesAffiliateLink bool
	IncludesDisclosure   bool
}

// AssertSingleAffiliatePath khẳng định đúng MỘT route tạo affiliate link (FR-AFFIL-002)
// và mọi response link đều kèm disclosure không rỗng.
func AssertSingleAffiliatePath(routes []Route) error {
	linkRoutes := 0
	for _, r := range routes {
		if r.CreatesAffiliateLink {
			linkRoutes++
			if !r.IncludesDisclosure {
				return fmt.Errorf("affiliate link route %s missing disclosure", r.Path)
			}
		}
	}
	if linkRoutes != 1 {
		return fmt.Errorf("expected exactly 1 affiliate-link path, found %d (no back-door)", linkRoutes)
	}
	return nil
}
