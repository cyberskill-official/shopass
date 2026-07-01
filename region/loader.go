package region

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Registry struct {
	byCountry map[string]CountryPolicy
	flags     map[string]map[string]bool // [flagName][country]
}

func Load(path string) (*Registry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Countries []CountryPolicy `yaml:"countries"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	reg := &Registry{
		byCountry: map[string]CountryPolicy{},
		flags:     map[string]map[string]bool{},
	}
	for _, p := range doc.Countries {
		if !isAlpha2(p.Country) {
			return nil, fmt.Errorf("country không hợp lệ: %q", p.Country)
		}
		if _, dup := reg.byCountry[p.Country]; dup {
			return nil, fmt.Errorf("trùng country: %s", p.Country) // §1 #10
		}
		if err := validateChannels(p.AffiliateChannelsAllowed); err != nil {
			return nil, err
		}
		reg.byCountry[p.Country] = p
	}
	return reg, nil
}

// Policy trả policy của nước; nước chưa cấu hình -> mặc định hạn chế nhất.
func (r *Registry) Policy(country string) CountryPolicy {
	if p, ok := r.byCountry[country]; ok {
		return p
	}
	return restrictivePolicy(country) // §1 #5
}

func isAlpha2(country string) bool {
	if len(country) != 2 {
		return false
	}
	for _, r := range country {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func validateChannels(channels []Channel) error {
	for _, c := range channels {
		switch c {
		case ChannelWeb, ChannelExtension, ChannelCoupon, ChannelApp:
			continue
		default:
			return fmt.Errorf("channel không hợp lệ: %s", c)
		}
	}
	return nil
}
