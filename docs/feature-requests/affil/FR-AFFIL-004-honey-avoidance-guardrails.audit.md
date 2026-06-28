---
fr_id: FR-AFFIL-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-AFFIL-004 đặc tả bộ guardrail né Honey ở mức triển khai được: biến lời hứa "không cookie-stuffing/dropping/pop-under/auto-redirect/forced-install" thành CI gate đỏ build. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test trong §5. Ba lớp phòng thủ: static scan pattern bị cấm trên host sàn, runtime assert user gesture (`isTrusted`), manifest least-privilege; cộng assertion đúng một cửa affiliate + disclosure ở affil-svc. Checklist ánh xạ từng mục policy Chrome (3/2025, thực thi 10/06/2025). Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Lời hứa không có cơ chế thực thi
"Không làm như Honey" trong tài liệu không chặn được một PR thêm `chrome.cookies.set`. Giải: §1 #1/#7 + DEC-AFFIL-19 - guardrail là CI gate bắt buộc, vi phạm đỏ build (exit khác 0); AC #2/#11 + `no-cookie-stuffing.test.ts`.

### ISS-002 - Vi phạm né static qua gián tiếp
Static lint có thể bị qua mặt. Giải: §1 #9 - kiểm cả static (pattern) lẫn runtime (`openAffiliate` assert `isTrusted`); AC #5/#6 + `single-affiliate-path.test.ts`.

### ISS-003 - Quyền hạn cho phép vi phạm
Cookie-stuffing/auto-redirect cần quyền chạm cookie/request. Giải: §1 #4 + DEC-AFFIL-18 - manifest audit chặn `cookies` permission + `webRequestBlocking`; AC #7/#8 + `manifest-audit.test.ts` (least-privilege).

### ISS-004 - Cửa affiliate thứ hai (back-door)
Một hàm thứ hai tự ghép affiliate mở lại lối Honey. Giải: §1 #5/#6 + DEC-AFFIL-20 - `AssertSingleAffiliatePath` khẳng định đúng một route + disclosure; AC #9/#10 + `TestAssert_BackDoor_Fails`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 no cookie-stuffing/dropping | #2,#3 | `no-cookie-stuffing.ts::scanSource` |
| #2 no pop-under/auto-redirect | #4 | scanSource window.open/tabs.update |
| #3 user gesture | #5,#6 | `user-gesture.ts` + assert isTrusted |
| #4 manifest least-privilege | #7,#8 | `manifest-audit.ts` |
| #5 đúng một cửa affiliate | #9,#10 | `AssertSingleAffiliatePath` + `TestAssert_BackDoor_Fails` |
| #6 disclosure bắt buộc | #10 | `IncludesDisclosure` + `TestAssert_NoDisclosure_Fails` |
| #7 CI gate đỏ build | #11 | exit-code test |
| #8 checklist policy | #12 | `CHROME-WEBSTORE-AFFILIATE-CHECKLIST.md` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/test backing, gồm ba lớp guardrail (static/runtime/manifest), assertion đúng-một-cửa + disclosure, và CI-gate-đỏ-build. Không có mệnh đề "mồ côi". Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-AFFIL-004.*
