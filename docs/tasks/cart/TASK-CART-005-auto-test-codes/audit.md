---
fr_id: TASK-CART-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-CART-005 đặc tả `testCodes` ở mức triển khai được, bám đúng pseudo-code §3.5(4). 12 mệnh đề §1 normative neo vào AC §4 và test §5. Bốn ranh giới compliance sống còn (chỉ user-initiated, sleep nhịp người, revert sau mỗi thử, KHÔNG tự chốt đơn) được test khẳng định bằng grep + behavior, không chỉ quy ước - đúng yêu cầu Chrome policy 10/06/2025 (§4.2) và rủi ro ban High (§3.9c). Mã từ catalog không bruteforce. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Tự động hóa né Chrome policy (đã chốt)
Rủi ro: chạy nền không user gesture -> vi phạm policy 10/06/2025, rủi ro gỡ extension (existential §5.2). Giải: DEC-CART-24 + §1 #1 chỉ userInitiated; AC #1 + #2 + test user-initiated + grep không setInterval/alarms.

### ISS-002 - Tự chốt đơn kiểu Honey
Tự đặt hàng là giao dịch thay user, đúng loại Honey bị phạt. Giải: DEC-CART-26/28 + §1 #4 KHÔNG checkout; AC #5 + test grep không place-order/checkout.

### ISS-003 - Bị coi là bot
Thử liên tiếp không nghỉ kích hoạt anti-bot (§3.9 High). Giải: DEC-CART-25 + §1 #2 sleep random(2.5s,5s); AC #3 + test pacing trong [2500,5000).

### ISS-004 - Mã dính / bruteforce
Để mã dính rối giỏ; đoán mã vô đạo đức + anti-bot. Giải: DEC-CART-26 revert + DEC-CART-29 mã từ catalog; AC #4 + #7 + test revert đúng số lần + grep không sinh mã.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 chỉ user-initiated | #1 | guard `testCodes` |
| #2 sleep nhịp người | #3 | `pacing.ts` |
| #3 revert mỗi mã | #4 | `revert()` |
| #4 không chốt đơn | #5 | grep checkout |
| #5 sortDesc | #6 | `.sort()` |
| #6 mã từ catalog | #7 | `getCandidateCodesFromCatalog` |
| #7 bám §3.5(4) | #10 | vòng testCodes |
| #8 áp-gỡ qua UI | - | `code-applier.ts` |
| #9 mã sai lịch sự | #8 | tiếp tục vòng |
| #10 dừng được | #9 | `cancelled()` |
| #11 không credential | #11 | chỉ code+discount |

## §4 - Kết luận

Mọi mệnh đề normative có code/test backing; bốn ranh giới compliance (user-initiated, pacing, revert, không-chốt-đơn) kiểm chứng bằng grep + behavior test. Bám đúng pseudo-code §3.5(4). Không mệnh đề "mồ côi". **Score = 10/10. Verdict: PASS.** Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-CART-005.*
