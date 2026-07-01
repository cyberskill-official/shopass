package affil

import (
	"net/url"
)

type NetworkTemplate struct {
	BaseURL     string
	TargetParam string
	SubIDParam  string
}

type Product interface {
	TargetURL() string
}

// BuildDeepLink ghép URL affiliate theo template network, nhúng sub_id làm tham số tracking.
// KHÔNG set cookie, KHÔNG chạm domain sàn - chỉ trả chuỗi URL (§1 #8).
func BuildDeepLink(tmpl NetworkTemplate, p Product, subID string) string {
	target := p.TargetURL()
	u, _ := url.Parse(tmpl.BaseURL)
	q := u.Query()
	if tmpl.TargetParam != "" {
		q.Set(tmpl.TargetParam, target)
	}
	if tmpl.SubIDParam != "" {
		q.Set(tmpl.SubIDParam, subID)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
