---
nfr_id: NFR-COMPLY-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-COMPLY-001 đặt ngưỡng tuân thủ PDPL cụ thể đo được: coverage consent = 100% trên mọi bề mặt xử lý dữ liệu cá nhân; 0 DPIA quá hạn 60 ngày (cập nhật mỗi 6 tháng); 0 breach quá hạn 72 giờ; 0 cleartext/token-on-server; 0 rò cross-user trong DSAR. Mỗi mệnh đề §1 nối một TASK-COMPLY thực thi và một phép đo/verify (counter, gauge, audit gate, property test). Statement gắn rõ chế tài PDPL (tới 5% doanh thu năm trước / 10 lần lợi nhuận bất chính / 3 tỷ VND). Số khớp §5.5 nguồn. related_tasks (TASK-COMPLY-001..008) đều resolve. Đạt 10/10.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-COMPLY-001 khớp tên file; category=compliance, priority=MUST, verification=T, phase=P1; slo định lượng (coverage 100%, các đếm = 0); source §5.5. Đạt.

Kiểm số/luật nguồn §5.5: line 354 PDPL Luật 91/2025/QH15 hiệu lực 01/01/2026, Nghị định 356/2025/NĐ-CP thay 13/2023; DPIA nộp trong 60 ngày + cập nhật mỗi 6 tháng; breach 72 giờ; chế tài tới 5% doanh thu năm trước (xuyên biên giới) / 10 lần lợi nhuận bất chính (mua bán dữ liệu) / 3 tỷ VND (vi phạm nghiêm trọng). §1 #3/#5/#9 và §2 dùng đúng các con số/mốc này. Không lệch.

Kiểm §1: 9 clause BCP-14. #1 coverage consent 100% (cổng tiến phase). #2 consent tự nguyện/cụ thể/đơn-mục-đích/tái-lập; im lặng/checkbox tích sẵn không tính - mặc định chưa đồng thuận. #4 TIA cross-border trước khi bật. #6 no-cleartext (argon2id + Vault + 0 token phiên sàn). #8 per-country deny-by-default. Ngưỡng số cụ thể, không mơ hồ.

Kiểm §3 đo lường: counter consent granted/withdrawn + coverage, gauge dpia_overdue/review_due_soon, gauge breach_overdue + clock acknowledged_at -> notified_authority_at < 72h, audit gate no_cleartext_gate.sh, property test rò cross-user, gauge gating_denied. Cụ thể.

Kiểm §4 verification: coverage audit (mọi handler gọi consent.IsAllowed đúng purpose), DPIA deadline test (>60 ngày -> overdue; >6 tháng -> review_overdue), breach clock test (>72h -> breach_overdue), no-cleartext gate trên toàn repo, DSAR cross-user property test, reconciliation định kỳ. Mỗi clause khóa có test.

Kiểm §5: coverage < 100% sev-2 chặn tiến phase, DPIA quá hạn sev-2, breach quá hạn sev-1, cleartext/token sev-1, rò cross-user sev-1, luật nước SEA sai sev-2. Phân sev hợp lý.

Kiểm khớp DATA-MODEL: consent_policy/consent_record, dpia/tia, dsar_request, breach_incident (acknowledged_at = đồng hồ 72h), country_rule deny-default - khớp các clause §1. Đạt.

Kiểm typo: prose ASCII thuần, tiếng Việt đủ dấu, không từ cấm. Không sửa gì.

## §3 - Kết luận

Toàn bộ statement có phép đo và task backing; ngưỡng số cụ thể (100% / 0) thay cam kết chung; mức nghiêm ngặt neo vào chế tài thật §5.5. Xử lý vi phạm phân sev hợp lý. Không tìm thấy defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-COMPLY-001.*
