---
fr_id: FR-PRICE-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tôi tái thẩm FR-PRICE-003 độc lập từ file `.md` hiện tại. Endpoint `GET /v1/products/{id}/price-history?range=90d` khớp đúng shape §3.7 và đặc tả ở mức triển khai được: 12 mệnh đề §1 normative (11 MUST + 1 SHOULD), mỗi mệnh đề có AC §4 và test §5. Hợp đồng đọc `price_daily` cho phần thân, stitch raw `price_snapshot` cho cái đuôi kể từ `date_trunc('day', now())`, trả `{product_id, range, daily[], tail[]}` không annotation - đúng phân vai với FR-DEAL-003 (chart-data có annotation). Mọi giá là int64 VND suốt đường DB tới JSON. Không có defect cần sửa. Score 10/10.

## §2 - Findings

Không phát hiện defect ở lần thẩm độc lập này. Các điểm tôi kiểm chính diện và xác nhận đạt:

- ISS-001 (kiểm, đạt): phân biệt rõ với FR-DEAL-003. PRICE-003 trả chuỗi giá raw (`daily` + `tail`), KHÔNG annotation; cross-ref ở §7 (FR-DEAL-003 là sibling, dùng chung quy ước range + int64) đúng và không lẫn vai.
- ISS-002 (kiểm, đạt): độ tươi điểm gần nhất giải bằng stitch raw tail (§1 #4, AC #4, `TestPriceHistory_StitchesRawTail`), không chờ cagg refresh hằng giờ.
- ISS-003 (kiểm, đạt): allowlist range `{7d,30d,90d,180d,1y}` + cap chặn quét raw vô giới hạn (§1 #2, #7, AC #2, #7, `TestPriceHistory_BadRange_400`).
- ISS-004 (kiểm, đạt): 404 tách khỏi 200-rỗng qua `ProductExists` (§1 #6, AC #6, `TestPriceHistory_UnknownProduct_404`).
- ISS-005 (kiểm, đạt): kiểu tiền tệ. DTO `DailyPoint`/`TailPoint` dùng `int64` cho `min_p/max_p/close_p/price`; `disallowed_tools` cấm float/string trong JSON; AC #9 + `TestPriceHistory_PayloadShape`. Đồng nhất DEC-PRICE-05.
- Typography: prose sạch ASCII; không glyph mũi tên ngoài code. Đạt rule O.

## §3 - Traceability §1 -> AC -> artefact

| §1 (mệnh đề) | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 route + parse id (400) | AC #1 | `price_history.go::HandlePriceHistory` |
| #2 range allowlist + 400 | AC #2 | `history_query.go::ParseRange` + `TestPriceHistory_BadRange_400` |
| #3 thân từ price_daily | AC #3 | `QueryDailyBody` |
| #4 stitch raw tail | AC #4 | `QueryRawTail` + `TestPriceHistory_StitchesRawTail` |
| #5 JSON shape | AC #5 | `HistoryResponse` + `TestPriceHistory_PayloadShape` |
| #6 404 vs 200-rỗng | AC #6 | `ProductExists` + `TestPriceHistory_UnknownProduct_404` |
| #7 p95 <500ms | AC #7 | cagg thân + 1 chunk đuôi + `price_history_duration_ms` |
| #8 close_p hiển thị | AC #8 | `DailyPoint.CloseP` (min_p/max_p vẫn có) |
| #9 int64 VND | AC #9 | DTO `int64` + `TestPriceHistory_PayloadShape` |
| #10 JWT gateway | AC #10 | `router.go` sau middleware FR-INFRA-001 |
| #11 OTel metric (SHOULD) | AC #11 | `price_history_requests_total{range,status}` |
| #12 sort + Content-Type | AC #12 | `ORDER BY day/ts` + header `application/json; charset=utf-8` |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; không mệnh đề mồ côi. API shape khớp §3.7; phân vai với FR-DEAL-003 đúng; giá luôn int64 VND. Phụ thuộc FR-PRICE-002 (price_daily + price_snapshot) và FR-INFRA-001 (JWT) nêu rõ ở §7. Frontmatter đủ, không "TBD"/"TODO". Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-PRICE-003.*
