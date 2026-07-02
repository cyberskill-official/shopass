package batch

import (
	"context"
)

// matchBottomRules trả các user_id có alert_rule type 'bottom_predicted' đang bật cho SKU.
func (b *Batch) matchBottomRules(ctx context.Context, productID int64) ([]int64, error) {
	rows, err := b.pool.Query(ctx, `
        SELECT user_id FROM alert_rule
        WHERE product_id = $1 AND rule_type = 'bottom_predicted' AND active = true`,
		productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []int64
	for rows.Next() {
		var u int64
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
