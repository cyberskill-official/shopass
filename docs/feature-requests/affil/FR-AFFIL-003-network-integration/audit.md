---
fr_id: FR-AFFIL-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-AFFIL-003 đặc tả tích hợp network (Involve Asia/Accesstrade) + webhook postback ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test trong §5. Mắt xích đối soát tiền được khóa: postback xác thực HMAC trước khi ghi, secret tham chiếu Vault (không cleartext), map trạng thái network sang vòng đời conversion (chỉ approved -> confirmed), idempotent với retry, ghi raw mọi postback làm bằng chứng. Last-click thực thi qua sub_id. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Postback giả mạo
Webhook là endpoint công khai; ghi conversion từ payload chưa xác thực mở cửa giả mạo để rút cashback. Giải: §1 #5 + DEC-AFFIL-13 - VerifyPostback HMAC-SHA256 (`hmac.Equal`) trước khi ghi, 401 + không ghi nếu sai; AC #5 + `TestPostback_BadSignature_401_NoConversion`.

### ISS-002 - Secret cleartext
Lưu `postback_secret` trong DB nghĩa lộ DB = giả mạo toàn bộ postback + vi phạm PDPL. Giải: §1 #2/#11 + DEC-AFFIL-11 - cột `postback_secret_ref` trỏ Vault (FR-INFRA-003), đọc lúc verify; AC #4 (review không có cột cleartext).

### ISS-003 - Confirm sớm / đếm trùng
Confirm khi network còn pending là trả tiền đơn có thể đảo; retry tạo conversion trùng. Giải: §1 #8 (map approved->confirm) + §1 #9 (idempotent UNIQUE click_id); AC #6/#7/#10 + `TestPostback_RetryIdempotent`.

### ISS-004 - Mất bằng chứng + conversion mồ côi
Không ghi raw postback mất bằng chứng tranh chấp; sub_id lạ tạo conversion không gắn click. Giải: §1 #6 (LogPostback luôn) + §1 #7 (ErrUnknownSubID -> 404); AC #9/#11 + `TestPostback_UnknownSubID_404_NoOrphan`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 schema network | #1 | `0003_affiliate_network.sql` |
| #2 secret tham chiếu Vault | #4 | `postback_secret_ref` + review |
| #3 TemplateFor | #2,#3 | `network.go` + seed |
| #5 verify trước khi ghi | #5 | `VerifyPostback` + `TestPostback_BadSignature_401_NoConversion` |
| #6 log raw | #5,#11 | `0004_affiliate_postback_log.sql` + `LogPostback` |
| #7 last-click + no orphan | #9 | `RecordConversion` + `TestPostback_UnknownSubID_404_NoOrphan` |
| #8 map status network | #6,#7,#8 | `HandlePostback` switch + `TestPostback_Approved_Confirms` |
| #9 idempotent | #10 | UNIQUE click_id + `TestPostback_RetryIdempotent` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có DDL/code/test backing, gồm verify-trước-khi-ghi, secret-Vault, map-trạng-thái và idempotent-retry. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-AFFIL-003.*
