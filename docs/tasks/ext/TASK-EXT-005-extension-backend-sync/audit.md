---
fr_id: TASK-EXT-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-EXT-005 đặc tả cầu nối đồng bộ extension <-> backend, neo vào §3.2 ("WebSocket giữ service worker sống khi cần realtime nhưng không lạm dụng; tác vụ nặng đẩy backend"). Tôi kiểm độc lập ranh giới niềm tin cốt lõi: request lên backend đính JWT của SănDeal, TUYỆT ĐỐI không token/cookie sàn - được test grep + introspection payload khẳng định, không chỉ quy ước. Hàng đợi bền trong chrome.storage giải đúng SW ephemeral; fail-closed khi thiếu JWT; WSS mở theo nhu cầu rồi đóng; JWT ở storage.session RAM-only. Kết luận: 10/10.

## §2 - Findings

Frontmatter hợp lệ (depends_on [TASK-EXT-003, TASK-AUTH-002] đúng - cần OutboundPayload sạch và JWT trước; các khóa bắt buộc không rỗng). §1 có 12 mệnh đề chính (kèm #4b tách nhánh 401); §4 có 11 AC; §5 có 4 test file (no-platform-token/queue-persist/ws-lifecycle nằm trong cùng bộ). Quét typography: mọi `->` chỉ trong code block ts; prose ASCII thuần.

- ISS-001 (kiểm, không phải lỗi): §1 #1 đính JWT SănDeal, cấm token sàn - khớp DATA-MODEL platform_account ("cố ý KHÔNG lưu cookie/token/password, DEC-AUTH-12"). Ranh giới token-not-on-server nhất quán toàn hệ. AC #1 + sync-no-platform-token test (grep + introspection cookie/SPC_/x-bogus/mstoken).
- ISS-002 (kiểm, không phải lỗi): đánh số §1 dùng "#4" rồi "#4b" rồi "#5" - không liên tục thuần nhưng mỗi mệnh đề vẫn rõ ràng, testable, và có AC (#5 retry/ack, #6 cho 401->refresh). Không phải defect nội dung; không sửa để tránh gold-plate.
- ISS-003 (kiểm, không phải lỗi): hàng đợi bền chrome.storage (AC #4 mô phỏng SW kill) là hệ quả trực tiếp của NFR-EXT-001; SyncEnvelope body chỉ OutboundPayload + clientTs, danh tính ở header (AC #2). Đúng tách bạch.
- Không phát hiện defect cần sửa trong lượt này.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact (§3/§5) |
|---|---|---|
| #1 JWT SănDeal, không token sàn | #1 | auth-bridge.ts + no-platform-token test |
| #2 chỉ OutboundPayload vào queue | #3 | grep enqueue -> minimize |
| #3 hàng đợi bền storage, không global | #4 | queue.ts + queue-persist test |
| #4 retry/backoff, ack chỉ 2xx | #5 | sender.ts flushQueue |
| #4b 401 -> refresh rồi thử lại | #6 | sender 401 branch |
| #5 fail-closed khi thiếu JWT | #7 | NoAuthError + fail-closed test |
| #6 WSS theo nhu cầu, đóng khi xong | #8 | ws-client.ts + ws-lifecycle test |
| #7 JWT ở storage.session RAM-only | #9 | getJwt session + grep |
| #8 HTTPS/WSS no-cleartext | (scheme) | TLS bắt buộc |
| #9 đầu đọc nhẹ, flush theo alarm | (ràng buộc) | §6 + không vòng lặp >5 phút |
| #10 body sạch, token ở header | #2 | SyncEnvelope introspection |
| #11 metric gửi/retry/fail-closed/WSS | (đo) | metrics.* |
| #12 đăng xuất xóa JWT + đóng WSS | #10 | logout path |

## §4 - Kết luận

Ranh giới cứng nhất của module (JWT SănDeal lên backend, token sàn KHÔNG BAO GIỜ rời client) được khóa bằng grep + payload introspection; hàng đợi bền và fail-closed giải đúng SW ephemeral. Mọi mệnh đề có AC/đo backing. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập TASK-EXT-005.*
