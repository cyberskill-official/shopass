package track

import (
	"errors"
	"fmt"
)

var validChannels = map[string]bool{"push": true, "email": true, "sms": true}

// ValidateRule kiểm quan hệ rule_type <-> threshold <-> channel
func ValidateRule(ruleType string, threshold *int64, channel []string) error {
	if len(channel) == 0 {
		return errors.New("channel rỗng")
	}
	for _, c := range channel {
		if !validChannels[c] {
			return fmt.Errorf("channel không hợp lệ: %s", c)
		}
	}
	switch ruleType {
	case "price_below":
		if threshold == nil || *threshold <= 0 {
			return errors.New("price_below cần threshold (VND) > 0")
		}
	case "drop_pct":
		if threshold == nil || *threshold < 1 || *threshold > 99 {
			return errors.New("drop_pct cần threshold trong [1,99]")
		}
	case "real_sale", "bottom_predicted":
		if threshold != nil {
			return fmt.Errorf("%s không nhận threshold", ruleType)
		}
	default:
		return fmt.Errorf("rule_type không hợp lệ: %s", ruleType)
	}
	return nil
}
