---
fr_id: TASK-WEB-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file TASK-WEB-001 hiện tại. task đặc tả khung Next.js 14 App Router + auth JWT + shell dashboard. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Hai ranh giới bảo mật lõi - token không vào web storage (in-memory + httpOnly refresh) và web không tự xác thực mật khẩu - được test khẳng định chứ không chỉ quy ước. API client tập trung + middleware guard giữ luồng đăng nhập an toàn đầu-cuối, tiêu thụ BFF qua JWT TASK-AUTH-002. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- Access token in-memory, refresh token httpOnly Secure SameSite cookie (§1 #3, DEC-WEB-02) - KHÔNG localStorage; có test "KHÔNG ghi token vào web storage".
- Web không tự hash/đối chiếu mật khẩu (§1 #4, DEC-WEB-03) ủy quyền auth-svc; AC #4 grep không pwd_hash.
- apiFetch refresh-on-401 đúng MỘT lần rồi logout (§1 #6); có test đếm đúng 3 lần fetch.
- Middleware guard redirect 307 + không render dữ liệu khi thiếu phiên (§1 #7, #8); có test middleware-guard.
- Base URL từ env NEXT_PUBLIC_API_BASE_URL (§1 #5, #9) trỏ gateway - không hardcode.
- §10 failure-modes 9 hàng không tầm thường (CSRF cookie thiếu Secure, refresh lặp vô hạn).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 App Router route group | #2 | cây (auth)/(app) |
| #2 shell layout | #2 | app/(app)/layout.tsx + app-shell.tsx |
| #3 token in-memory + httpOnly | #3,#10 | setAccessToken + test storage |
| #4 không tự xác thực mật khẩu | #4 | lib/auth.ts + §8 login payload |
| #5 API client tập trung | #5 | lib/api.ts apiFetch |
| #6 refresh-on-401 một lần | #6 | apiFetch + test refresh 3 fetch |
| #7 middleware guard | #7,#8 | middleware.ts + test guard |
| #8 không render khi thiếu phiên | #8 | matcher + redirect |
| #9 base URL từ env | #9 | .env.example + BASE |
| #11 tsc/test | #1 | strict + npm test |
| #12 i18n vi-VN | - | locale nền |

## §4 - Kết luận

Mọi mệnh đề normative có mã/test backing; hai ranh giới bảo mật (storage token, xác thực mật khẩu) được kiểm chứng bằng test, không chỉ quy ước. Không mệnh đề mồ côi. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit TASK-WEB-001.*
