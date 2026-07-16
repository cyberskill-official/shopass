---
id: NFR-COMPLY-001
title: "PDPL compliance - consent tự nguyện/cụ thể, DPIA nộp 60 ngày + cập nhật 6 tháng, breach 72 giờ, no-cleartext; chế tài tới 5% doanh thu / 10 lần lợi nhuận bất chính / 3 tỷ VND"
module: COMPLY
category: compliance
verification: T
priority: MUST
phase: P1
slo: "Coverage consent = 100% trên mọi bề mặt xử lý dữ liệu cá nhân; 0 DPIA quá hạn 60 ngày; 0 vi phạm quá hạn báo cáo 72 giờ; 0 cleartext credential và 0 token phiên sàn trên server"
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_tasks: [TASK-COMPLY-001, TASK-COMPLY-002, TASK-COMPLY-003, TASK-COMPLY-004, TASK-COMPLY-005, TASK-COMPLY-006, TASK-COMPLY-007, TASK-COMPLY-008]
source: "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.5 (PDPL Luật 91/2025/QH15 hiệu lực 01/01/2026, NĐ 356/2025; consent, DPIA, 72h breach, no-cleartext; chế tài tới 5% doanh thu năm trước / 10 lần lợi nhuận bất chính / 3 tỷ VND)"
---

## §1 - Statement (BCP-14 normative)

1. Mọi bề mặt xử lý dữ liệu cá nhân **MUST** có cơ sở pháp lý là một bản ghi consent hợp lệ (TASK-COMPLY-001): coverage consent = 100%; xử lý không có consent tương ứng là vi phạm cứng (cổng tiến phase §7 BACKLOG).
2. Đồng thuận **MUST** tự nguyện, cụ thể, đơn mục đích, tái lập được; im lặng/checkbox tích sẵn KHÔNG được coi là đồng thuận. Hệ thống mặc định trạng thái chưa đồng thuận.
3. Mỗi hoạt động xử lý **MUST** có DPIA nộp trong vòng 60 ngày kể từ khi bắt đầu xử lý và được cập nhật tối thiểu mỗi 6 tháng (TASK-COMPLY-002); số DPIA quá hạn nộp hoặc quá hạn rà soát phải bằng 0.
4. Hoạt động chuyển dữ liệu xuyên biên giới **MUST** có TIA gắn kèm (nước nhận + cơ chế bảo vệ) trước khi bật.
5. Vi phạm dữ liệu **MUST** được thông báo cơ quan trong vòng 72 giờ kể từ khi tổ chức nhận biết (TASK-COMPLY-004); vi phạm nghiêm trọng còn phải thông báo chủ thể dữ liệu. Số sự cố quá hạn 72 giờ chưa thông báo phải bằng 0.
6. Hệ thống **MUST** giữ no-cleartext: mật khẩu chỉ tồn tại dạng argon2id, secrets chỉ trong Vault, và KHÔNG có token phiên sàn nào trên server (TASK-COMPLY-005); số phát hiện cleartext/token-on-server qua audit gate phải bằng 0.
7. Quyền chủ thể dữ liệu (truy cập/sửa/xóa/di chuyển) **MUST** thực thi được qua quy trình DSAR (TASK-COMPLY-003) trong thời hạn phản hồi đã định; rò dữ liệu chéo user = 0.
8. Khung tuân thủ **MUST** điều chỉnh theo nước khi mở SEA (TASK-COMPLY-006/007) và theo luật TMĐT VN (TASK-COMPLY-008); nước chưa cấu hình giữ deny-by-default.
9. Mức tuân thủ này tương xứng với chế tài PDPL: tới 5% doanh thu năm trước cho vi phạm chuyển dữ liệu xuyên biên giới, tới 10 lần lợi nhuận bất chính cho mua bán dữ liệu trái phép, tới 3 tỷ VND cho vi phạm nghiêm trọng - rủi ro hạng High (§9).

## §2 - Vì sao ràng buộc này

