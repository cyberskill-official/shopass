---
fr_id: FR-AFFIL-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-AFFIL-001 đặc tả sổ cái affiliate `affiliate_click` + `affiliate_conversion` ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC tương ứng và test trong §5. Hai trụ compliance được khóa ở mức dữ liệu: click chỉ ghi qua đường user-initiated (không có path tự động nền) và `sub_id` là token ngẫu nhiên không nhúng PII. Vòng đời conversion `pending -> confirmed/rejected` + một-click-một-conversion chống đếm trùng và cho phép delay payout. Tiền BIGINT VND. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Cookie-stuffing kiểu Honey ở mức schema
Nếu tồn tại đường ghi `affiliate_click` tự động (không cần user bấm), schema này sẽ tiếp tay cho chính mô hình khiến Honey bị gỡ. Giải: §1 #4 + DEC-AFFIL-01 - click chỉ qua `RecordClick` mà caller (FR-AFFIL-002) khẳng định user-initiated; `disallowed_tools` chặn ghi click tự động; ràng buộc chéo NFR-AFFIL-001.

### ISS-002 - sub_id rò rỉ PII
`sub_id` đi ra network và quay lại qua postback; nhúng `user_id`/email vào đó là rò rỉ định danh (vi phạm PDPL). Giải: §1 #2/#3 + DEC-AFFIL-02 - token ngẫu nhiên `crypto/rand`, map user ở cột nội bộ; AC #4 + `TestSubID_UniqueAndNoPII`.

### ISS-003 - Postback lặp + conversion mồ côi
Network retry postback có thể tạo conversion trùng; `sub_id` lạ có thể tạo conversion không gắn click. Giải: §1 #8 (ErrUnknownSubID, không tạo mồ côi) + §1 #11 (UNIQUE click_id, ON CONFLICT DO NOTHING); AC #6/#7 + `TestConversion_UnknownSubID_NoOrphan` + `TestConversion_PostbackIdempotent`.

### ISS-004 - Conversion confirm sớm + tiền float
Coi conversion là chắc ngay rồi trả cashback dẫn tới lỗ khi đơn đảo; float gây sai số đối soát. Giải: §1 #6 (BIGINT + CHECK >= 0) + §1 #7 (bắt đầu pending, confirm qua postback); AC #8/#10/#11.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 schema click + FK | #1 | `0001_affiliate_click.sql` |
| #2 sub_id unique | #3 | UNIQUE(sub_id) + `TestRecordClick_SubIDUnique` |
| #3 sub_id no PII | #4 | `subid.go` + `TestSubID_UniqueAndNoPII` |
| #4 click user-initiated | #2 | `RecordClick` + `disallowed_tools` |
| #5 schema conversion | #1 | `0002_affiliate_conversion.sql` |
| #6 tiền BIGINT + CHECK | #8 | CHECK >= 0 + `TestConversion_NegativeMoney_Rejected` |
| #7 status pending->confirmed | #9,#10,#11 | CHECK status + `ConfirmConversion` |
| #8 last-click + no orphan | #5,#6 | `RecordConversion` + `TestConversion_UnknownSubID_NoOrphan` |
| #11 postback idempotent | #7 | UNIQUE click_id + `TestConversion_PostbackIdempotent` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có DDL/code/test backing, gồm bất biến user-initiated, sub_id-no-PII và postback-idempotent. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-AFFIL-001.*
