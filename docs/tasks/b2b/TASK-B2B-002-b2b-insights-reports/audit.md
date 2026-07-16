---
fr_id: TASK-B2B-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-B2B-002 đặc tả lớp báo cáo B2B + subscription gating ở mức triển khai được. 11 mệnh đề §1 normative, mỗi mệnh đề có AC và test. Quy tắc bảo mật cốt lõi - chỉ đọc `market_trend_daily` qua TASK-B2B-001 (ô đã phát hành) - được giữ xuyên suốt builder. Gating server-side theo `b2b_subscription` với 402 (hết hạn) và 403 (vượt entitlement, rõ ràng không câm). Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Đường vòng đọc nguồn raw (đã chốt)
Cám dỗ "làm giàu" báo cáo bằng cách đọc price_snapshot/price_daily raw sẽ phá k-anonymity của TASK-B2B-001. Giải: §1 #1 + DEC-B2B-10 ép builder chỉ gọi `QueryCells` (lọc suppressed); AC #11 + disallowed_tools.

### ISS-002 - Leo thang quyền qua tham số client
Nếu tin tier do client gửi thì ai cũng tự xưng enterprise. Giải: §1 #11 + DEC-B2B-11 nạp entitlement từ bản ghi subscription của caller; test `TestReport_InactiveSubscription_402` + review.

### ISS-003 - Cắt scope im lặng
Trả phần dữ liệu khi vượt gói làm khách tưởng dữ liệu thiếu và ta mất cơ hội nâng cấp. Giải: §1 #5 + DEC-B2B-13 trả 403 + thông điệp; test `TestScope_TooManyCategories_403` + `TestScope_HistoryTooDeep_403`.

### ISS-004 - Nội suy ô suppress
Nội suy số ẩn danh là bịa và có thể tái dựng tín hiệu cá thể. Giải: §1 #4 + DEC-B2B-12 rỗng-có-lý-do; test `TestBuild_AllSuppressed_EmptyWithReason`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 chỉ đọc market_trend_daily | #11 | `builder.go` (QueryCells) |
| #2 schema subscription | #1 | `0002_b2b_subscription.sql` |
| #3 active/expires gating | #2,#3 | `report_handler.go` + 402 test |
| #4 rỗng-có-lý-do | #6 | `TestBuild_AllSuppressed` |
| #5 kẹp scope -> 403 | #4,#5 | `entitlement.go::checkScope` + 2 test |
| #6 export gating | #7,#8 | `export.go` + `TestExport_NoIdentifierColumns` |
| #7 độ tươi | #9 | `builder.go` cached_at/source_computed_at |
| #8 endpoints | #2,#7 | `report_handler.go` + router.go |
| #11 entitlement server-side | #10 | nạp từ DB caller |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/SQL/test backing. Bảo mật (chỉ đọc ô đã phát hành) và doanh thu (gating server-side, 402/403 rõ ràng) đều được kiểm bằng test. Không mệnh đề mồ côi. Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-B2B-002.*