PDPL (Luật 91/2025/QH15, hiệu lực 01/01/2026; Nghị định 356/2025/NĐ-CP thay Nghị định 13/2023) là điều kiện vận hành hợp pháp của SănDeal tại VN. Sản phẩm chạm dữ liệu cá nhân ở nhiều điểm: extension đọc giỏ hàng, đăng ký, theo dõi giá theo tài khoản, alert. Vi phạm không chỉ là phạt tiền (tới 5% doanh thu năm trước cho vi phạm xuyên biên giới) mà còn phá định vị niềm tin hậu-Honey vốn là lợi thế cạnh tranh cốt lõi (§5.4, §5.6) - ~45% người tiêu dùng VN lo ngại lừa đảo/lộ dữ liệu. Tuân thủ PDPL vừa là nghĩa vụ pháp lý vừa là tài sản thương hiệu. Đây là ràng buộc nền quyết định cả tính hợp pháp lẫn sự sống còn của mô hình.

## §3 - Đo lường (measurement)

- Counter `consent_granted_total{purpose}` / `consent_withdrawn_total{purpose}` + báo cáo coverage: tỷ lệ bề mặt xử lý có gọi `IsAllowed` (mục tiêu 100%).
- Gauge `dpia_overdue_total`, `dpia_review_due_soon_total` (TASK-COMPLY-002): mục tiêu overdue = 0.
- Gauge `breach_overdue_total` (TASK-COMPLY-004): mục tiêu = 0; thời gian từ `acknowledged_at` tới `notified_authority_at` < 72 giờ.
- Audit gate `no_cleartext_gate.sh` (TASK-COMPLY-005): số vi phạm theo rule = 0 ở mọi PR.
- Property test rò cross-user trong DSAR (TASK-COMPLY-003): mục tiêu 0 rò rỉ.
- Gauge `gating_denied_total{country,gate}` (TASK-COMPLY-006): theo dõi nước/tính năng bị chặn.

## §4 - Verification

- Coverage audit (T): liệt kê mọi handler xử lý dữ liệu cá nhân, khẳng định mỗi handler gọi `consent.IsAllowed` đúng purpose; thiếu là fail (cổng tiến phase).
- DPIA deadline test (T): seed hoạt động bắt đầu >60 ngày chưa nộp -> `Status=overdue`; >6 tháng chưa review -> `review_overdue` (TASK-COMPLY-002).
- Breach clock test (T): sự cố `acknowledged_at` >72h chưa thông báo -> `breach_overdue`; high/critical không `closed` được nếu chưa `notified_subjects` (TASK-COMPLY-004).
- No-cleartext gate (T): CI chạy `Scan` trên toàn repo; fixture token/cleartext bị bắt, mã sạch pass (TASK-COMPLY-005).
- DSAR cross-user (T): property test sinh nhiều user, xuất từng user, khẳng định bundle chỉ chứa dữ liệu đúng chủ (TASK-COMPLY-003).
- Reconciliation định kỳ: 0 DPIA overdue, 0 breach overdue, 0 audit-gate finding, coverage consent = 100%.

## §5 - Xử lý khi vi phạm

- Coverage consent < 100% (một bề mặt xử lý thiếu `IsAllowed`) -> sev-2; chặn tiến phase cho tới khi bịt; thêm gọi `IsAllowed` hoặc gỡ bề mặt.
- DPIA quá hạn 60 ngày hoặc quá hạn rà soát 6 tháng -> sev-2; nộp/rà soát ngay; kiểm cảnh báo due-soon vì sao không nổ.
- Vi phạm quá hạn 72 giờ chưa thông báo cơ quan -> sev-1; kích hoạt runbook ứng cứu, thông báo ngay, rà soát vì sao đồng hồ không báo động.
- Audit gate phát hiện cleartext/token-on-server -> sev-1; chặn merge; gỡ ngay; nếu đã lên production thì coi là sự cố bảo mật, kích hoạt TASK-COMPLY-004.
- Rò dữ liệu chéo user trong DSAR -> sev-1; rút tính năng export, vá truy vấn buộc `user_id`, coi là breach nếu đã lộ.
- Vi phạm luật nước SEA do adapter sai/thiếu -> sev-2; tắt tính năng cho nước đó (deny-by-default), sửa adapter TASK-COMPLY-007.

---

*Hết NFR-COMPLY-001.*
