---
fr_id: TASK-EXT-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-EXT-003 đặc tả pipeline tối thiểu hóa dữ liệu client - van an toàn cuối trước khi dữ liệu rời máy, neo vào §3.2 ("gửi backend dạng tối thiểu hóa: chỉ productId, giá, số lượng - KHÔNG gửi cookie"). Tôi kiểm độc lập: ba trụ (allowlist thay vì denylist; lọc lần hai độc lập defense-in-depth; fail-closed qua validator) đều normative trong §1 và có AC + test. Điểm mạnh đáng ghi: §1 #6 quét cả GIÁ TRỊ chuỗi tìm pattern credential/PII, không chỉ tên trường - bịt kẽ "nhồi cookie vào productId được phép". Kết luận: 10/10.

## §2 - Findings

Frontmatter hợp lệ (depends_on TASK-EXT-002, blocks TASK-EXT-005 đúng dòng dữ liệu; các khóa bắt buộc không rỗng). §1 có 12 mệnh đề; §4 có 11 AC; §5 có 3 test file (no-pii-leak/allowlist/minimize). Quét typography: glyph `->` duy nhất ở dòng 211 nằm trong code block test (miễn); prose ASCII thuần.

- ISS-001 (kiểm, không phải lỗi): OutboundPayload (platform/items{productId,price,qty}/vouchers{code,minSpend,discountText}) là kiểu message client, không phải bảng CSDL phát lệnh; cart_item thật thuộc TASK-CART-002. Tiền tệ ở đây là `price`/`minSpend` ràng số nguyên (Number.isInteger, >=0, < 1e12) - nhất quán tinh thần BIGINT VND của DEC-PRICE-05 trong DATA-MODEL.
- ISS-002 (kiểm, không phải lỗi): allowlist (pick-by-name) được AC #3 + #7 khẳng định "khóa lạ mới mặc nhiên biến mất không cần sửa code lọc" - đúng bản chất whitelist của DEC-EXT-14, khác denylist.
- ISS-003 (kiểm, không phải lỗi): fail-closed (minimize trả null khi lệch schema) + metric rejected - AC #4. Không có nhánh "gửi đại". Đúng DEC-EXT-17.
- Không phát hiện defect cần sửa trong lượt này.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact (§3/§5) |
|---|---|---|
| #1 chỉ tập tối thiểu OutboundPayload | #1 | schema.ts + introspection test |
| #2 allowlist không denylist | #3, #7 | allowlist.ts pick + allowlist.test |
| #3 lọc lần hai độc lập | #2 | minimize.ts + no-pii-leak.test |
| #4 không cookie/token/PII trong payload | #2 | no-pii-leak.test |
| #5 fail-closed qua validator | #4 | validatePayload + minimize.test |
| #6 redact quét giá trị, không chỉ tên | #5 | redact.ts + ID_RE + redact test |
| #7 local-first trên client | #10 happy path | minimize.ts client-side |
| #8 productId khớp mẫu ID hợp lệ | #5 | ID_RE validate |
| #9 price/qty số nguyên trong ngưỡng | #6 | validatePayload bounds |
| #10 metric đếm passed/dropped/rejected | #9 | metrics.passed/rejected |
| #11 mọi outbound qua minimize -> queue | #8 | grep no bypass path |
| #12 không gắn token (việc TASK-EXT-005) | (ràng buộc) | §6 + tách lớp đồng bộ |

## §4 - Kết luận

Van an toàn cuối được đặc tả với mặc-định-an-toàn (allowlist), defense in depth, fail-closed và quét giá trị - bốn lớp đúng cho dữ liệu cá nhân. Mọi mệnh đề có AC + test. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập TASK-EXT-003.*
