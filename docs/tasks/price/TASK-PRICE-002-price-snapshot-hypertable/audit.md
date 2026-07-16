---
fr_id: TASK-PRICE-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
authoring_compliance: "khớp house style task CyberOS (§1 BCP-14 normative -> §11 ghi chú + frontmatter đầy đủ)"
---

## §1 - Tóm tắt verdict

TASK-PRICE-002 đặc tả hypertable `price_snapshot` ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC tương ứng và test trong §5. DDL TimescaleDB đầy đủ (hypertable + continuous aggregate + compression + retention). Delta-only - đòn bẩy storage lớn nhất - được đặc tả rõ với hàm `changed()` so đủ trường. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Đơn vị tiền tệ (đã chốt)
Ban đầu chưa nói rõ kiểu dữ liệu giá. Rủi ro: dùng float -> sai số trên phép so `current_price >= median90 * 0.97` của sale ảo. Giải: §1 #3 + DEC-PRICE-05 ép BIGINT VND; AC #7 + #8 + CHECK constraint.

### ISS-002 - Delta-only so sánh thiếu trường
Bản đầu chỉ so `price`. Nếu `flash_sale` flip mà giá bằng -> mất tín hiệu sale. Giải: §1 #4 + `changed()` so `(price, list_price, stock, flash_sale)`; AC #6 + test `TestDelta_FlashFlip_Writes`.

### ISS-003 - Migration cagg khóa bảng lớn
Tạo continuous aggregate trên bảng tỷ-dòng có thể khóa dài lúc deploy. Giải: §3 + §6 + §10 - tạo `WITH NO DATA` rồi để policy backfill.

### ISS-004 - Retention chính sách
Ban đầu chưa phân biệt raw vs aggregate. Giải: §1 #7 - raw 18 tháng, aggregate vô hạn (input cho TASK-DEAL-005 cần >=180 ngày).

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 schema + PK | #1 | `0002_price_snapshot.sql` |
| #2 hypertable chunk 7d | #1,#2 | `create_hypertable` + test `TestHypertable_Exists` |
| #3 BIGINT VND | #7 | CHECK `price > 0` + types.go int64 |
| #4 delta-only | #3,#4,#5,#6 | `delta.go::changed` + 3 delta tests |
| #5 compression | #12 | `0004_compression_policy.sql` |
| #6 continuous aggregate | #11 | `0003_price_daily_cagg.sql` |
| #7 retention | - | `add_retention_policy` 18m |
| #8 repo funcs | #10,#11 | `repo.go` + `QueryRange`/`QueryDaily` |
| #11 ON CONFLICT | #9 | `TestConflict_Idempotent` |
| #12 CHECK constraints | #7,#8 | SQL CHECK + tests |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/SQL/test backing. Không có mệnh đề "mồ côi". **Score = 10/10. Verdict: PASS.** Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-PRICE-002.*
