---
fr_id: FR-TRACK-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file FR-TRACK-003 hiện tại. FR đặc tả schema + API alert_rule + bảng alert. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Kiểm khớp rubric chuyên biệt: rule_type CHECK in {price_below, drop_pct, real_sale, bottom_predicted} (§1 #2, DDL §3); channel TEXT[] mỗi phần tử in {push, email, sms} (§1 #3); threshold BIGINT có chủ đích (§1 #4, DEC-TRACK-22) - khác NUMERIC §3.4 nguồn, ghi chú rõ KHÔNG đổi về NUMERIC, đồng nhất DATA-MODEL. Partial index WHERE active=true khớp truy vấn engine FR-TRACK-004. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- rule_type 4 loại chốt, CHECK ở DB (tuyến cuối) + validate API (tuyến đầu); rác bị chặn hai tầng.
- Ngữ nghĩa threshold đổi theo rule_type (VND / phần trăm 1..99 / NULL) validate quan hệ ở §1 #5; bắt luật chết âm thầm lúc tạo (drop_pct=89000 vô nghĩa). Có 3 test validate.
- threshold BIGINT intentional - đúng khớp yêu cầu rubric (NOT NUMERIC); ghi chú DEC-TRACK-22 trong DDL comment + §2.
- channel mặc định push (rẻ nhất) khi client không gửi - khớp cost model push>email>sms.
- Kiểm chủ sở hữu 404 (§1 #9, DEC-TRACK-24) chống IDOR, đồng nhất TRACK-002; TestCrossUser_404.
- §10 failure-modes 10 hàng không tầm thường (alias active/enabled, race, CASCADE).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 alert_rule schema | #1 | 0003_alert_rule.sql |
| #2 CHECK rule_type enum | #2 | CHECK + ValidateRule default branch |
| #3 channel TEXT[] in {push,email,sms} | #3 | validChannels + TestValidate_Channel |
| #4 threshold BIGINT theo loại | #4,#5,#6 | ValidateRule + 3 test |
| #5 validate quan hệ rule<->threshold | #4,#5,#6 | TestValidate_PriceBelow/DropPct/SignalRules |
| #6 bảng alert + CASCADE | #11 | 0003_alert_rule.sql |
| #7 create + FK product | #7 | TestCreate_UnknownProduct_400 |
| #8 list/patch/delete | #8 | alert_rule.go handlers |
| #9 kiểm chủ sở hữu 404 | #9 | TestCrossUser_404 |
| #11 partial index active=true | #10,#12 | idx_ar_eval + TestToggleActive_RemovesFromEvalIndex |
| #12 OTel | #12 | alert_rule_ops_total |

## §4 - Kết luận

Mọi mệnh đề normative có code/SQL/test backing; không mệnh đề mồ côi. Hợp đồng tên cột (active/channel/threshold) khớp đúng engine FR-TRACK-004 + batch FR-DEAL-006 đọc; alias enabled được ghi chú tránh lệch. threshold BIGINT intentional đúng đặc tả. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit FR-TRACK-003.*
