---
id: FR-EXT-006
title: "UI settings + consent PDPL lúc cài - disclosure dữ liệu rõ ràng, opt-in từng loại dữ liệu (im lặng != đồng thuận), gắn consent record FR-COMPLY-001"
module: EXT
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-EXT-001, FR-EXT-002, FR-EXT-003, FR-COMPLY-001, FR-COMPLY-003, FR-TRUST-001, FR-TRUST-002, NFR-EXT-001]
depends_on: [FR-EXT-001, FR-COMPLY-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.4 (trust: chính sách dữ liệu minh bạch không gửi cookie/mật khẩu; xử lý tối thiểu hóa local-first; disclosure rõ Chrome Web Store)"
  - "docs/... §5.5 (PDPL Luật 91/2025: đồng thuận tự nguyện/cụ thể/đơn mục đích/tái lập được; im lặng != đồng thuận; quyền chủ thể dữ liệu)"
source_decisions:
  - "DEC-EXT-29: hiển thị màn hình consent + disclosure lúc cài (onboarding); người dùng phải opt-in TƯỜNG MINH từng loại dữ liệu trước khi extension đọc/gửi bất cứ gì; im lặng/mặc định-bật KHÔNG được coi là đồng thuận (PDPL §5.5)"
  - "DEC-EXT-30: consent có hạt (granular per-purpose): đọc giỏ hàng, đọc voucher, đồng bộ backend là các mục opt-in riêng; người dùng bật/tắt độc lập"
  - "DEC-EXT-31: ghi consent record (phiên bản chính sách, thời điểm, mục đã đồng ý) gửi cho FR-COMPLY-001 - đồng thuận phải TÁI LẬP ĐƯỢC (reproducible: chứng minh được đã đồng ý gì, khi nào)"
  - "DEC-EXT-32: disclosure nêu rõ KHÔNG thu thập mật khẩu, token/cookie sàn KHÔNG rời máy, chỉ gửi productId/giá/số lượng (khớp FR-EXT-003) - minh bạch hậu-Honey §5.4"
  - "DEC-EXT-33: trước khi có consent cho một mục, code đường dữ liệu của mục đó MUST bị chặn (consent gate kiểm tra trước mọi đọc/gửi); rút consent có hiệu lực ngay (dừng đọc/gửi mục đó)"
  - "DEC-EXT-34: cung cấp lối quản lý consent + xóa dữ liệu (liên kết DSAR FR-COMPLY-003) trong settings; người dùng xem/đổi/thu hồi bất cứ lúc nào"

language: "TypeScript 5.x; Manifest V3; HTML/CSS UI (options page + onboarding); chrome.storage consent record"
service: shopass/extension/
new_files:
  - extension/src/ui/onboarding.html
  - extension/src/ui/onboarding.ts
  - extension/src/ui/settings.html
  - extension/src/ui/settings.ts
  - extension/src/consent/consent-store.ts
  - extension/src/consent/consent-gate.ts
  - extension/test/consent-gate.test.ts
  - extension/test/consent-record.test.ts
  - extension/test/consent-default-off.test.ts
modified_files:
  - extension/manifest.json                 # thêm options_page (settings) + onboarding khi onInstalled
  - extension/src/background/service-worker.ts   # onInstalled mở onboarding; consent-gate chặn đường dữ liệu
  - extension/src/shared/types.ts                # thêm ConsentRecord, ConsentPurpose
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - mặc định bật consent / coi im lặng là đồng thuận (vi phạm DEC-EXT-29 - trái PDPL §5.5)
  - đọc giỏ/voucher hoặc đồng bộ trước khi có opt-in tường minh của mục đó (vi phạm DEC-EXT-33 - consent gate)
  - gộp mọi mục vào một consent "tất cả hoặc không" thay vì granular per-purpose (vi phạm DEC-EXT-30)
  - bỏ ghi consent record tái lập được (vi phạm DEC-EXT-31 - không chứng minh được đã đồng ý gì)

effort_hours: 5
sub_tasks:
  - "1.0h: onboarding.html + onboarding.ts - màn hình consent lúc cài, disclosure rõ, opt-in từng mục (mặc định TẮT)"
  - "1.0h: settings.html + settings.ts - options page xem/đổi/thu hồi consent + lối xóa dữ liệu (DSAR FR-COMPLY-003)"
  - "1.0h: consent-store.ts - lưu ConsentRecord (version chính sách, ts, mục đã đồng ý) chrome.storage + gửi FR-COMPLY-001"
  - "1.0h: consent-gate.ts - hàm gate kiểm consent trước mọi đọc/gửi; rút consent có hiệu lực ngay"
  - "0.5h: nối service-worker onInstalled mở onboarding; gate chặn đường dữ liệu FR-EXT-002/003/005"
  - "0.5h: consent-default-off + consent-gate + consent-record tests"

risk_if_skipped: "Đây là cổng pháp lý cứng: PDPL (Luật 91/2025, hiệu lực 01/01/2026) yêu cầu đồng thuận tự nguyện, cụ thể, đơn mục đích, tái lập được, và im lặng KHÔNG phải đồng thuận (§5.5). Một extension đọc ngữ cảnh đăng nhập sàn mà không có consent tường minh granular là vi phạm trực diện - chế tài tới 5% doanh thu năm trước cho vi phạm xuyên biên giới + tới 3 tỷ VND cho vi phạm nghiêm trọng. Nếu mặc định bật hoặc gộp 'tất cả hoặc không', đồng thuận không hợp lệ. Nếu không có consent gate, code có thể đọc/gửi trước khi người dùng đồng ý - đúng kiểu vi phạm. Nếu không ghi consent record tái lập được, ta không chứng minh được đã đồng ý gì khi cơ quan kiểm tra. Về niềm tin (§5.4), disclosure rõ ('không thu thập mật khẩu, token không rời máy, chỉ gửi productId/giá/số lượng') là điểm bán sống còn hậu-Honey - ~45% người tiêu dùng VN lo lộ dữ liệu. UI consent vừa là tuân thủ vừa là tài sản niềm tin."
---

## §1 - Mô tả (BCP-14 normative)

UI consent + settings **MUST** thu đồng thuận PDPL tường minh, granular per-purpose lúc cài (im lặng không phải đồng thuận), chặn mọi đường dữ liệu cho tới khi có opt-in, và ghi consent record tái lập được. Hợp đồng:

1. Extension **MUST** hiển thị màn hình consent + disclosure lúc cài (onboarding, mở khi `chrome.runtime.onInstalled`) trước khi đọc/gửi bất cứ dữ liệu nào (DEC-EXT-29). Người dùng **MUST** opt-in tường minh; mặc định mọi mục là TẮT.
2. Đồng thuận **MUST** granular per-purpose (DEC-EXT-30): các mục riêng cho "đọc giỏ hàng", "đọc voucher", "đồng bộ backend". Người dùng **MUST** bật/tắt từng mục độc lập. **MUST NOT** gộp thành một consent "tất cả hoặc không".
3. Im lặng hoặc mặc định-bật **MUST NOT** được coi là đồng thuận (PDPL §5.5): không checkbox tick sẵn, không "tiếp tục nghĩa là đồng ý" cho việc xử lý dữ liệu. Đồng thuận chỉ hợp lệ khi người dùng chủ động bật.
4. Trước mọi đọc/gửi của một mục, `consent-gate.ts` **MUST** kiểm consent của mục đó; nếu chưa đồng ý, đường dữ liệu **MUST** bị chặn (DEC-EXT-33). Content script (FR-EXT-002) đọc giỏ chỉ chạy khi consent "đọc giỏ" bật; đồng bộ (FR-EXT-005) chỉ chạy khi consent "đồng bộ" bật.
5. `consent-store.ts` **MUST** ghi `ConsentRecord` tái lập được (DEC-EXT-31): phiên bản chính sách, thời điểm (epoch), danh sách mục đã đồng ý. Record **MUST** gửi cho FR-COMPLY-001 (khung consent trung tâm) để chứng minh đã đồng ý gì, khi nào.
6. Disclosure **MUST** nêu rõ, bằng ngôn ngữ người dùng hiểu (tiếng Việt): KHÔNG thu thập mật khẩu; token/cookie sàn KHÔNG rời máy client; chỉ gửi `productId`/giá/số lượng về backend (khớp FR-EXT-003) (DEC-EXT-32).
7. Rút consent **MUST** có hiệu lực ngay: khi người dùng tắt một mục trong settings, đường dữ liệu của mục đó **MUST** dừng đọc/gửi tức thì (DEC-EXT-33); và ghi consent record mới (đã rút).
8. `settings.ts` (options page) **MUST** cho người dùng xem trạng thái consent hiện tại, đổi từng mục, và truy lối xóa dữ liệu (liên kết DSAR FR-COMPLY-003) (DEC-EXT-34) bất cứ lúc nào.
9. Khi phiên bản chính sách đổi (disclosure cập nhật), extension **MUST** yêu cầu lại consent cho phần thay đổi - không tự gia hạn consent cũ cho chính sách mới.
10. Consent state **MUST** lưu bền trong `chrome.storage` (sống qua SW kill, NFR-EXT-001) và là nguồn sự thật mà consent-gate đọc - KHÔNG giữ trong biến global.
11. UI **MUST** không dùng pattern lừa (dark pattern): nút đồng ý và từ chối ngang nhau về độ nổi bật; không ẩn nút từ chối; không gây hiểu lầm rằng phải đồng ý mới dùng được phần không liên quan.
12. Trước khi có BẤT KỲ consent nào, extension **MUST** ở trạng thái trơ về dữ liệu: không content script đọc, không đồng bộ; chỉ hiển thị onboarding. Cài đặt extension không tự nó là đồng thuận xử lý dữ liệu.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao consent tường minh lúc cài, mặc định tắt (DEC-EXT-29)?** PDPL (§5.5) nói rõ: đồng thuận phải tự nguyện, cụ thể, và im lặng không phải đồng thuận. Một extension đọc ngữ cảnh đăng nhập sàn là xử lý dữ liệu nhạy cảm; bật mặc định hay coi việc cài là đồng ý đều khiến đồng thuận vô hiệu - và mời chế tài tới 5% doanh thu. Mặc định tắt + opt-in chủ động là tư thế đúng luật và đúng niềm tin.

**Vì sao granular per-purpose (DEC-EXT-30)?** PDPL đòi đồng thuận "đơn mục đích". Gộp "đọc giỏ + đọc voucher + đồng bộ" vào một nút là ép người dùng đồng ý trọn gói - không hợp lệ. Tách từng mục để người dùng đồng ý đúng cái họ muốn (ví dụ chỉ đọc giỏ tại máy, chưa cho đồng bộ) là bản chất của đồng thuận hợp lệ.

**Vì sao consent gate chặn trước đọc/gửi (DEC-EXT-33)?** Consent trên giấy mà code vẫn đọc trước khi đồng ý thì vô nghĩa. Gate biến đồng thuận thành ràng buộc thực thi: mỗi đường dữ liệu hỏi gate trước, gate đọc consent state, chưa đồng ý thì chặn. Đây là điểm nối giữa UI consent và hành vi thực của extension - không phải trang trí.

**Vì sao consent record tái lập được (DEC-EXT-31)?** PDPL đòi đồng thuận "tái lập được". Khi cơ quan kiểm tra hỏi "người dùng đã đồng ý gì, phiên bản chính sách nào, khi nào", ta phải trả lời bằng bằng chứng. ConsentRecord (version + ts + mục) gửi FR-COMPLY-001 là bằng chứng đó. Không ghi record là không chứng minh được tuân thủ.

**Vì sao disclosure nêu rõ ranh giới kỹ thuật (DEC-EXT-32)?** Hậu-Honey, người dùng VN nghi extension đọc cookie là scam (~45% lo lộ dữ liệu, §5.4). Nói thẳng "không thu mật khẩu, token không rời máy, chỉ gửi productId/giá/số lượng" - và để FR-EXT-003 + FR-TRUST-003 chứng minh - biến disclosure từ lời hứa thành cam kết kiểm chứng được. Đây vừa là tuân thủ vừa là điểm bán.

**Vì sao rút consent có hiệu lực ngay (§1 #7)?** Đồng thuận rút được là quyền chủ thể dữ liệu. Nếu tắt mục mà đường dữ liệu vẫn chạy thêm một lúc, đó là xử lý không có cơ sở pháp lý trong khoảng đó. Hiệu lực tức thì (gate đọc state mới ngay) giữ đúng nguyên tắc.

---

## §3 - Hợp đồng API / DDL

### types.ts (consent, typed)

```ts
// extension/src/shared/types.ts (thêm)
export type ConsentPurpose = "read_cart" | "read_voucher" | "sync_backend";

export interface ConsentRecord {
  policyVersion: string;          // ví dụ "2026-06-27"
  decidedAt: number;              // epoch ms - "khi nào"
  granted: ConsentPurpose[];      // các mục đã đồng ý - "đồng ý gì"
  // tái lập được: đủ để chứng minh đã đồng ý gì, phiên bản nào, khi nào
}
```

### consent-store.ts (ghi record bền + gửi FR-COMPLY-001)

```ts
// extension/src/consent/consent-store.ts
const KEY = "sandeal:consent";

export async function getConsent(): Promise<ConsentRecord> {
  const o = await chrome.storage.local.get(KEY);
  return (o[KEY] as ConsentRecord) ?? { policyVersion: POLICY_VERSION, decidedAt: 0, granted: [] };
  // mặc định granted: [] - KHÔNG mục nào bật (DEC-EXT-29)
}

export async function setConsent(granted: ConsentPurpose[]): Promise<void> {
  const rec: ConsentRecord = { policyVersion: POLICY_VERSION, decidedAt: Date.now(), granted };
  await chrome.storage.local.set({ [KEY]: rec });
  await reportConsentToCompliance(rec);   // FR-COMPLY-001: tái lập được
}
```

### consent-gate.ts (chặn trước mọi đọc/gửi)

```ts
// extension/src/consent/consent-gate.ts
export async function ensureConsent(p: ConsentPurpose): Promise<boolean> {
  const rec = await getConsent();
  return rec.granted.includes(p);   // chưa đồng ý → false → đường dữ liệu bị chặn
}

// Ví dụ dùng ở content-script orchestration / sync:
// if (!(await ensureConsent("read_cart"))) return; // KHÔNG đọc giỏ khi chưa opt-in
// if (!(await ensureConsent("sync_backend"))) return; // KHÔNG đồng bộ khi chưa opt-in
```

---

## §4 - Acceptance criteria

1. `onInstalled` mở onboarding; trước khi có consent, không content script đọc + không đồng bộ (test: gate trả false cho mọi mục khi `granted: []`).
2. Consent mặc định TẮT: `getConsent()` trên cài mới trả `granted: []` (không mục nào bật).
3. Granular: bật riêng "read_cart" không tự bật "sync_backend" (test từng mục độc lập).
4. `ensureConsent(p)` trả false khi mục p chưa đồng ý -> đường dữ liệu của p bị chặn (test gate chặn đọc giỏ/đồng bộ).
5. `setConsent` ghi `ConsentRecord` có `policyVersion` + `decidedAt` + `granted`, và gọi `reportConsentToCompliance` (FR-COMPLY-001).
6. Rút consent (tắt mục trong settings) -> `ensureConsent` trả false ngay; ghi record mới với `granted` đã bỏ mục đó.
7. Disclosure UI chứa các câu cốt lõi: không thu mật khẩu, token/cookie sàn không rời máy, chỉ gửi productId/giá/số lượng (test nội dung).
8. Không dark pattern: không checkbox tick sẵn (test onboarding mặc định mọi mục off); nút từ chối hiện diện.
9. Đổi `policyVersion` -> yêu cầu lại consent cho phần thay đổi (test record cũ version khác không tự áp cho version mới).
10. Consent state bền: ghi consent, mô phỏng SW kill, đọc lại từ `chrome.storage` còn nguyên.
11. `settings.ts` có lối xóa dữ liệu liên kết DSAR (FR-COMPLY-003) (test có entry point).
12. `npm test` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/consent-default-off.test.ts
import { getConsent } from "../src/consent/consent-store";

test("cài mới: mọi consent TẮT (im lặng != đồng thuận)", async () => {
  globalThis.chrome = fakeChromeStorage();           // chưa ghi gì
  const rec = await getConsent();
  expect(rec.granted).toEqual([]);                   // KHÔNG mục nào bật
});

test("onboarding không tick sẵn mục nào", async () => {
  const html = await readFile("src/ui/onboarding.html", "utf8");
  expect(html).not.toMatch(/<input[^>]*type=["']checkbox["'][^>]*checked/i); // không checked sẵn
});
```

```ts
// extension/test/consent-gate.test.ts
import { ensureConsent } from "../src/consent/consent-gate";
import { setConsent } from "../src/consent/consent-store";

test("chưa opt-in → gate chặn đọc giỏ + đồng bộ", async () => {
  globalThis.chrome = fakeChromeStorage();
  expect(await ensureConsent("read_cart")).toBe(false);
  expect(await ensureConsent("sync_backend")).toBe(false);
});

test("granular: bật read_cart không tự bật sync_backend", async () => {
  globalThis.chrome = fakeChromeStorage();
  await setConsent(["read_cart"]);
  expect(await ensureConsent("read_cart")).toBe(true);
  expect(await ensureConsent("sync_backend")).toBe(false);  // độc lập
});

test("rút consent có hiệu lực ngay", async () => {
  await setConsent(["read_cart"]);
  await setConsent([]);                               // tắt
  expect(await ensureConsent("read_cart")).toBe(false);
});
```

```ts
// extension/test/consent-record.test.ts
test("setConsent ghi record tái lập được + gửi compliance", async () => {
  const sent: ConsentRecord[] = [];
  mockReportConsent(r => sent.push(r));
  await setConsent(["read_cart", "read_voucher"]);
  expect(sent[0].policyVersion).toBeTruthy();
  expect(sent[0].decidedAt).toBeGreaterThan(0);
  expect(sent[0].granted).toEqual(["read_cart", "read_voucher"]);
});

test("disclosure nêu rõ ranh giới kỹ thuật", async () => {
  const html = (await readFile("src/ui/onboarding.html", "utf8")).toLowerCase();
  expect(html).toMatch(/mật khẩu/);                  // không thu mật khẩu
  expect(html).toMatch(/token|cookie/);              // token không rời máy
  expect(html).toMatch(/productid|số lượng|giá/);    // chỉ gửi tối thiểu
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `consent-store.ts` (ConsentRecord bền + gửi FR-COMPLY-001, mặc định granted rỗng) -> `consent-gate.ts` (ensureConsent chặn trước đọc/gửi) -> `onboarding.html` + `onboarding.ts` (consent lúc cài, mặc định tắt, disclosure rõ) -> `settings.html` + `settings.ts` (xem/đổi/thu hồi + lối DSAR) -> nối `service-worker.ts` (onInstalled mở onboarding; gate chặn FR-EXT-002/003/005) -> tests. Content script đọc giỏ (FR-EXT-002) và đồng bộ (FR-EXT-005) đều gọi `ensureConsent` trước. Consent state là nguồn sự thật bền trong chrome.storage; mọi đổi consent ghi record mới gửi compliance.

---

## §7 - Phụ thuộc

- **FR-EXT-001** - scaffold MV3 + chrome.storage + onInstalled + options page làm khung UI consent.
- **FR-COMPLY-001** - khung consent PDPL trung tâm nhận ConsentRecord; FR này là bề mặt thu consent của extension.
- **FR-EXT-002** - content script đọc giỏ chỉ chạy khi consent "read_cart"/"read_voucher" bật (gate).
- **FR-EXT-003 / FR-EXT-005** - tối thiểu hóa + đồng bộ chỉ chạy khi consent "sync_backend" bật.
- **FR-COMPLY-003** - DSAR (xem/xóa dữ liệu); settings cung cấp lối vào.
- **FR-TRUST-001 / FR-TRUST-002** - open-source + minh bạch local-first; disclosure ở đây khớp cam kết đó.
- **NFR-EXT-001** - consent state bền trong storage để sống qua SW kill.

---

## §8 - Payload ví dụ

### ConsentRecord lưu trong chrome.storage (tái lập được)

```json
{
  "sandeal:consent": {
    "policyVersion": "2026-06-27",
    "decidedAt": 1782000000000,
    "granted": ["read_cart", "read_voucher"]
  }
}
```

### Đoạn disclosure (onboarding, tiếng Việt)

```text
SănDeal đọc giỏ hàng và voucher trong tab đã đăng nhập của chính bạn để giúp tối ưu giá.
- KHÔNG thu thập mật khẩu của bạn.
- Token/cookie phiên của sàn KHÔNG rời khỏi máy bạn.
- Chỉ gửi về máy chủ: mã sản phẩm (productId), giá, số lượng.
Bạn chọn bật từng mục dưới đây. Không bật mục nào thì SănDeal không đọc/gửi gì.
[ ] Đọc giỏ hàng    [ ] Đọc voucher    [ ] Đồng bộ với máy chủ SănDeal
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đa ngôn ngữ disclosure cho SEA (ID/TH...) - thêm khi mở per-country (FR-COMPLY-006/007), slice 1 tiếng Việt.
- Consent cho mục đích phụ (ví dụ phân tích ẩn danh B2B) - chỉ thêm khi FR-B2B-001 tới, opt-in riêng.
- Nhắc lại consent định kỳ (re-consent theo thời gian) - chốt cùng FR-COMPLY-002 (DPIA cập nhật mỗi 6 tháng).

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Mặc định bật consent | consent-default-off test | đồng thuận vô hiệu, vi phạm PDPL | Mặc định granted rỗng (DEC-EXT-29) |
| Im lặng coi là đồng ý | onboarding không-checked test | vi phạm §5.5 | Opt-in chủ động, không tick sẵn (§1 #3) |
| Gộp "tất cả hoặc không" | granular test | đồng thuận không đơn mục đích | Per-purpose độc lập (DEC-EXT-30) |
| Đọc/gửi trước khi đồng ý | consent-gate test | xử lý không cơ sở pháp lý | Gate chặn trước mọi đọc/gửi (DEC-EXT-33) |
| Không ghi consent record | consent-record test | không chứng minh được tuân thủ | Record tái lập được -> FR-COMPLY-001 (DEC-EXT-31) |
| Rút consent không hiệu lực ngay | rút-consent test | xử lý sau khi rút | Gate đọc state mới tức thì (§1 #7) |
| Disclosure mơ hồ | nội dung test | nghi malware (§5.4) | Nêu rõ không mật khẩu/token/chỉ tối thiểu (DEC-EXT-32) |
| Dark pattern ẩn nút từ chối | review UI | đồng thuận ép buộc | Nút đồng ý/từ chối ngang nhau (§1 #11) |
| Consent trong biến global | persist test | mất khi SW kill | Lưu bền chrome.storage (§1 #10) |
| Gia hạn consent cũ cho policy mới | version test | đồng ý sai phạm vi | Yêu cầu lại consent khi đổi version (§1 #9) |

---

## §11 - Ghi chú

- UI consent vừa là cổng pháp lý PDPL (§5.5: tự nguyện/cụ thể/đơn mục đích/tái lập được; im lặng != đồng thuận) vừa là tài sản niềm tin hậu-Honey (§5.4).
- Consent gate biến đồng thuận thành ràng buộc thực thi: mỗi đường dữ liệu (FR-EXT-002/005) hỏi gate trước; chưa opt-in thì chặn - không phải trang trí.
- ConsentRecord tái lập được (version + ts + mục) là bằng chứng tuân thủ gửi FR-COMPLY-001 - trả lời được "đồng ý gì, khi nào".
- Disclosure nêu thẳng ranh giới kỹ thuật (không mật khẩu, token không rời máy, chỉ productId/giá/số lượng) và để FR-EXT-003 + FR-TRUST-003 chứng minh - lời hứa kiểm chứng được.
- Mặc định tắt + granular + rút-được-ngay là ba trụ giữ đồng thuận hợp lệ; mọi đổi consent ghi record mới.

---

*Hết FR-EXT-006. Status: ready_to_implement (mục tiêu audit 10/10).*
