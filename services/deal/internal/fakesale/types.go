package fakesale

// Verdict là enum đóng cho kết quả phân loại sale ảo (DEC-DEAL-04).
type Verdict string

const (
    SaleAo   Verdict = "SALE_AO"   // giá gốc bị thổi, giảm giả
    SaleXin  Verdict = "SALE_XIN"  // giảm thật, sát đáy lịch sử
    TamDuoc  Verdict = "TAM_DUOC"  // không ảo, cũng chưa phải sale xịn
    Unknown  Verdict = "UNKNOWN"   // dưới 14 ngày dữ liệu (cold-start)
)

const minHistoryPoints = 14
