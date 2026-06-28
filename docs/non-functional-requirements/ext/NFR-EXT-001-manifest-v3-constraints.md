---
id: NFR-EXT-001
title: "Ràng buộc Manifest V3 - service worker ephemeral, chrome.alarms >=30s (không setInterval), no global state (dùng chrome.storage), tác vụ nặng đẩy backend"
module: EXT
category: reliability
priority: MUST
verification: T
phase: P1
slo: "0 lỗi 'mất state khi SW kill' trong test suite; 0 setInterval/setTimeout-lập-lịch trong service worker; mọi tác vụ trong SW kết thúc <5 phút và mọi fetch <30s; alarm polling >=30s"
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-EXT-001, FR-EXT-002, FR-EXT-004, FR-EXT-005, FR-EXT-006, FR-EXT-007, FR-EXT-008]
source: "docs/... §3.2 (Ràng buộc Manifest V3 & cách kiến trúc vòng quanh: SW ephemeral, chrome.alarms >=30s, no global state, tác vụ nặng đẩy backend)"
---

## §1 - Statement (BCP-14 normative)

1. Service worker của extension **MUST** được thiết kế như ephemeral: Chrome kill SW sau ~30 giây không hoạt động, khi một sự kiện chạy >5 phút, hoặc khi một `fetch()` không phản hồi >30 giây. Mọi module trong SW **MUST** chịu được việc bị kill bất cứ lúc nào mà không mất dữ liệu.
2. Mã trong service worker **MUST NOT** giữ state nghiệp vụ có ý nghĩa trong biến module-global; mọi state bền **MUST** lưu vào `chrome.storage` (`.local` cho state thường, `.session` cho state phiên-nhạy-cảm RAM-only). Biến trong bộ nhớ chỉ được dùng làm cache nhất thời, không phải nguồn sự thật.
3. Polling/lập lịch định kỳ trong service worker **MUST** dùng `chrome.alarms` với `periodInMinutes >= 0.5` (>=30 giây, trần Chrome 120+). Mã **MUST NOT** dùng `setInterval` hay `setTimeout` chu kỳ dài để lập lịch trong SW - chúng không sống qua chu kỳ kill.
4. Tác vụ nặng hoặc dài (đọc nhiều tab, đồng bộ khối lượng lớn, tính toán) **MUST** đẩy lên backend; service worker **MUST** giữ vai "đầu đọc nhẹ" và **MUST NOT** chạy vòng lặp làm một sự kiện vượt ngưỡng 5 phút.
5. Listener sự kiện (`onInstalled`, `onAlarm`, `onMessage`...) **MUST** đăng ký ở top-level đồng bộ khi SW khởi động - **MUST NOT** đăng ký bên trong callback bất đồng bộ (MV3 chỉ giao sự kiện wake cho listener đã đăng ký synchronously).
6. Mọi `fetch()` trong service worker **MUST** đặt timeout < 30 giây (AbortController) để một request treo không kéo SW bị kill liên đới và mất kết quả.
7. Tài liệu offscreen (FR-EXT-004) và WebSocket (FR-EXT-005) **MUST** có vòng đời ngắn (mở theo nhu cầu, đóng khi xong); **MUST NOT** mở thường trực để né cơ chế kill - điều đó phá chính ràng buộc ephemeral và lạm dụng tài nguyên.

## §2 - Vì sao ràng buộc này

Vòng đời service worker ephemeral là khác biệt nền tảng lớn nhất giữa Manifest V3 và V2, và là nguồn lỗi extension phổ biến nhất. Code chạy đúng lúc dev (SW còn sống) rồi hỏng trong thực tế khi SW bị kill: state trong biến global mất sạch, `setInterval` chết khi SW ngủ, listener đăng ký trễ bỏ lỡ sự kiện wake. Những lỗi này không tái hiện được trong dev và chỉ lộ ra ở máy người dùng - đắt để phát hiện và sửa. Ràng buộc này biến các quy ước MV3 (state ở chrome.storage, alarms thay setInterval, việc nặng off-device, listener top-level) thành bất biến kiểm chứng được bằng test, áp cho mọi FR-EXT (scaffold, content scripts 3 sàn, offscreen, đồng bộ, consent). Nó là điều kiện độ tin cậy: nếu vi phạm, toàn bộ vòng đọc giỏ - đồng bộ của extension hỏng âm thầm, kéo theo trải nghiệm và niềm tin (§5.4).

## §3 - Đo lường (measurement)

- Test suite (CI): grep/AST khẳng định không có biến module-global mang state trong `src/background/**`; không có `setInterval(`/`setTimeout(`-lập-lịch trong service worker; listener đăng ký top-level.
- Test "kill-survive": mô phỏng SW restart (reset module) sau khi ghi state, khẳng định state đọc lại từ `chrome.storage` còn nguyên - đếm số case mất state (mục tiêu 0).
- Lint rule: cấm `setInterval`/`setTimeout` chu kỳ dài trong thư mục service worker; cấm import API DOM-only vào SW.
- Metric runtime (tùy chọn, qua FR-EXT-005): đếm số fetch bị abort do timeout (>30s) và số sự kiện gần ngưỡng 5 phút - cảnh báo nếu khác 0.
- Kiểm alarm: mọi `chrome.alarms.create` có `periodInMinutes >= 0.5` (test đọc tham số).

## §4 - Verification

- Static test (T): grep/AST trên `src/background/**` + thư mục SW - không global state, không setInterval/setTimeout-lập-lịch, listener top-level, alarm >=30s. Đây là cổng CI bắt buộc cho mọi PR đụng extension.
- Kill-survive test (T): với mỗi state bền (hàng đợi đồng bộ FR-EXT-005, consent FR-EXT-006, state scaffold FR-EXT-001), `jest.resetModules()` rồi đọc lại từ storage - phải khớp giá trị đã ghi.
- Timeout test (T): fetch giả treo >30s -> bị AbortController hủy, không kéo SW; kết quả xử lý qua retry (FR-EXT-005).
- Lifecycle test (T): offscreen (FR-EXT-004) đóng sau khi xong; WSS (FR-EXT-005) không mở top-level thường trực.
- Cross-FR audit: mỗi FR-EXT có ít nhất một test bám một mệnh đề của NFR này (traceability trong audit từng FR).

## §5 - Xử lý khi vi phạm

- Phát hiện global state trong SW (static test đỏ) -> chặn merge; chuyển state sang `storage.ts` (FR-EXT-001) trước khi qua cổng.
- Phát hiện `setInterval`/`setTimeout`-lập-lịch trong SW -> chặn merge; thay bằng `chrome.alarms` >=30s; nếu cần nhịp <30s, xem lại thiết kế (trần là ràng buộc nền tảng, không vượt được).
- Kill-survive test đỏ (mất state) -> sev-2; truy module giữ state trong biến global; chuyển qua storage + thêm rehydrate khi wake.
- Fetch không timeout / sự kiện gần 5 phút -> sev-3; thêm AbortController <30s; tách việc nặng đẩy backend (DEC-EXT-05).
- Offscreen/WSS mở thường trực -> sev-3; sửa vòng đời (đóng khi xong); xác nhận không dùng để né kill.
- Listener đăng ký trong async callback (sự kiện wake thỉnh thoảng lỡ) -> sev-2; chuyển đăng ký lên top-level đồng bộ của entrypoint.

---

*Hết NFR-EXT-001.*
