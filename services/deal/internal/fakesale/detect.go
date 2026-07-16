package fakesale

// DetectFakeSale phân loại trạng thái giảm giá theo §3.5(1).
// Hàm thuần, tất định: cùng (hist, current, list) luôn cho cùng Verdict.
// Mọi so ngưỡng dùng integer-safe math trên int64 (DEC-DEAL-03).
func DetectFakeSale(hist []int64, current, list int64) Verdict {
    if len(hist) < minHistoryPoints {
        return Unknown // cold-start, bàn giao TASK-DEAL-002
    }
    median90 := percentile(hist, 50)
    p10 := percentile(hist, 10)
    trailingMin := minInt64(hist)

    // inflated: list_price > median90 * 1.15  ->  list*100 > median90*115
    inflated := list*100 > median90*115
    // not_real_discount: current_price >= median90 * 0.97  ->  current*100 >= median90*97
    notRealDiscount := current*100 >= median90*97

    if inflated && notRealDiscount {
        return SaleAo
    }
    // SALE_XIN: current <= p10 AND current <= trailing_min * 1.02
    if current <= p10 && current*100 <= trailingMin*102 {
        return SaleXin
    }
    return TamDuoc
}
