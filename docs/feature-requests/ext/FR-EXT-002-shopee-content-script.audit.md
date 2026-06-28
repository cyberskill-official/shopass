---
fr_id: FR-EXT-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-EXT-002 đặc tả content script Shopee đọc giỏ/voucher theo nguyên tắc session piggyback của §3.2. Tôi đối chiếu lại từng mệnh đề: nguyên tắc lõi ("content script chạy trong tab đăng nhập, mượn cookie first-party, KHÔNG thu mật khẩu, token KHÔNG rời client, chỉ gửi productId/giá/qua tối thiểu hóa") được hiện thực đúng - `fetch` dùng `credentials:'include'` để trình duyệt tự gắn cookie mà code KHÔNG đọc giá trị cookie; cấm `document.cookie` được test grep khẳng định, không chỉ quy ước. Ưu tiên internal JSON `/api/v4/cart/get` rồi fallback DOM nhiều selector (A/B resilient) + health signal khớp §3.2 mục Shopee. Kết luận độc lập: 10/10.

## §2 - Findings

Frontmatter đầy đủ và hợp lệ (id khớp tên file, module EXT, priority MUST, status ready_to_implement; depends_on FR-EXT-001 đúng, blocks gồm FR-EXT-007/008 hợp lý; new_files/sub_tasks/risk_if_skipped không rỗng). §1 có 12 mệnh đề; §4 có 11 AC; §5 có 3 test file. Quét typography: mọi `->` chỉ trong code block ts; prose ASCII thuần, không em-dash/curly quote.

- ISS-001 (kiểm, không phải lỗi): ranh giới "mượn cookie nhưng không đọc cookie" - AC #2 (grep no document.cookie / Set-Cookie / Authorization / password) + AC #3 (introspection payload không khóa cookie/token/session/auth) phủ đúng §1 #3 và #4. Đủ mạnh.
- ISS-002 (kiểm, không phải lỗi): fetch timeout 25s < 30s (AC #4) tôn trọng ràng buộc MV3 "fetch treo >30s làm SW liên đới kill" của §3.2.
- ISS-003 (kiểm, không phải lỗi): chỉ-đọc, không mutate giỏ (AC #10 grep endpoint mutate) khớp §3.9 (b) "đọc giỏ client-side Low risk" và tránh hành vi tự động hóa High risk của §3.9 (c).
- Không có bảng CSDL phát lệnh ở FR này (cart_snapshot/cart_item thuộc FR-CART-002); CartItem/VoucherItem là kiểu message client, không cần đối chiếu DATA-MODEL. Không phát hiện defect cần sửa trong lượt này.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact (§3/§5) |
|---|---|---|
| #1 content_scripts shopee.vn, document_idle | #1 | manifest.json matches |
| #2 JSON ưu tiên, DOM fallback | #5, #6 | cart-reader.ts + api-client.ts |
| #3 không đọc/copy cookie/token | #2, #3 | shopee-no-secret-leak.test |
| #4 không thu mật khẩu | #2 | grep input[type=password] |
| #5 chỉ productId/price/qty + voucher | #3 | normalize.ts + introspection |
| #6 nhiều selector dự phòng | #7 | dom-selectors.ts + dom-fallback test |
| #7 health signal khi hỏng hẳn | #8 | reportHealth -> FR-SCRAPE-006 |
| #8 chỉ đọc, không sửa giỏ | #10 | grep no mutate endpoint |
| #9 CartReadMessage typed, không credential | #3 | types.ts + payload introspection |
| #10 fetch timeout < 30s | #4 | api-client.ts AbortController |
| #11 chưa đăng nhập xử lý lịch sự | #9 | cart-reader test rỗng lịch sự |
| #12 local-first, chỉ tối thiểu đi tiếp | #3, #8 | normalize.ts + FR-EXT-003 |

## §4 - Kết luận

Ranh giới niềm tin cao nhất của module (mượn phiên qua trình duyệt, không bao giờ đọc cookie) được test khẳng định chứ không chỉ tuyên bố. Mọi mệnh đề có AC + test. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-EXT-002.*
