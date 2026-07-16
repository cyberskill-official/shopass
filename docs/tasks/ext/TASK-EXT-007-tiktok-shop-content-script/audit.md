---
fr_id: TASK-EXT-007
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-EXT-007 đặc tả content script TikTok Shop, bám §3.2 ("cart/checkout trong webview/SPA -> content script đọc DOM; cơ chế ký msToken/_signature/X-Bogus + app attestation mạnh -> ưu tiên đọc DOM render thay vì gọi API ký") và §3.9 (TikTok Shop High risk scraping). Tôi kiểm độc lập ba điểm đặc thù: cấm ngược-kỹ-nghệ chữ ký (grep test no msToken/_signature/X-Bogus), MutationObserver chờ SPA render xong rồi mới đọc, tái dùng khung normalize/health của TASK-EXT-002 (không phân nhánh). Cam kết niềm tin (không cookie/mật khẩu, chỉ đọc) giữ nguyên qua no-secret-leak test. Kết luận: 10/10.

## §2 - Findings

Frontmatter hợp lệ (depends_on [TASK-EXT-002] đúng - tái dùng khung; priority MUST, phase P2; các khóa bắt buộc không rỗng). §1 có 12 mệnh đề; §4 có 11 AC; §5 có 3 test file (no-api-sign/cart-reader/spa-observer). Quét typography: mọi `->` chỉ trong code block ts; prose ASCII thuần (kèm dấu tiếng Việt hợp lệ).

- ISS-001 (kiểm, không phải lỗi): §1 #3 cấm tự sinh/ký msToken/_signature/X-Bogus - khớp chính xác chỉ thị §3.2 "ưu tiên đọc DOM render thay vì gọi API ký". AC #2 grep test khẳng định cho 4 file content/tiktok. Đúng ranh giới giòn + ToS/pháp lý.
- ISS-002 (kiểm, không phải lỗi): khác biệt SPA so với Shopee (route đổi không tải lại trang) - §1 #4 + spa-observer (AC #6) + §1 #12 chờ render thay vì báo hỏng nhầm. Phủ đúng cái bẫy "đọc DOM trống tạm thời".
- ISS-003 (kiểm, không phải lỗi): tái dùng normalize/health TASK-EXT-002 (AC #8 grep import `../shared/`) giữ một điểm kiểm soát tối thiểu hóa duy nhất - đúng DEC-EXT-37, tránh nhân đôi bề mặt rò.
- CartItem/VoucherItem là kiểu message tái dùng, không phải bảng CSDL; không đối chiếu DATA-MODEL. Không phát hiện defect cần sửa.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact (§3/§5) |
|---|---|---|
| #1 content_scripts tiktok per-domain | #1 | manifest matches *.tiktok.com |
| #2 đọc DOM render | #4 | cart-reader.ts readCartFromDom |
| #3 không API ký msToken/_signature/X-Bogus | #2 | tiktok-no-api-sign test |
| #4 MutationObserver cho SPA | #6 | spa-observer.ts + test |
| #5 selector dự phòng | #5 | dom-selectors.ts + variant test |
| #6 health signal khi hỏng | #7 | reportHealth -> TASK-SCRAPE-006 |
| #7 tái dùng normalize/health TASK-EXT-002 | #8 | grep import shared |
| #8 cam kết niềm tin không cookie/mật khẩu | #3, #9 | no-secret-leak test |
| #9 chỉ đọc, không sửa giỏ | (§10 grep mutate) | cart-reader read-only |
| #10 CartReadMessage platform tiktok tối thiểu | #9 | introspection payload |
| #11 consent gate read_cart | #10 | ensureConsent trước đọc |
| #12 chưa đăng nhập/SPA chưa render lịch sự | #6 | observer + empty-state |

## §4 - Kết luận

Đường né API ký (§3.2/§3.9) + SPA observer + tái dùng khung TASK-EXT-002 đều có test; cam kết niềm tin khẳng định cho cả TikTok. Mọi mệnh đề có AC backing. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập TASK-EXT-007.*
