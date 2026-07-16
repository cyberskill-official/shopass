---
fr_id: TASK-TRACK-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file TASK-TRACK-004 hiện tại. task đặc tả engine kích hoạt alert: đánh giá alert_rule sau mỗi InsertSnapshot written=true, dedup rising-edge, tạo dòng alert (status pending), bàn giao notification (TASK-NOTIF-001) - KHÔNG tự gửi. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Schema alert_fired_state khớp DATA-MODEL (PK alert_rule_id, last_condition_met, last_fired_at). Bốn nhánh rule_type rõ ràng; bottom_predicted có ý bỏ qua (TASK-DEAL-006 lo) tránh bắn trùng. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- Kích hoạt theo sự kiện written=true (§1 #1, DEC-TRACK-30) ăn khớp delta-only TASK-PRICE-002; lọc product_id+active qua idx_ar_eval, không quét toàn bảng.
- drop_pct dùng median7 từ price_daily (§1 #4, DEC-TRACK-34) làm mốc ổn định, tránh nhiễu flash sale.
- real_sale tái dùng DetectFakeSale TASK-DEAL-001 (§1 #5) - một nguồn sự thật sale thật, engine không nhân bản.
- Dedup rising-edge (§1 #7) + reset cạnh (§1 #8) chống spam mỗi snapshot; có TestRisingEdge_NoSpam + ResetAllowsRefire.
- Ranh giới cứng: tạo alert pending rồi enqueue NOTIF, KHÔNG gọi FCM/email/sms (§1 #10); có test DirectSendCount()==0.
- Bền vững lỗi cục bộ (§1 #11): một luật lỗi không hỏng luật khác; enqueue lỗi giữ pending để retry.
- §10 failure-modes 9 hàng không tầm thường (race hai snapshot, median7 thiếu, bắn trùng).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 kích hoạt written=true | #1 | hook price-written |
| #2 lọc product_id+active | #2 | ActiveByProduct (idx_ar_eval) |
| #3 price_below | #3 | conditionMet + TestPriceBelow_Met |
| #4 drop_pct median7 | #4 | Median7d + TestDropPct_Median7Reference |
| #5 real_sale DetectFakeSale | #5 | deal.DetectFakeSale + TestRealSale_OnlyOnSaleXin |
| #6 bỏ bottom_predicted | #6 | TestBottomPredicted_Skipped |
| #7 rising-edge | #7 | dedup.go + TestRisingEdge_NoSpam |
| #8 reset cạnh | #8 | TestRisingEdge_ResetAllowsRefire |
| #9 tạo alert pending | #9 | handoff.CreateAndEnqueue |
| #10 bàn giao NOTIF không tự gửi | #10 | DirectSendCount()==0 |
| #11 bền vững lỗi cục bộ | #11 | log + continue |
| #12 OTel | #12 | alert_fired_total + alert_dedup_skipped_total |

## §4 - Kết luận

Mọi mệnh đề normative có code/test backing; không mệnh đề mồ côi. Ranh giới với NOTIF (gửi), TASK-DEAL-001 (sale thật) và TASK-DEAL-006 (bottom_predicted) rõ ràng, không lấn, không trùng. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit TASK-TRACK-004.*
