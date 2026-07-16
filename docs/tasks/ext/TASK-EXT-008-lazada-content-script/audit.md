---
fr_id: TASK-EXT-008
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-EXT-008 đặc tả content script Lazada, bám §3.2 ("Lazada (Alibaba) thường dùng Akamai -> ưu tiên đọc DOM đã render") và §3.9 (Lazada Medium-High risk, Akamai). Tôi kiểm độc lập điểm sắc nhất: task định nghĩa rõ "Akamai-aware" nghĩa là TRÁNH gây nghi (đọc thụ động như người dùng), KHÔNG phải vượt Akamai - cấm né/giả/sinh sensor `_abck`/`bm_sz`/`sensor_data` (grep test) và cấm thử vượt challenge. Tái dùng khung normalize/health của TASK-EXT-002; cam kết niềm tin (không cookie/mật khẩu, chỉ đọc) giữ nguyên qua no-secret-leak test (gồm cả sensor). Kết luận: 10/10.

## §2 - Findings

Frontmatter hợp lệ (depends_on [TASK-EXT-002] đúng; priority MUST, phase P2; các khóa bắt buộc không rỗng). §1 có 12 mệnh đề; §4 có 11 AC; §5 có 3 test file (no-akamai-evasion/cart-reader/dom-fallback). Quét typography: glyph `->` duy nhất ở dòng 147 trong code block ts (miễn); prose ASCII thuần (kèm dấu tiếng Việt hợp lệ).

- ISS-001 (kiểm, không phải lỗi): §1 #3 + #4 phân biệt rạch ròi "Akamai-aware = tránh, không phải vượt" - đúng chỉ thị §3.2 "ưu tiên đọc DOM đã render" và né rủi ro §9. AC #2 grep no `_abck`/`bm_sz`/`sensor_data` cho 3 file; AC #9 challenge -> rỗng lịch sự, không vượt. Phủ đầy đủ.
- ISS-002 (kiểm, không phải lỗi): tái dùng normalize/health TASK-EXT-002 (AC #7 grep import `../shared/`) - giữ một điểm kiểm soát tối thiểu hóa duy nhất, đúng DEC-EXT-43, giống TikTok TASK-EXT-007. Chỉ lớp selector là per-sàn.
- ISS-003 (kiểm, không phải lỗi): no-secret-leak test (AC #3/#8) khẳng định payload không cookie/token/sensor cho cả Lazada - cam kết niềm tin không phụ thuộc sàn. Đúng.
- CartItem/VoucherItem là kiểu message tái dùng, không phải bảng CSDL; không đối chiếu DATA-MODEL. Không phát hiện defect cần sửa.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact (§3/§5) |
|---|---|---|
| #1 content_scripts lazada per-domain | #1 | manifest matches www.lazada.vn |
| #2 đọc DOM render, không API qua Akamai | #4 | cart-reader.ts readCartFromDom |
| #3 không né/giả sensor Akamai | #2 | lazada-no-akamai-evasion test |
| #4 đọc thụ động "như người dùng" | #2, #9 | grep + challenge test |
| #5 selector dự phòng | #5 | dom-selectors.ts + variant test |
| #6 health signal khi hỏng | #6 | reportHealth -> TASK-SCRAPE-006 |
| #7 tái dùng normalize/health TASK-EXT-002 | #7 | grep import shared |
| #8 cam kết niềm tin không cookie/mật khẩu | #3, #8 | no-secret-leak test |
| #9 chỉ đọc, không sửa giỏ | (§10 grep mutate) | cart-reader read-only |
| #10 CartReadMessage platform lazada tối thiểu | #8 | introspection payload |
| #11 consent gate read_cart | #10 | ensureConsent trước đọc |
| #12 chưa đăng nhập/challenge lịch sự, không vượt | #9 | challenge path test |

## §4 - Kết luận

Đường né API qua Akamai (§3.2/§9) + đọc thụ động không vượt challenge + tái dùng khung TASK-EXT-002 đều có test. Bộ ba sàn (Shopee/TikTok/Lazada) hoàn thiện cho moat so sánh chéo. Mọi mệnh đề có AC backing. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập TASK-EXT-008.*
