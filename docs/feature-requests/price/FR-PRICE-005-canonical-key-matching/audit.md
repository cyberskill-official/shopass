---
fr_id: FR-PRICE-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tôi tái thẩm FR-PRICE-005 độc lập từ file `.md` hiện tại. Thuật toán matching `canonical_key` đặc tả ở mức triển khai được: 13 mệnh đề §1 normative (11 MUST + 1 SHOULD + 1 MAY), mỗi mệnh đề testable có AC §4 và test §5. Pipeline bốn tầng rõ: chuẩn hóa title (fold dấu VN + bỏ nhiễu marketing) -> key xác định (SHA-256 cắt 12 hex trên thuộc tính đã sắp xếp) -> fuzzy pg_trgm theo ngưỡng -> hàng đợi duyệt tay. DDL `canonical_review_queue` khớp DATA-MODEL từng cột: `product_id REFERENCES tracked_product(id)`, `candidate_key TEXT`, `confidence REAL CHECK (>=0 AND <=1)`, `status CHECK IN (pending|approved|rejected)`, `UNIQUE(product_id, candidate_key)`. Van an toàn (không bao giờ auto-merge dưới ngưỡng) được đặc tả và có test. Không có defect cần sửa. Score 10/10.

## §2 - Findings

Không phát hiện defect ở lần thẩm độc lập này. Các điểm tôi kiểm chính diện và xác nhận đạt:

- ISS-001 (kiểm, đạt): `confidence` ràng buộc đúng 0..1 qua `CHECK (confidence >= 0 AND confidence <= 1)` ở DDL §3, kiểu `REAL`. Khớp DATA-MODEL (`confidence REAL (0..1)`). AC #11 kiểm CHECK tồn tại.
- ISS-002 (kiểm, đạt): `status` ràng buộc `CHECK (status IN ('pending','approved','rejected'))` + `UNIQUE(product_id, candidate_key)`, khớp DATA-MODEL chính xác.
- ISS-003 (kiểm, đạt): không bao giờ auto-merge dưới ngưỡng. Vùng `[0,60; 0,82)` bắt buộc vào review queue (§1 #7), dưới `0,60` skip (§1 #8); AC #7 + #8 + `TestMatch_LowConfidence_GoesToReviewQueue`. Đây là van an toàn chính (DEC-PRICE-23), bảo vệ moat khỏi giá so sánh sai.
- ISS-004 (kiểm, đạt): key xác định. `CanonicalKey` sắp xếp thuộc tính trước khi hash -> cùng input cho cùng output bất kể thứ tự map; AC #3 + `TestCanonicalKey_Deterministic`. Recompute idempotent (§1 #10, AC #10).
- ISS-005 (kiểm, đạt): pg_trgm + GIN index `idx_tp_title_trgm` trên `tracked_product(title)` để similarity không quét toàn bảng (§1 #11, AC #12 yêu cầu EXPLAIN không Seq Scan). Enqueue `ON CONFLICT DO NOTHING` (§10) idempotent.
- ISS-006 (kiểm, đạt): tiền tệ không xuất hiện ở FR này (chỉ matching), không có rủi ro float trên giá.
- Typography: prose sạch ASCII; glyph mũi tên chỉ trong khối code và khối `text`/`sql` minh họa của §8 (ngoài phạm vi rule typography, đồng nhất house style của exemplar FR-PRICE-002). Đạt rule O.

## §3 - Traceability §1 -> AC -> artefact

| §1 (mệnh đề) | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 chuẩn hóa đa tầng | AC #1 | `normalize.go::Normalize` + `TestNormalize_StripsMarketingNoise` |
| #2 tách brand/model/attr | AC #2 | `normalize.go::Extract` |
| #3 key xác định | AC #3 | `key.go::CanonicalKey` + `TestCanonicalKey_Deterministic` |
| #4 nhóm chính xác trước | AC #4 | `match.go` (key trước, fuzzy sau) |
| #5 fuzzy pg_trgm | AC #5, #12 | `match.go` similarity + `idx_tp_title_trgm` |
| #6 confidence + merge_threshold | AC #5, #13 | `match.go::bestCandidate` + `TestMatch_SameProductDifferentPlatform_Merges` |
| #7 review queue (không auto-merge) | AC #7 | `0005_canonical_review.sql` + `TestMatch_LowConfidence_GoesToReviewQueue` |
| #8 bỏ qua dưới low_threshold | AC #6, #8 | `match.go` Action="skip" + `TestMatch_DifferentProduct_DoesNotMerge` |
| #9 SetCanonicalKey idempotent | AC #9 | `product_repo.go::SetCanonicalKey` |
| #10 recompute idempotent | AC #10 | `Normalize`/`CanonicalKey` thuần (pure) |
| #11 pg_trgm + GIN index | AC #11, #12 | `0005_canonical_review.sql` (CREATE EXTENSION + GIN) |
| #12 OTel metric (SHOULD) | AC #13 | `canon_merge_auto_total` / `canon_review_enqueued_total` |
| #13 embedding phụ (MAY) | - | helper Python ngoài tiến trình, cộng vào confidence |

## §4 - Kết luận

Mọi mệnh đề normative testable có code/SQL/test backing; không mệnh đề mồ côi. DDL `canonical_review_queue` khớp chính xác DATA-MODEL (`confidence REAL 0..1`, `status` CHECK, UNIQUE kép). Van an toàn review queue (không bao giờ auto-merge dưới ngưỡng) được đặc tả rõ và có test, bảo vệ moat đa sàn (§5.6) khỏi rủi ro hiển thị giá sai. Mệnh đề MAY (#13) là tín hiệu phụ, không phải điều kiện normative nên không yêu cầu test. Frontmatter đủ, không "TBD"/"TODO". Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-PRICE-005.*
