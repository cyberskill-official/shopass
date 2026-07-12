---
fr_id: FR-PRICE-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tôi tái thẩm FR-PRICE-001 độc lập từ file `.md` hiện tại, không tin audit cũ. Bảng `tracked_product` được đặc tả ở mức triển khai được: 12 mệnh đề §1 normative (11 MUST + 1 SHOULD), mỗi mệnh đề ánh xạ tới AC §4 và test §5. DDL §3 khớp DATA-MODEL và §3.4 nguồn từng cột: `id BIGSERIAL PK`, `platform_id SMALLINT REFERENCES platform(id)`, `platform_item_id TEXT NOT NULL`, `canonical_key TEXT` nullable, `first_seen TIMESTAMPTZ DEFAULT now()`, `UNIQUE(platform_id, platform_item_id)`, `idx_tp_canonical`. Ranh giới với FR-PRICE-005 sạch: FR này để `canonical_key = NULL`, không chạy matching. Không có defect cần sửa. Score 10/10.

## §2 - Findings

Không phát hiện defect ở lần thẩm độc lập này. Các điểm tôi kiểm chính diện và xác nhận đạt:

- ISS-001 (kiểm, đạt): khóa duy nhất là kép `(platform_id, platform_item_id)`, không phải toàn cục trên `platform_item_id`. Đúng vì item_id chỉ duy nhất trong một sàn (§1 #3, AC #3, test `TestUnique_PlatformItem`).
- ISS-002 (kiểm, đạt): nhánh `DO UPDATE` chỉ chạm `shop_id/title/category_id`, KHÔNG ghi đè `first_seen`. Mốc cold-start bất biến qua upsert (§1 #9, AC #8, assert `a.FirstSeen == b.FirstSeen`).
- ISS-003 (kiểm, đạt): `canonical_key` để NULL khi insert; FR-PRICE-005 mới UPDATE. AC #4 + test `TestCanonicalKey_NullOnInsert` chốt ranh giới.
- ISS-004 (kiểm, đạt): kiểu cột không lệch DATA-MODEL. `platform_id` là SMALLINT (khớp FK tới `platform.id` SMALLINT), `CategoryID *int64` cho `category_id BIGINT`. Tiền tệ không xuất hiện ở bảng này nên không có rủi ro float.
- Typography: không có em-dash, en-dash, dấu nháy cong hay glyph mũi tên trong prose; mũi tên chỉ nằm trong khối code (ngoài phạm vi rule typography). Đạt rule O.

## §3 - Traceability §1 -> AC -> artefact

| §1 (mệnh đề) | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 schema + PK + cột | AC #1 | `0001_tracked_product.sql` (CREATE TABLE) |
| #2 FK platform_id | AC #2 | `REFERENCES platform(id)` + AC #2 lỗi FK |
| #3 UNIQUE kép | AC #3 | `UNIQUE(platform_id, platform_item_id)` + `TestUnique_PlatformItem` |
| #4 canonical_key NULL on insert | AC #4 | `Upsert` không nhận key + `TestCanonicalKey_NullOnInsert` |
| #5 index idx_tp_canonical | AC #5 | `CREATE INDEX idx_tp_canonical` + AC #5 pg_indexes |
| #6 first_seen DEFAULT now() | AC #1, #6 | DDL DEFAULT + `TestUpsert_New` |
| #7 Upsert(ctx, p) | AC #6, #7 | `product_repo.go::Upsert` (RETURNING) |
| #8 ON CONFLICT idempotent | AC #7 | `TestUpsert_Conflict_Idempotent` (cùng id, 1 dòng) |
| #9 first_seen bất biến | AC #8 | DO UPDATE bỏ qua first_seen + assert test |
| #10 GetByID / FindByPlatformItem | AC #9, #10 | `product_repo.go` hai hàm tra cứu |
| #11 GetByCanonicalKey + guard rỗng | AC #11 | `GetByCanonicalKey` + `TestGetByCanonicalKey` |
| #12 OTel metric (SHOULD) | AC #12 | `metrics.ProductUpsert` |

## §4 - Kết luận

Mọi mệnh đề normative có DDL/code/test backing; không mệnh đề mồ côi. Schema khớp chính xác DATA-MODEL và §3.4. Ranh giới với FR-PRICE-002 (đích FK) và FR-PRICE-005 (điền canonical_key) rạch ròi. Frontmatter đủ khóa, `new_files`/`sub_tasks`/`risk_if_skipped` không rỗng, không "TBD"/"TODO". Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-PRICE-001.*
