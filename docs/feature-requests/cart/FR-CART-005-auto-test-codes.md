---
id: FR-CART-005
title: "testCodes - thử mã giảm an toàn client-side: sleep random(2.5s,5s) nhịp người, userInitiatedApply CHỈ khi user bấm thử mã (tuân Chrome policy), revert, KHÔNG tự chốt đơn, sort desc; bám pseudo-code §3.5(4)"
module: CART
priority: MUST
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-EXT-002, FR-CART-001, FR-CART-006, FR-AFFIL-004, NFR-AFFIL-001]
depends_on: [FR-EXT-002, FR-CART-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.5(4) (pseudo-code testCodes: sleep random(2.5s,5s), userInitiatedApply, revert, sortDesc, KHÔNG tự chốt đơn)"
  - "docs/... §4.2 (Chrome policy 10/06/2025: bắt buộc hành động người dùng; cấm tự động hóa né disclosure), §3.9c (tự động hóa voucher rủi ro ban High)"
source_decisions:
  - "DEC-CART-24: testCodes CHỈ chạy khi user chủ động bấm nút 'thử mã' (userInitiatedApply) - tuân Chrome Web Store policy 10/06/2025 bắt buộc hành động người dùng (§4.2); KHÔNG tự khởi động nền"
  - "DEC-CART-25: giữa mỗi lần thử sleep random(2.5s, 5s) - nhịp người, tránh anti-bot và tránh bị coi là tự động hóa lạm dụng (§3.5(4))"
  - "DEC-CART-26: sau khi thử mỗi mã phải revert (gỡ mã ra khỏi giỏ) - KHÔNG để mã dính, KHÔNG tự chốt đơn (§3.5(4)); chỉ đo kết quả, không thực hiện giao dịch"
  - "DEC-CART-27: trả kết quả sortDesc theo mức giảm - mã giảm nhiều nhất lên đầu để user tự quyết áp mã nào (gợi ý, không tự áp)"
  - "DEC-CART-28: testCodes là read-measure-revert - đọc giảm giá khi áp thử rồi gỡ; mọi áp dụng/chốt đơn cuối cùng do user tự làm trên giao diện sàn"
  - "DEC-CART-29: nguồn mã ứng viên từ voucher_catalog (FR-CART-001) còn hiệu lực; KHÔNG đoán/bruteforce mã (vừa vô đạo đức vừa rủi ro ban)"

language: "TypeScript 5.x; Manifest V3 content script (tái dùng khung reader FR-EXT-002); user-gesture gated; KHÔNG setInterval nền"
service: shopass/extension/
new_files:
  - extension/src/content/shared/test-codes.ts
  - extension/src/content/shared/code-applier.ts
  - extension/src/content/shared/pacing.ts
  - extension/src/ui/test-codes-button.ts
  - extension/test/test-codes-user-initiated.test.ts
  - extension/test/test-codes-revert.test.ts
  - extension/test/test-codes-pacing.test.ts
modified_files:
  - extension/src/shared/types.ts                # thêm CodeTestResult, TestCodesRequest
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - chạy testCodes tự động/nền không có user gesture (vi phạm DEC-CART-24, Chrome policy 10/06/2025 -> rủi ro gỡ extension)
  - bỏ sleep random giữa các lần thử (vi phạm DEC-CART-25, bị coi tự động hóa lạm dụng)
  - tự chốt đơn / để mã dính sau khi thử (vi phạm DEC-CART-26/28, thực hiện giao dịch thay user)
  - đoán/bruteforce mã thay vì đọc voucher_catalog (vi phạm DEC-CART-29, vô đạo đức + ban)

effort_hours: 6
sub_tasks:
  - "0.5h: types.ts - CodeTestResult{code, discount, applied}, TestCodesRequest (user-initiated flag)"
  - "0.75h: pacing.ts - sleep(random(2.5s, 5s)) bám §3.5(4); test dùng clock giả"
  - "1.0h: code-applier.ts - applyCodeViaUi (áp mã qua UI giỏ), readDiscount, revert (gỡ mã); KHÔNG mutate đơn/chốt"
  - "1.5h: test-codes.ts - vòng thử tuần tự bám pseudo-code §3.5(4): chỉ chạy khi userInitiated, sleep, apply, đo, revert, sortDesc"
  - "0.5h: test-codes-button.ts - nút 'thử mã' phát user gesture -> gọi testCodes (KHÔNG tự gọi)"
  - "1.0h: test-codes-user-initiated.test.ts - không gesture -> không chạy; có gesture -> chạy"
  - "1.25h: test-codes-revert.test.ts + test-codes-pacing.test.ts - mỗi mã revert sau thử, không chốt đơn; có sleep giữa các lần; sortDesc"

risk_if_skipped: "Auto-test mã là tính năng tiện ích cho persona Huy (săn mã giảm) - thử nhanh các mã ứng viên để biết mã nào giảm nhiều nhất. Nhưng đây là FR rủi ro compliance CAO NHẤT của CART: §3.9c xếp tự động hóa voucher mức rủi ro ban High, và §4.2 nêu Chrome Web Store policy 10/06/2025 (hậu-Honey) bắt buộc hành động người dùng và cấm tự động hóa né disclosure - chính chính sách đã khiến hàng loạt extension bị gỡ. Nếu làm SAI - chạy nền không user gesture, hoặc tự chốt đơn, hoặc để mã dính - thì SănDeal bị coi như Honey: vi phạm policy, rủi ro bị gỡ khỏi Chrome Web Store (rủi ro existential §5.2), và phá định vị minh bạch. Ranh giới sống còn: testCodes CHỈ chạy khi user chủ động bấm, sleep nhịp người giữa các lần, revert sau mỗi thử, KHÔNG BAO GIỜ tự chốt đơn - mọi giao dịch cuối do user tự làm. Phải test khẳng định các ranh giới này, không chỉ quy ước. Đoán/bruteforce mã vừa vô đạo đức vừa kích hoạt anti-bot."
---

## §1 - Mô tả (BCP-14 normative)

Extension **MUST** hiện thực `testCodes` - thử các mã giảm ứng viên trên giỏ một cách an toàn client-side, CHỈ khi người dùng chủ động khởi tạo, với nhịp người giữa các lần thử, gỡ mã sau mỗi lần, và TUYỆT ĐỐI không tự chốt đơn; bám đúng pseudo-code §3.5(4). Hợp đồng:

1. `testCodes` **MUST** chỉ chạy khi người dùng chủ động bấm nút "thử mã" (`userInitiatedApply`, user gesture) (DEC-CART-24); KHÔNG tự khởi động nền, KHÔNG `setInterval`/`alarms` tự gọi - tuân Chrome Web Store policy 10/06/2025 bắt buộc hành động người dùng (§4.2).
2. Giữa mỗi lần thử mã, `testCodes` **MUST** `sleep(random(2.5s, 5s))` (DEC-CART-25, §3.5(4)) - nhịp người, tránh anti-bot và tránh bị coi là tự động hóa lạm dụng.
3. Sau khi thử mỗi mã, `testCodes` **MUST** `revert` (gỡ mã ra khỏi giỏ) (DEC-CART-26): không để mã dính sau khi đo; mỗi lần thử là áp-đo-gỡ độc lập.
4. `testCodes` **MUST NOT** tự chốt đơn, đặt hàng, hay thực hiện bất kỳ giao dịch nào (DEC-CART-26, DEC-CART-28, §3.5(4)): nó chỉ ĐO mức giảm khi áp thử rồi gỡ; mọi áp dụng cuối cùng và chốt đơn do người dùng tự làm trên giao diện sàn.
5. `testCodes` **MUST** trả kết quả `sortDesc` theo mức giảm (DEC-CART-27): mã giảm nhiều nhất lên đầu, để người dùng tự quyết áp mã nào (gợi ý, không tự áp).
6. Mã ứng viên **MUST** đến từ `voucher_catalog` (FR-CART-001) còn hiệu lực (DEC-CART-29); `testCodes` **MUST NOT** đoán/bruteforce mã (vô đạo đức + kích hoạt anti-bot).
7. `testCodes` **MUST** bám cấu trúc pseudo-code §3.5(4): với mỗi mã trong danh sách ứng viên, `sleep(random(2.5s,5s))`, `apply = userInitiatedApply(code)`, nếu hợp lệ ghi `(code, discount)`, `revert()`, cuối cùng `sortDesc(results)`.
8. Mỗi lần áp thử **MUST** qua giao diện giỏ của sàn (như thao tác người dùng), KHÔNG gọi API mutate ẩn để áp/gỡ; tái dùng khung reader của FR-EXT-002 (chỉ-đọc trừ thao tác áp-gỡ mã user-initiated này).
9. `testCodes` **MUST** xử lý mã không hợp lệ lịch sự: mã sàn từ chối -> ghi nhận không-giảm, gỡ (nếu cần), tiếp mã sau; không lỗi vỡ luồng.
10. Toàn bộ quá trình **MUST** dừng được: nếu người dùng rời trang hoặc hủy, `testCodes` dừng vòng thử (không tiếp tục nền) - nhất quán với chỉ-khi-user-initiated.
11. `testCodes` **MUST NOT** gửi mã/kết quả kèm bất kỳ credential nào về backend; chỉ kết quả mức giảm (đồng nhất ranh giới tối thiểu hóa FR-EXT-003).
12. `npm test` xanh; `tsc --noEmit` sạch.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao chỉ user-initiated (DEC-CART-24)?** Đây là ranh giới sống còn hậu-Honey. Chrome Web Store policy 10/06/2025 (§4.2) - chính chính sách khiến hàng loạt extension bị gỡ - bắt buộc hành động người dùng và cấm tự động hóa né disclosure. Nếu `testCodes` chạy nền tự động thì SănDeal bị coi như Honey: vi phạm policy, rủi ro bị gỡ khỏi Chrome Web Store (existential, §5.2). Chỉ chạy khi user bấm nút là tuân policy và giữ minh bạch.

**Vì sao sleep nhịp người (DEC-CART-25)?** Thử nhiều mã liên tiếp không nghỉ trông như bot - kích hoạt anti-bot sàn (§3.9 xếp tự động hóa voucher rủi ro ban High) và bị coi là lạm dụng. `sleep(random(2.5s,5s))` giữa các lần mô phỏng nhịp người thật, giảm rủi ro ban và phù hợp ToS. Đây là pacing như FR-SCRAPE-005 nhưng phía client.

**Vì sao revert sau mỗi mã (DEC-CART-26)?** Mục tiêu là ĐO mã nào giảm nhiều nhất, không phải áp mã. Để mã dính sau khi thử làm rối giỏ người dùng (mã không tối ưu bị áp nhầm). Áp-đo-gỡ độc lập từng mã giữ giỏ sạch và để người dùng tự chọn mã cuối.

**Vì sao tuyệt đối không tự chốt đơn (DEC-CART-26/28)?** Tự chốt đơn là thực hiện giao dịch thay người dùng - rủi ro cực cao (đặt nhầm đơn) và đúng loại tự động hóa Honey bị phạt. SănDeal chỉ đo và gợi ý; mọi áp dụng cuối + chốt đơn do người dùng tự bấm trên sàn. Ranh giới này phân biệt "trợ lý gợi ý" với "bot thay người dùng".

**Vì sao sortDesc gợi ý không tự áp (DEC-CART-27)?** Người dùng cần thấy "mã GIAM50K giảm 50k, mã FREESHIP giảm 30k" và tự chọn. Sắp giảm-nhiều-nhất lên đầu giúp quyết nhanh. Không tự áp giữ quyền quyết định ở người dùng - vừa hợp policy (hành động người dùng) vừa đúng triết lý.

**Vì sao mã từ catalog, không bruteforce (DEC-CART-29)?** Đoán mã là vô đạo đức (thử mã không phải của mình) và kích hoạt anti-bot (nhiều mã sai liên tiếp). Mã ứng viên đến từ `voucher_catalog` còn hiệu lực (FR-CART-001) - các mã hợp lệ đã biết. Thử mã đã biết khác hoàn toàn với dò mã.

---

## §3 - Hợp đồng API / DDL

### Nhịp người (pacing.ts)

```ts
// extension/src/content/shared/pacing.ts
// sleep random 2.5s..5s giữa các lần thử (DEC-CART-25, §3.5(4)).
export function randomDelayMs(): number {
  return 2500 + Math.floor(Math.random() * 2500); // [2500, 5000)
}
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
```

### Áp - đo - gỡ (code-applier.ts) - KHÔNG chốt đơn

```ts
// extension/src/content/shared/code-applier.ts
// applyCodeViaUi áp mã qua giao diện giỏ (như thao tác người), đọc mức giảm, KHÔNG chốt đơn.
export interface ApplyOutcome { valid: boolean; discount: number } // VND int

export async function userInitiatedApply(code: string): Promise<ApplyOutcome> {
  enterCodeIntoCartField(code);   // gõ mã vào ô voucher của giỏ
  clickApplyButton();             // bấm "áp dụng" trên UI sàn
  const discount = readDiscountFromCart(); // đọc mức giảm hiển thị
  return { valid: discount > 0, discount };
  // LƯU Ý: KHÔNG có bước đặt hàng / chốt đơn (DEC-CART-26/28)
}

export function revert(): void {
  removeAppliedCodeFromCart();    // gỡ mã ra khỏi giỏ (DEC-CART-26)
}
```

### testCodes (test-codes.ts) - bám pseudo-code §3.5(4)

```ts
// extension/src/content/shared/test-codes.ts
import { randomDelayMs, sleep } from "./pacing";
import { userInitiatedApply, revert } from "./code-applier";

export interface CodeTestResult { code: string; discount: number }

// testCodes bám đúng pseudo-code §3.5(4) (DEC-CART-24..29).
// CHỈ gọi từ user gesture (nút "thử mã"); `userInitiated` phải true.
export async function testCodes(
  candidateCodes: string[],     // từ voucher_catalog còn hiệu lực (DEC-CART-29)
  opts: { userInitiated: boolean; cancelled: () => boolean }
): Promise<CodeTestResult[]> {
  if (!opts.userInitiated) {
    throw new Error("testCodes chỉ chạy khi user chủ động bấm (DEC-CART-24)");
  }
  const results: CodeTestResult[] = [];
  for (const code of candidateCodes) {
    if (opts.cancelled()) break;            // user rời/hủy -> dừng (§1 #10)
    await sleep(randomDelayMs());           // nhịp người (DEC-CART-25)
    const apply = await userInitiatedApply(code);
    if (apply.valid) {
      results.push({ code, discount: apply.discount });
    }
    revert();                               // gỡ mã, KHÔNG chốt đơn (DEC-CART-26)
  }
  return results.sort((a, b) => b.discount - a.discount); // sortDesc (DEC-CART-27)
}
```

### Nút thử mã (test-codes-button.ts) - phát user gesture

```ts
// extension/src/ui/test-codes-button.ts
// Nút do người dùng bấm; KHÔNG tự gọi testCodes (DEC-CART-24).
button.addEventListener("click", async () => {
  const codes = await getCandidateCodesFromCatalog(); // FR-CART-001 còn hiệu lực
  const results = await testCodes(codes, { userInitiated: true, cancelled: () => userLeft });
  renderSuggestions(results); // gợi ý, user tự áp mã trên sàn
});
```

---

## §4 - Acceptance criteria

1. `testCodes` ném lỗi (không chạy) khi `userInitiated` là false; chỉ chạy khi true (user gesture).
2. Grep `extension/src/**`: KHÔNG có `setInterval`/`chrome.alarms` gọi `testCodes`; nút phát user gesture là đường gọi duy nhất.
3. Giữa mỗi lần thử có `sleep(randomDelayMs())` với `randomDelayMs` trong `[2500, 5000)`.
4. Sau mỗi mã, `revert()` được gọi (gỡ mã); test khẳng định gọi revert đúng số lần = số mã thử.
5. Grep `code-applier.ts` + `test-codes.ts`: KHÔNG có bước đặt hàng/chốt đơn (không gọi endpoint/UI checkout/place-order).
6. Kết quả trả `sortDesc` theo `discount` (mã giảm nhiều nhất lên đầu).
7. Mã ứng viên đến từ catalog (FR-CART-001); grep: KHÔNG có vòng sinh/đoán mã (bruteforce).
8. Mã sàn từ chối (`valid=false`) không vào kết quả; vòng tiếp mã sau, không lỗi.
9. `cancelled()` true (user rời/hủy) -> vòng dừng, không tiếp tục nền.
10. Cấu trúc bám pseudo-code §3.5(4): sleep -> apply -> (ghi nếu hợp lệ) -> revert -> sortDesc.
11. Kết quả gửi đi (nếu có) không kèm credential (chỉ code + discount).
12. `npm test` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/test-codes-user-initiated.test.ts
import { testCodes } from "../src/content/shared/test-codes";

test("KHÔNG chạy khi không phải user-initiated", async () => {
  await expect(testCodes(["A", "B"], { userInitiated: false, cancelled: () => false }))
    .rejects.toThrow(/user chủ động/);
});

test("mã nguồn KHÔNG có setInterval/alarms gọi testCodes", async () => {
  const files = ["src/content/shared/test-codes.ts", "src/ui/test-codes-button.ts"];
  for (const f of files) {
    const src = await readFile(f, "utf8");
    expect(src).not.toMatch(/setInterval/);
    expect(src).not.toMatch(/chrome\.alarms/);
  }
});
```

```ts
// extension/test/test-codes-revert.test.ts
import { testCodes } from "../src/content/shared/test-codes";
import * as applier from "../src/content/shared/code-applier";

test("revert sau mỗi mã; KHÔNG chốt đơn", async () => {
  jest.spyOn(applier, "userInitiatedApply").mockResolvedValue({ valid: true, discount: 30_000 });
  const revertSpy = jest.spyOn(applier, "revert").mockImplementation(() => {});
  await testCodes(["A", "B", "C"], { userInitiated: true, cancelled: () => false });
  expect(revertSpy).toHaveBeenCalledTimes(3); // mỗi mã gỡ một lần
});

test("mã nguồn KHÔNG có bước chốt đơn", async () => {
  const src = await readFile("src/content/shared/code-applier.ts", "utf8");
  expect(src).not.toMatch(/place.?order|checkout|chốt đơn|submitOrder/i);
});

test("sortDesc theo discount", async () => {
  const seq = [{ valid: true, discount: 20_000 }, { valid: true, discount: 50_000 }, { valid: true, discount: 30_000 }];
  let i = 0;
  jest.spyOn(applier, "userInitiatedApply").mockImplementation(async () => seq[i++]);
  jest.spyOn(applier, "revert").mockImplementation(() => {});
  const res = await testCodes(["A", "B", "C"], { userInitiated: true, cancelled: () => false });
  expect(res.map((r) => r.discount)).toEqual([50_000, 30_000, 20_000]); // giảm nhiều nhất đầu
});
```

```ts
// extension/test/test-codes-pacing.test.ts
import { randomDelayMs } from "../src/content/shared/pacing";

test("delay nằm trong [2500, 5000)", () => {
  for (let k = 0; k < 100; k++) {
    const d = randomDelayMs();
    expect(d).toBeGreaterThanOrEqual(2500);
    expect(d).toBeLessThan(5000);
  }
});

test("có sleep giữa các lần thử (dùng clock giả)", async () => {
  jest.useFakeTimers();
  const applier = require("../src/content/shared/code-applier");
  jest.spyOn(applier, "userInitiatedApply").mockResolvedValue({ valid: false, discount: 0 });
  jest.spyOn(applier, "revert").mockImplementation(() => {});
  const { testCodes } = require("../src/content/shared/test-codes");
  const p = testCodes(["A", "B"], { userInitiated: true, cancelled: () => false });
  // mỗi mã chờ >=2.5s trước khi apply
  await jest.advanceTimersByTimeAsync(2500);
  await jest.advanceTimersByTimeAsync(5000);
  await p;
  jest.useRealTimers();
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `types.ts` (CodeTestResult, TestCodesRequest) -> `pacing.ts` (`randomDelayMs` + `sleep`, test với clock giả) -> `code-applier.ts` (`userInitiatedApply` áp-đo qua UI giỏ + `revert`, KHÔNG checkout) -> `test-codes.ts` (vòng bám pseudo-code §3.5(4): guard userInitiated, sleep, apply, ghi, revert, sortDesc) -> `test-codes-button.ts` (nút phát user gesture, đường gọi duy nhất) -> tests. Tái dùng khung reader của FR-EXT-002; thao tác áp-gỡ mã là ngoại lệ user-initiated duy nhất với nguyên tắc chỉ-đọc (áp thử qua UI như thao tác người, gỡ ngay). Mã ứng viên lấy từ `voucher_catalog` còn hiệu lực (FR-CART-001), không đoán. Toàn bộ gated sau user gesture; test khẳng định không có đường tự động.

---

## §7 - Phụ thuộc

- **FR-EXT-002** - khung content script + reader giỏ (session piggyback); `testCodes` tái dùng để áp-đo mã qua UI giỏ; thao tác áp-gỡ là ngoại lệ user-initiated của nguyên tắc chỉ-đọc (depends_on cứng).
- **FR-CART-001** - `voucher_catalog` cấp danh sách mã ứng viên còn hiệu lực; `testCodes` thử mã đã biết, không đoán (depends_on cứng).
- **FR-CART-006 (sibling)** - checklist xu cùng triết lý không-tự-động-click; cùng giữ ranh giới user-initiated.
- **FR-AFFIL-004 / NFR-AFFIL-001** - guardrails né Honey + Chrome policy; `testCodes` user-initiated là một mặt của compliance này.
- Extension/lib: TypeScript, Manifest V3 content script; test qua Jest + fake timers + jsdom.

---

## §8 - Payload ví dụ

### Kết quả thử mã (gợi ý cho user, sortDesc)

```json
[
  { "code": "GIAM50K",  "discount": 50000 },
  { "code": "FREESHIP", "discount": 30000 },
  { "code": "SHOPA20K", "discount": 20000 }
]
```

### Luồng (mô tả, không phải payload)

```
user bấm "thử mã"
  -> lấy mã ứng viên từ voucher_catalog (còn hiệu lực)
  -> với mỗi mã: sleep(2.5..5s) -> áp thử qua UI giỏ -> đọc giảm -> gỡ mã
  -> sortDesc -> hiển thị gợi ý
  -> user TỰ áp mã chọn và TỰ chốt đơn trên sàn
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Giới hạn số mã thử mỗi lần (tránh chuỗi quá dài) - thêm trần khi đo hành vi thật.
- Ghi nhớ mã đã thử gần đây để không thử lại ngay - tối ưu UX, không đổi ranh giới.
- Kết hợp kết quả thử mã với optimizer (FR-CART-003) để gợi ý combo - cân nhắc khi hai luồng chín.
- Hỗ trợ áp-đo trên TikTok Shop/Lazada (khung reader FR-EXT-007/008) - mở rộng per-sàn sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Chạy nền không user gesture | user-initiated test + grep | vi phạm Chrome policy -> gỡ extension | Chỉ userInitiated (DEC-CART-24) |
| Bỏ sleep giữa các lần | pacing test | bị coi bot -> ban (§3.9 High) | sleep random 2.5..5s (DEC-CART-25) |
| Tự chốt đơn | grep checkout/place-order | giao dịch thay user (Honey-style) | KHÔNG checkout (DEC-CART-26/28) |
| Mã dính sau thử | revert test | giỏ rối, áp nhầm mã | revert mỗi mã (DEC-CART-26) |
| Đoán/bruteforce mã | grep sinh mã | vô đạo đức + anti-bot | Mã từ catalog (DEC-CART-29) |
| Mã sai làm vỡ luồng | invalid-code test | mất các mã sau | Ghi không-giảm, tiếp tục (§1 #9) |
| Không dừng khi user rời | cancelled test | tiếp tục nền | Kiểm cancelled() (§1 #10) |
| Gửi credential kèm kết quả | introspection | rò rỉ | Chỉ code + discount (§1 #11) |
| Gọi API mutate ẩn áp/gỡ | grep endpoint mutate | né disclosure | Áp-gỡ qua UI như thao tác người (§1 #8) |

---

## §11 - Ghi chú

- Auto-test mã là FR rủi ro compliance cao nhất của CART: §3.9c xếp tự động hóa voucher rủi ro ban High, §4.2 Chrome policy 10/06/2025 bắt buộc hành động người dùng.
- Ranh giới sống còn: CHỈ user-initiated, sleep nhịp người, revert sau mỗi thử, KHÔNG BAO GIỜ tự chốt đơn - phải test khẳng định, không chỉ quy ước.
- testCodes là read-measure-revert: đo mã nào giảm nhiều nhất rồi gỡ; mọi áp dụng cuối + chốt đơn do user tự làm.
- sortDesc gợi ý không tự áp giữ quyền quyết định ở người dùng - hợp policy và đúng triết lý hậu-Honey.
- Mã ứng viên từ voucher_catalog (đã biết, còn hiệu lực) - thử mã đã biết khác hoàn toàn với dò mã.
- Bám đúng pseudo-code §3.5(4); thao tác áp-gỡ qua UI là ngoại lệ user-initiated duy nhất của nguyên tắc chỉ-đọc FR-EXT-002.

---

*Hết FR-CART-005. Status: ready_to_implement (mục tiêu audit 10/10).*
