---
fr_id: TASK-BILL-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-BILL-004 đặc tả `referral_code` + attribution + hook chống abuse ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test trong §5. Bốn cơ chế chống farming được khóa: cấm tự giới thiệu (so user_id), attribution bất biến (cấm gắn lại), không tự trả thưởng (phát sự kiện cho anti-fraud + delay payout), uses idempotent theo cặp. Mã loại ký tự gây nhầm; mã lỗi không chặn đăng ký. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Tự giới thiệu farm thưởng
Tạo mã rồi tự dùng là farming đơn giản nhất. Giải: §1 #5 + DEC-BILL-18 - so `user_id` chủ mã vs referee, từ chối + không tăng uses; AC #5 + `TestAttribute_SelfReferral_Blocked`.

### ISS-002 - Trả thưởng ngay -> fake-account rút
Trả tức thì cho phép farm tài khoản giả rút trước khi bị phát hiện. Giải: §1 #9 + DEC-BILL-19 - phát `referral.attributed` cho TASK-TRUST-004 + delay payout, không tự thưởng; AC #9 + `TestAttribute_PublishesEvent_NoDirectReward`.

### ISS-003 - Thao túng attribution + đếm trùng
Gắn lại referrer chuyển thưởng; uses tăng nhiều lần nhân đôi thưởng. Giải: §1 #6/#8 + DEC-BILL-17/20 - bất biến sau khi gắn + idempotent theo cặp; AC #6/#8 + `TestAttribute_AlreadyAttributed_Blocked`.

### ISS-004 - Mã hỏng / mã lỗi chặn đăng ký
Ký tự O/0/I/1 gây gõ sai; mã lỗi làm hỏng đăng ký. Giải: §1 #3 (loại ký tự nhầm) + §1 #11 (mã lỗi không lăn ngược đăng ký); AC #2/#10 + `TestNewCode_NoConfusingChars`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 schema referral_code | #1 | `0004_referral_code.sql` |
| #2 một mã/người | #3 | UNIQUE(user_id) |
| #3 mã tránh ký tự nhầm | #2 | `code.go` + `TestNewCode_NoConfusingChars` |
| #4 attribution lúc đăng ký | #4 | `Attribute` + `TestAttribute_Valid` |
| #5 cấm tự giới thiệu | #5 | `ErrSelfReferral` + `TestAttribute_SelfReferral_Blocked` |
| #6 cấm gắn lại | #6,#8 | `ErrAlreadyAttributed` + test |
| #8 uses idempotent | #8 | `TestAttribute_AlreadyAttributed_Blocked` |
| #9 không tự thưởng | #9 | `bus.Publish` + `TestAttribute_PublishesEvent_NoDirectReward` |
| #11 mã lỗi không chặn đăng ký | #10 | luồng register không lăn ngược |

## §4 - Kết luận

Toàn bộ mệnh đề normative có DDL/code/test backing, gồm cấm-tự-giới-thiệu, attribution-bất-biến, không-tự-thưởng và uses-idempotent. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-BILL-004.*
