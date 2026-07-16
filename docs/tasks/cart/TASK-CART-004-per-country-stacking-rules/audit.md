---
fr_id: TASK-CART-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-CART-004 đặc tả engine luật stacking per-country ở mức triển khai được. 12 mệnh đề §1 normative neo vào AC §4 và test §5. Hai chính sách đối lập được mã hóa đúng luật thị trường: VN cho stack (1 shop + 1 platform + freeship), MY/PH 2025 bỏ stack (freeship gộp nhóm platform, loại trừ). Luật đọc CountryPolicy (TASK-INFRA-005) thay vì hardcode; mặc định no-stack an toàn cho nước chưa cấu hình. Hai con số vàng §3.5(3) (110k VN, 80k MY/PH) được test khẳng định qua optimizer thật. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Áp sai luật nước làm con số sai (đã chốt)
Rủi ro: áp luật VN cho MY/PH -> hiển thị 110k trong khi luật thực tế cho 80k. Giải: DEC-CART-20/21 + §1 #2/#3 hai policy riêng; AC #8 + test OptimizeExample_110k và _80k qua optimizer thật.

### ISS-002 - Hardcode if-country
Nhúng country vào logic làm mở nước phải sửa code. Giải: DEC-CART-19 + §1 #1 đọc CountryPolicy + factory map; AC #1 grep logic ValidStack không if-country.

### ISS-003 - Mặc định cho stack nước chưa kiểm
Cho stack nước chưa cấu hình hiển thị giảm cao hơn luật thực tế. Giải: DEC-CART-23 + §1 #7 mặc định no-stack; AC #7 + test DefaultsToNoStack.

### ISS-004 - Gộp nhóm freeship-platform
MY/PH gộp freeship nhóm platform phải loại trừ đúng. Giải: DEC-CART-22 + §1 #4 stack_group loại trừ + FreeshipGroupedWithPF; AC #3 + #5 + test RejectsPlatformPlusFreeship.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 đọc CountryPolicy | #1 | `policy_rules.go` |
| #2 VN stack | #2 | `vn_rules.go` |
| #3 MY/PH no-stack | #3,#4 | `mypph_rules.go` |
| #4 stack_group loại trừ | #5 | `hasSameStackGroupConflict` |
| #5 đọc cờ policy | #1 | `PolicyStackRules` |
| #6 factory theo country | #6 | `factory.go` |
| #7 mặc định no-stack | #7 | factory default |
| #8 số voucher cho phép | #2,#3 | ValidStack |
| #9 chỉ bool không tính tiền | #9 | grep no nhân/cộng |
| #10 khớp 110k/80k | #8 | optimize example tests |
| #11 tất định | #10 | determinism |

## §4 - Kết luận

Mọi mệnh đề normative có code/test backing; hai chính sách per-country đúng luật thị trường, đọc CountryPolicy thay hardcode, mặc định no-stack an toàn. Hai con số vàng §3.5(3) (110k/80k) kiểm chứng qua optimizer thật. Không mệnh đề "mồ côi". **Score = 10/10. Verdict: PASS.** Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-CART-004.*
