---
nfr_id: NFR-EXT-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-EXT-001 khóa cứng ràng buộc Manifest V3 (§3.2 nguồn) thành bất biến độ tin cậy cho toàn module EXT: service worker ephemeral, no global state (dùng chrome.storage), chrome.alarms periodInMinutes >=0.5 (>=30s) thay setInterval, tác vụ nặng đẩy backend, listener top-level đồng bộ, fetch timeout <30s, offscreen/WSS vòng đời ngắn. 7 mệnh đề §1 đều đo được (static/AST grep, kill-survive test, đọc periodInMinutes) và có verify. SLO định lượng rõ (0 mất-state, 0 setInterval-lập-lịch, fetch <30s, sự kiện <5 phút). related_tasks (TASK-EXT-001/002/004/005/006/007/008) đều resolve. Đạt 10/10.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-EXT-001 khớp tên file; category=reliability, priority=MUST, verification=T, phase=P1; slo định lượng đa-điều-kiện; source §3.2. Đạt.

Kiểm §1 vs §3.2 nguồn: SW ephemeral (Chrome kill sau ~30s idle / >5 phút sự kiện / fetch treo >30s), no global state -> chrome.storage, chrome.alarms >=30s không setInterval, tác vụ nặng đẩy backend. Bốn ràng buộc nguồn đều thành clause normative. Khớp.

Kiểm §1 chi tiết: #2 tách `.local`/`.session` đúng ngữ nghĩa MV3. #3 trần periodInMinutes >=0.5 là ràng buộc nền tảng Chrome 120+ (đúng). #5 listener đăng ký top-level synchronous (đúng yêu cầu MV3 wake). #6 AbortController <30s. #7 offscreen/WSS không mở thường trực. Mỗi clause đo được.

Kiểm §3 đo lường: static/AST khẳng định không global state trong `src/background/**`, không setInterval/setTimeout-lập-lịch, listener top-level; kill-survive test (resetModules -> đọc lại storage, đếm mất-state mục tiêu 0); lint rule; đọc `periodInMinutes >= 0.5`. Cụ thể.

Kiểm §4 verification: static test (cổng CI bắt buộc), kill-survive test với mỗi state bền (sync queue/consent/scaffold), timeout test (fetch treo bị abort), lifecycle test (offscreen đóng, WSS không top-level), cross-FR audit traceability. Mỗi mệnh đề có test.

Kiểm §5: global state -> chặn merge, setInterval-lập-lịch -> chặn merge, kill-survive đỏ sev-2, fetch không timeout sev-3, offscreen/WSS thường trực sev-3, listener async sev-2. Hợp lý.

Kiểm typo: prose ASCII thuần, tiếng Việt đủ dấu, không từ cấm; backtick code/API hợp lệ. Không sửa gì.

## §3 - Kết luận

NFR phủ đúng §3.2 và áp xuyên 7 TASK-EXT (scaffold -> content scripts 3 sàn -> offscreen -> đồng bộ -> consent). Mỗi clause đo được + verify được bằng static + kill-survive test; SLO định lượng; §5 phân sev rõ. Không tìm thấy defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-EXT-001.*
