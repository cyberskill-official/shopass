---
nfr_id: NFR-TRUST-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-TRUST-001 đặt ngưỡng bảo mật và niềm tin cụ thể đo được: 0 cleartext finding mọi tầng; 0 token/cookie phiên sàn rời client hoặc tồn tại ở backend; 100% mật khẩu argon2id; 100% secrets ứng dụng trong Vault/AWS Secrets Manager; bộ hook security audit (egress + SBOM + reproducible) tái chạy được PASS trên mọi bản ship. Mỗi mệnh đề §1 nối một task thực thi và một phép đo (counter/gauge/boolean). Số khớp §3.8 + §5.4 nguồn (no-cleartext, token không rời client, argon2id, Vault/AWS SM, ~45% người tiêu dùng VN lo lộ dữ liệu - Ken Research). related_tasks (TASK-EXT-003, TASK-COMPLY-005, TASK-TRUST-001/002/003, TASK-AUTH-001, TASK-INFRA-003) đều resolve. Đạt 10/10.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-TRUST-001 khớp tên file; category=security, priority=MUST, verification=T, phase=P1; slo định lượng (0 cleartext, 0 token rời client, 100% argon2id/Vault, audit hook PASS); source §3.8/§5.4. Đạt.

Kiểm số nguồn: §3.8 line 281 (KHÔNG cleartext credential; token phiên không rời client; secrets trong Vault HashiCorp/AWS Secrets Manager; argon2id). §5.4 line 351 (open-source extension, security audit độc lập, không gửi cookie/mật khẩu, tối thiểu hóa local-first, disclosure Chrome Web Store; ~45% người tiêu dùng VN lo lừa đảo/lộ dữ liệu - Ken Research). §1 và §2 dùng đúng. Không lệch.

Kiểm §1: 7 clause BCP-14. #1 no-cleartext (argon2id PHC + Vault) đếm = 0. #2 token phiên sàn KHÔNG rời client và KHÔNG ở backend; egress test ghi nhận 0 token rời máy. #4 tối thiểu hóa: chỉ tập `{platform, productId, price, qty}` (+ voucher hiển thị) rời máy. #5 ship cần (a) open-source + reproducible match, (b) disclosure khớp pipeline, (c) audit hook PASS. #7 device fingerprint hash một chiều, không PII sàn. Ngưỡng số cụ thể.

Kiểm §3 đo lường: counter cleartext_finding_total=0, egress_credential_leak_total=0, gauge password_hash_non_argon2id_total=0, secrets_outside_vault_total=0, boolean reproducible_build_match + security_audit_pass, outbound_to_unknown_host_total=0, fraud_signal_pii_finding_total=0. Cụ thể.

Kiểm §4 verification: kiểm tĩnh no_cleartext_gate.sh, egress test Playwright với phiên đăng nhập (negative control phải FAIL), reproducible verify so SHA-256, run-security-audit.sh bốn hook (egress + SBOM + reproducible + payload_guard), argon2id test, disclosure parity. Cưỡng chế ba lớp (tĩnh + động + reproducible) tái chạy được bởi bên thứ ba. Mỗi clause khóa có test.

Kiểm §5: egress leak/cleartext sev-1 chặn ship, token phiên ở backend sev-1 (xóa + xoay + PDPL High), reproducible/audit false sev-2, secrets ngoài Vault/non-argon2id sev-2, outbound host lạ sev-2, PII trong fraud signal sev-2. Phân sev hợp lý.

Kiểm typo: prose ASCII thuần, tiếng Việt đủ dấu, không từ cấm; backtick metric/tên hook hợp lệ. Không sửa gì.

## §3 - Kết luận

Toàn bộ statement có phép đo và task backing; ngưỡng số cụ thể (0 / 100%) thay cam kết chung; cưỡng chế ba lớp tái chạy được (tĩnh + động + reproducible), không dựa kỷ luật thủ công. Mức nghiêm ngặt neo vào rủi ro thật (~45% lo lộ dữ liệu, Honey, PDPL High). Không tìm thấy defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-TRUST-001.*
