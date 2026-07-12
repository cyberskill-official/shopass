---
fr_id: FR-AFFIL-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-AFFIL-005 đặc tả cashback layering theo mô hình hold-then-release ở mức triển khai được. 9 mệnh đề §1 normative, mỗi mệnh đề có AC tương ứng (§4) và test (§5). Lằn ranh chống farm được khóa rõ: cashback chỉ tính trên `affiliate_conversion` đã `confirmed`, giữ `pending` qua cửa sổ điều tra, hỏi FR-TRUST-005 trước khi release, và clawback khi network thu hồi. Tiền lưu BIGINT VND, idempotent theo `conversion_id`, payout chỉ qua ngưỡng tối thiểu. Schema `cashback_entry` + `payout_request` khớp DATA-MODEL.md. Sau khi sửa một sai sót nhỏ ở chú thích mã (xem ISS-005), FR đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Trả cashback trên tiền chưa có thật
Tính cashback trên click hoặc conversion `pending` mở cửa cho kẻ gian lận tạo đơn rồi hủy để rút sạch. Giải: §1 #1 + DEC-AFFIL-50 chỉ tạo entry khi `status='confirmed'`; AC #2 + test `TestCashback_ConfirmedCreatesPending` (conversion `pending` không tạo entry).

### ISS-002 - Thiếu cửa sổ chống farm
Release ngay khi confirm là lỗ hổng farm vì đơn affiliate có thể bị hủy/hoàn sau đó. Giải: §1 #4 hold-then-release qua `available_at = confirmed_at + investigation_window` + hỏi FR-TRUST-005; AC #5/#6/#7 + `TestRelease_HoldsBeforeWindow` + `TestRelease_FlaggedUserNotReleased`.

### ISS-003 - Clawback khi đơn bị thu hồi
Đơn đã confirm vẫn có thể bị network thu hồi (chuyển về `rejected`). Giải: §1 #5 chuyển entry sang `clawed_back` và trừ khỏi số dư nếu chưa `paid`; AC #8 + dòng tương ứng trong §10.

### ISS-004 - Đếm trùng entry khi postback lặp
Postback network có thể lặp. Giải: §1 #7 `UNIQUE(conversion_id)` + `ON CONFLICT DO NOTHING`; AC #4 + `TestCashback_Idempotent`.

### ISS-005 - Chú thích share_rate mâu thuẫn số học (đã sửa surgical)
§3 `ledger.go` tính `userShare := c.Commission * rate / 100` (số nguyên), nhưng chú thích ghi `// free 0.30, premium 0.50` ám chỉ phân số - mâu thuẫn với phép chia nguyên và với test (commission 100k premium -> user_share 50k cần `rate=50`). Đã sửa chú thích thành "phan tram nguyen: free 30, premium 50" để khớp số học BIGINT và §1 #3 ("free 30%, Premium 50%"); không đổi logic, chỉ làm rõ `rate` là phần trăm nguyên.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 chỉ trên confirmed | #1,#2 | `ledger.go::OnConversionConfirmed` + `TestCashback_ConfirmedCreatesPending` |
| #2 schema cashback_entry | #1 | `0005_cashback_ledger.sql` |
| #3 user_share + kept_margin (int VND) | #3 | `ledger.go` tính + CHECK >= 0 |
| #4 hold-then-release + FR-TRUST-005 | #5,#6,#7 | `release.go::ReleaseDue` + idx + 2 release tests |
| #5 clawback | #8 | trạng thái `clawed_back` + postback rejected |
| #6 ngưỡng payout | #9,#10 | `payout.go` |
| #7 idempotent | #4 | `UNIQUE(conversion_id)` + `TestCashback_Idempotent` |
| #9 disclosure pending/clawback | #12 | `GET /v1/cashback/summary` field `note` + `next_available_at` |
| §3 CHECK tiền | #11 | SQL CHECK `commission`/`user_share` >= 0 |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; lằn ranh chống farm (confirmed-only, hold-then-release, hỏi FR-TRUST-005, clawback) khớp §5.3 và DEC-AFFIL-50/51. Tiền BIGINT VND, idempotent theo conversion. Một sai sót chú thích share_rate đã được sửa surgical (ISS-005) để khớp số học số nguyên. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-AFFIL-005.*
