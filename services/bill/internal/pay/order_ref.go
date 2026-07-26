package pay

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseOrderRef extracts user id and plan tier from NewOrderRef output.
func ParseOrderRef(orderRef string) (userID int64, planTier string, err error) {
	// order_<userID>_<tier> — tier may contain underscores (premium_basic).
	if !strings.HasPrefix(orderRef, "order_") {
		return 0, "", fmt.Errorf("invalid order_ref")
	}
	rest := strings.TrimPrefix(orderRef, "order_")
	i := strings.IndexByte(rest, '_')
	if i <= 0 || i == len(rest)-1 {
		return 0, "", fmt.Errorf("invalid order_ref")
	}
	userID, err = strconv.ParseInt(rest[:i], 10, 64)
	if err != nil || userID <= 0 {
		return 0, "", fmt.Errorf("invalid order_ref user")
	}
	return userID, rest[i+1:], nil
}
