package proxy

import "fmt"

type ProxySession struct {
	URL     string
	User    string
	Pass    string
	Country string
	IP      string
}

// Giả định FingerprintProfile từ thư mục farm. Ta chỉ cần định nghĩa interface nhẹ để bind.
type FingerprintProfile interface {
	GetCountry() string
}

func BindProfile(sess ProxySession, p FingerprintProfile) (ProxySession, error) {
	if sess.Country != p.GetCountry() {
		return ProxySession{}, fmt.Errorf("proxy country %s does not match profile country %s", sess.Country, p.GetCountry())
	}
	return sess, nil
}
