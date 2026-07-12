---
id: FR-EXT-008
title: "Content script Lazada - Akamai-aware, đọc DOM giỏ đã render; KHÔNG né/giả Akamai sensor từ client; tái dùng khung reader/normalize/health của FR-EXT-002"
module: EXT
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-EXT-002, FR-EXT-003, FR-EXT-006, FR-SCRAPE-008, FR-SCRAPE-006, FR-TRUST-002, NFR-EXT-001]
depends_on: [FR-EXT-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (Lazada: đọc DOM giỏ hàng; Lazada (Alibaba) thường dùng Akamai -> ưu tiên đọc DOM đã render)"
  - "docs/... §5.4 (token không rời client), §9 (rủi ro ToS/anti-bot)"
source_decisions:
  - "DEC-EXT-40: Lazada (Alibaba) dùng Akamai (Bot Manager) -> content script đọc DOM giỏ ĐÃ RENDER trong tab đăng nhập; KHÔNG gọi internal API như Shopee (Akamai chặn/đòi sensor)"
  - "DEC-EXT-41: KHÔNG né/giả/sinh Akamai sensor data (_abck, bm_sz...) từ client - đó là chống bot-detection chủ động, giòn + rủi ro ToS/pháp lý; reader chỉ đọc DOM người dùng đã thấy"
  - "DEC-EXT-42: Akamai-aware nghĩa là KHÔNG kích hoạt heuristic bot của Akamai: đọc thụ động DOM đã render, KHÔNG bơm request bất thường, KHÔNG thao tác tốc độ máy; hành vi như người dùng đọc giỏ của chính họ"
  - "DEC-EXT-43: tái dùng khung reader + normalize + health signal của FR-EXT-002; chỉ thay lớp selector đặc thù Lazada (DOM riêng, đổi theo A/B)"
  - "DEC-EXT-44: giữ nguyên cam kết niềm tin FR-EXT-002: KHÔNG mật khẩu, token/cookie KHÔNG rời client; chỉ trích productId/giá/qty đã render; chỉ đọc, KHÔNG sửa giỏ/đặt hàng"

language: "TypeScript 5.x; Manifest V3 content script; tái dùng shared/normalize + shared/health"
service: shopass/extension/
new_files:
  - extension/src/content/lazada/cart-reader.ts
  - extension/src/content/lazada/dom-selectors.ts
  - extension/src/content/lazada/index.ts
  - extension/test/lazada-cart-reader.test.ts
  - extension/test/lazada-no-akamai-evasion.test.ts
  - extension/test/lazada-dom-fallback.test.ts
modified_files:
  - extension/manifest.json                 # thêm content_scripts cho lazada.vn + host_permissions
  - extension/src/shared/types.ts           # tái dùng CartItem/VoucherItem; platform "lazada"
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - né/giả/sinh Akamai sensor (_abck, bm_sz, sensor_data) hay gọi internal API qua tầng chống-bot (vi phạm DEC-EXT-41 - giòn + rủi ro ToS/pháp lý)
  - đọc/sao chép cookie/session token vào biến hay payload (vi phạm DEC-EXT-44 - phá cam kết niềm tin FR-EXT-002)
  - thu thập mật khẩu hoặc tự sửa giỏ/đặt hàng (vi phạm DEC-EXT-44 - chỉ đọc)
  - phân nhánh khung reader riêng thay vì tái dùng normalize/health của FR-EXT-002 (vi phạm DEC-EXT-43)

effort_hours: 8
sub_tasks:
  - "1.0h: manifest content_scripts matches https://www.lazada.vn/* + host_permissions; run_at document_idle"
  - "2.5h: dom-selectors.ts - bảng selector nhiều phương án cho item/giá/qty giỏ Lazada (A/B resilient)"
  - "2.0h: cart-reader.ts - đọc DOM giỏ đã render (KHÔNG API qua Akamai); orchestrate qua khung FR-EXT-002; health khi hỏng"
  - "0.5h: index.ts - entry content script Lazada; gửi CartReadMessage(platform:lazada) cho SW"
  - "1.0h: lazada-no-akamai-evasion.test.ts - KHÔNG _abck/bm_sz/sensor_data; KHÔNG document.cookie"
  - "1.0h: lazada-cart-reader.test.ts + lazada-dom-fallback.test.ts - parse DOM mẫu; selector chính hỏng -> fallback; parse hỏng -> health"

risk_if_skipped: "Lazada là sàn thứ ba hoàn thiện moat so sánh giá chéo 3 sàn - thiếu nó, lời hứa 'đa sàn thật' và bài toán so sánh chéo (điểm khác biệt cốt lõi vs BeeCost) không trọn. Lazada thuộc Alibaba và dùng Akamai Bot Manager - một hệ chống bot dựa sensor data (_abck, bm_sz) sinh từ JS đo hành vi. Cám dỗ kỹ thuật là né/giả sensor để gọi internal API; đây là cái bẫy giống TikTok: sensor đổi liên tục (giòn) và né bot-detection chủ động là ranh giới ToS/pháp lý nguy hiểm (§9, §5.5). Đường đúng theo tài liệu là đọc DOM đã render trong tab đăng nhập của chính người dùng - thụ động, không kích hoạt heuristic bot của Akamai. 'Akamai-aware' ở đây nghĩa là biết để TRÁNH gây nghi, không phải để vượt qua: đọc như người dùng đọc giỏ của họ. Nếu phân nhánh khung reader riêng thay vì tái dùng normalize/health của FR-EXT-002, ta nhân đôi bug và mất nhất quán tối thiểu hóa. FR này đóng bộ ba sàn cho moat mà vẫn giữ cam kết niềm tin và né rủi ro chống-bot."
---

## §1 - Mô tả (BCP-14 normative)

Content script Lazada **MUST** đọc giỏ hàng qua DOM đã render trong tab đăng nhập một cách thụ động (Akamai-aware), KHÔNG né/giả sensor Akamai từ client, và tái dùng khung reader/normalize/health của FR-EXT-002. Hợp đồng:

1. `manifest.json` **MUST** khai `content_scripts` với `matches` cho domain Lazada (ví dụ `https://www.lazada.vn/*`), `run_at: "document_idle"`, và thêm `host_permissions` tương ứng (per-domain, không `<all_urls>`, theo FR-EXT-001).
2. Content script **MUST** đọc dữ liệu giỏ bằng cách parse DOM đã render (DEC-EXT-40) - KHÔNG gọi internal API như Shopee, vì Lazada dùng Akamai chặn/đòi sensor.
3. Content script **MUST NOT** né, giả, hay sinh Akamai sensor data (`_abck`, `bm_sz`, `sensor_data`...) hay gọi internal API qua tầng chống-bot (DEC-EXT-41). Né bot-detection chủ động bị cấm: giòn (sensor đổi liên tục) + rủi ro ToS/pháp lý.
4. Hành vi đọc **MUST** thụ động và "như người dùng" (DEC-EXT-42): đọc DOM đã render, KHÔNG bơm request bất thường, KHÔNG thao tác tốc độ/đo hành vi máy để vượt Akamai. "Akamai-aware" là tránh gây nghi, không phải vượt qua.
5. `dom-selectors.ts` **MUST** cung cấp nhiều selector dự phòng cho mỗi trường (item/giá/qty) vì DOM Lazada đổi theo A/B test (DEC-EXT-43). Selector chính trượt -> thử dự phòng trước khi coi là hỏng.
6. Khi mọi selector hỏng, content script **MUST** phát health signal về service worker để FR-SCRAPE-006 ghi nhận - KHÔNG nuốt lỗi im lặng (tái dùng cơ chế FR-EXT-002).
7. Content script **MUST** tái dùng khung reader + `normalize.ts` + health signal của FR-EXT-002 (DEC-EXT-43); chỉ thay lớp selector đặc thù Lazada. **MUST NOT** phân nhánh một pipeline tối thiểu hóa riêng.
8. Content script **MUST** giữ nguyên cam kết niềm tin FR-EXT-002 (DEC-EXT-44): KHÔNG đọc `document.cookie`/token, KHÔNG thu mật khẩu, token/cookie KHÔNG rời client; chỉ trích `productId`/giá/qty đã render.
9. Content script **MUST** chỉ đọc: KHÔNG tự sửa giỏ, KHÔNG đặt hàng, KHÔNG áp voucher tự động (DEC-EXT-44).
10. Kết quả **MUST** gửi cho service worker dạng `CartReadMessage` typed với `platform: "lazada"`, chứa CHỈ `CartItem[]` + `VoucherItem[]` tối thiểu - đi tiếp qua pipeline FR-EXT-003.
11. Content script **MUST** chỉ chạy khi consent "read_cart" (FR-EXT-006) bật; gọi `ensureConsent` trước khi đọc.
12. Content script **MUST** xử lý chưa-đăng-nhập / giỏ rỗng / Akamai chặn trang (challenge) lịch sự: không lỗi, báo "chưa có dữ liệu giỏ", KHÔNG thử vượt challenge.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao đọc DOM, không gọi API qua Akamai (DEC-EXT-40/41)?** Lazada (Alibaba) dùng Akamai Bot Manager - sensor JS (`_abck`, `bm_sz`) đo hàng loạt tín hiệu để phân biệt người với bot. Né/giả sensor để gọi internal API là cái bẫy kép: sensor đổi liên tục (code vỡ mỗi lần Akamai cập nhật) VÀ né bot-detection chủ động là ranh giới ToS/pháp lý nguy hiểm (§9). Đọc DOM đã render trong tab đăng nhập là đường thụ động, bền, đúng tinh thần session piggyback.

**Vì sao "Akamai-aware" là tránh, không phải vượt (DEC-EXT-42)?** Hiểu lầm nguy hiểm là coi "Akamai-aware" như "biết cách qua mặt Akamai". Ngược lại: ta biết Akamai ở đó để KHÔNG kích hoạt heuristic bot của nó - đọc thụ động DOM người dùng đã thấy, không bơm request, không đo/giả hành vi máy. Hành vi của content script phải không phân biệt được với người dùng thật đọc giỏ của chính họ. Đây vừa là kỹ thuật vừa là ranh giới đạo đức.

**Vì sao tái dùng khung FR-EXT-002 (DEC-EXT-43)?** Giống TikTok (FR-EXT-007), phân nhánh pipeline riêng cho Lazada nhân đôi bug và mở bề mặt mới có thể lọt credential. Khung reader + normalize + health của FR-EXT-002 đã được test cho cam kết "không rò cookie"; tái dùng giữ một điểm kiểm soát tối thiểu hóa duy nhất. Chỉ lớp selector là per-sàn - Lazada có DOM riêng nhưng quy trình đọc/chuẩn hóa/báo hỏng là chung.

**Vì sao giữ nguyên cam kết niềm tin (DEC-EXT-44)?** Ranh giới "không cookie, không mật khẩu, chỉ đọc" là định vị hậu-Honey của toàn extension (§5.4), không phụ thuộc sàn. Lazada không phải ngoại lệ. Test no-secret-leak tương tự FR-EXT-002/007 khẳng định ranh giới cho cả Lazada.

**Vì sao không vượt Akamai challenge (§1 #12)?** Khi Akamai nghi và chặn trang bằng challenge, đúng đắn là dừng lại báo "chưa có dữ liệu giỏ" - tuyệt đối không thử giải/vượt challenge. Vượt challenge là né bot-detection chủ động (đúng điều cấm), và nếu Akamai chặn thì người dùng thật cũng bị chặn, nên không có gì để đọc một cách hợp lệ. Dừng lịch sự là tư thế đúng.

---

## §3 - Hợp đồng API / DDL

### manifest.json (bổ sung content script Lazada)

```jsonc
// extension/manifest.json (thêm)
{
  "host_permissions": ["https://shopee.vn/*", "https://*.tiktok.com/*", "https://www.lazada.vn/*"],
  "content_scripts": [
    {
      "matches": ["https://www.lazada.vn/*"],
      "js": ["content/lazada/index.js"],
      "run_at": "document_idle"
    }
  ]
}
```

### cart-reader.ts (đọc DOM thụ động, KHÔNG sensor Akamai; tái dùng normalize + health)

```ts
// extension/src/content/lazada/cart-reader.ts
import { normalizeCart } from "../shared/normalize";     // tái dùng FR-EXT-002
import { reportHealth } from "../shared/health";         // tái dùng FR-EXT-002

export async function readLazadaCart(): Promise<CartReadMessage> {
  if (!(await ensureConsent("read_cart"))) {             // FR-EXT-006 gate
    return { type: "CART_READ", platform: "lazada", items: [], vouchers: [] };
  }
  let raw = readCartFromDom();                            // CHỈ DOM render; KHÔNG _abck/bm_sz/sensor
  if (raw === null) {
    reportHealth({ platform: "lazada", broke: "cart", source: "dom" }); // FR-SCRAPE-006
    raw = [];
  }
  const items = normalizeCart(raw);                      // tối thiểu hóa chung
  return { type: "CART_READ", platform: "lazada", items, vouchers: readVouchersFromDom() };
  // LƯU Ý: KHÔNG document.cookie, KHÔNG né/giả sensor Akamai, KHÔNG gọi internal API qua chống-bot.
}
```

### dom-selectors.ts (nhiều selector dự phòng, đặc thù Lazada)

```ts
// extension/src/content/lazada/dom-selectors.ts
export const CART_ITEM_SELECTORS = [
  ".cart-item",                 // selector chính
  "[data-tracking='cart-item']",// dự phòng A
  ".item-content"               // dự phòng B (biến thể A/B)
];
export function firstMatch(root: ParentNode, sels: string[]): Element[] {
  for (const s of sels) {
    const found = root.querySelectorAll(s);
    if (found.length) return [...found];
  }
  return [];                    // không khớp → caller phát health
}
```

---

## §4 - Acceptance criteria

1. `manifest.json` có `content_scripts` match `https://www.lazada.vn/*`, `run_at: document_idle`; `host_permissions` thêm domain Lazada, KHÔNG `<all_urls>`.
2. Grep `src/content/lazada/**`: KHÔNG có `_abck`, `bm_sz`, `sensor_data`, KHÔNG né/giả Akamai sensor, KHÔNG gọi internal API qua tầng chống-bot.
3. Grep `src/content/lazada/**`: KHÔNG có `document.cookie`, KHÔNG đọc `input[type=password]`.
4. Reader đọc giỏ từ DOM mẫu Lazada -> trả đúng item tối thiểu (productId/price/qty).
5. Selector chính trượt -> dùng selector dự phòng; vẫn parse được item từ DOM "biến thể A/B".
6. Mọi selector hỏng -> phát health signal (broke: "cart") VÀ trả `items: []`, không ném lỗi.
7. Reader tái dùng `normalize.ts` + health của FR-EXT-002 (grep import từ `content/shared/`), không phân nhánh riêng.
8. `CartReadMessage` có `platform: "lazada"`, chỉ `items` + `vouchers`; không cookie/token (introspection payload).
9. Akamai challenge / giỏ rỗng -> trả message rỗng lịch sự, không thử vượt challenge (test path).
10. Reader chỉ chạy khi consent "read_cart" bật (gọi `ensureConsent`); chưa bật -> trả rỗng.
11. `npm test` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/lazada-no-akamai-evasion.test.ts
test("KHÔNG né/giả Akamai sensor, KHÔNG đọc cookie/mật khẩu", async () => {
  for (const f of ["cart-reader", "dom-selectors", "index"]) {
    const src = await readFile(`src/content/lazada/${f}.ts`, "utf8");
    expect(src).not.toMatch(/_abck|bm_sz|sensor_data/i);     // không sensor Akamai
    expect(src).not.toMatch(/document\.cookie/);             // không cookie
    expect(src).not.toMatch(/input\[type=["']?password/i);   // không mật khẩu
  }
});

test("payload Lazada KHÔNG chứa cookie/token/sensor", async () => {
  setConsent(["read_cart"]);
  document.body.innerHTML = lazadaCartFixtureMain;
  const msg = await readLazadaCart();
  const flat = JSON.stringify(msg).toLowerCase();
  for (const b of ["cookie", "token", "_abck", "bm_sz", "sensor", "password"]) {
    expect(flat).not.toContain(b);
  }
});
```

```ts
// extension/test/lazada-cart-reader.test.ts
test("đọc giỏ từ DOM Lazada đã render", async () => {
  setConsent(["read_cart"]);
  document.body.innerHTML = lazadaCartFixtureMain;          // 2 item
  const msg = await readLazadaCart();
  expect(msg.platform).toBe("lazada");
  expect(msg.items.length).toBe(2);
});

test("reader tái dùng normalize/health FR-EXT-002", async () => {
  const src = await readFile("src/content/lazada/cart-reader.ts", "utf8");
  expect(src).toMatch(/from ["']\.\.\/shared\/normalize["']/);
  expect(src).toMatch(/from ["']\.\.\/shared\/health["']/);
});

test("Akamai challenge / giỏ rỗng → rỗng lịch sự, không vượt challenge", async () => {
  setConsent(["read_cart"]);
  document.body.innerHTML = "<div id='akamai-challenge'>...</div>";
  const msg = await readLazadaCart();
  expect(msg.items).toEqual([]);                            // không thử vượt
});
```

```ts
// extension/test/lazada-dom-fallback.test.ts
test("selector chính trượt → dùng selector dự phòng (A/B variant)", () => {
  document.body.innerHTML = lazadaCartFixtureVariant;       // class đổi
  const items = readCartFromDom();
  expect(items && items.length).toBeGreaterThan(0);
});

test("parse hỏng hẳn → health signal + items rỗng", async () => {
  setConsent(["read_cart"]);
  document.body.innerHTML = "<div>không khớp</div>";
  const sp = jest.fn(); setHealthReporter(sp);
  const msg = await readLazadaCart();
  expect(sp).toHaveBeenCalledWith(expect.objectContaining({ broke: "cart", platform: "lazada" }));
  expect(msg.items).toEqual([]);
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: cập nhật `manifest.json` (content_scripts lazada + host_permissions) -> `dom-selectors.ts` (bảng selector dự phòng Lazada) -> `cart-reader.ts` (đọc DOM thụ động, tái dùng normalize + health FR-EXT-002, gate consent) -> `index.ts` (entry, gửi CartReadMessage platform lazada) -> tests. KHÔNG chạm `_abck`/`bm_sz`/sensor, KHÔNG vượt Akamai challenge. Khung reader/normalize/health dùng chung với Shopee (FR-EXT-002) và TikTok (FR-EXT-007); chỉ lớp selector là đặc thù Lazada. Kết quả qua pipeline FR-EXT-003.

---

## §7 - Phụ thuộc

- **FR-EXT-002** - cung cấp khung reader + normalize + health signal tái dùng; ranh giới niềm tin chung.
- **FR-EXT-003** - pipeline tối thiểu hóa nhận `CartReadMessage(platform:lazada)`, lọc lần hai.
- **FR-EXT-006** - consent "read_cart" gate trước khi đọc.
- **FR-SCRAPE-006** - DOM-change monitoring nhận health signal khi parser Lazada hỏng.
- **FR-SCRAPE-008 (song song, P2)** - adapter Lazada backend (Akamai, residential bắt buộc); reader này là phía client.
- **FR-TRUST-002** - local-first + tối thiểu hóa; reader giữ cam kết.
- **NFR-EXT-001** - content script gửi message cho SW ephemeral; không giữ state global.

---

## §8 - Payload ví dụ

### Message content Lazada -> service worker (tối thiểu, KHÔNG credential/sensor)

```json
{
  "type": "CART_READ",
  "platform": "lazada",
  "items": [
    { "productId": "L-5520118", "price": 215000, "qty": 1 },
    { "productId": "L-4410092", "price": 38000, "qty": 2 }
  ],
  "vouchers": [
    { "code": "LAZFREESHIP", "minSpend": 0, "discountText": "Freeship toàn sàn" }
  ]
}
```

### Health signal khi parser Lazada hỏng (-> FR-SCRAPE-006)

```json
{ "type": "PARSER_HEALTH", "platform": "lazada", "broke": "cart", "source": "dom", "ts": 1790000000000 }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đọc trang sản phẩm Lazada (PDP) ngoài giỏ - slice sau; P2 slice 1 nhắm giỏ + voucher.
- Phân biệt LazMall vs marketplace seller trong dữ liệu giỏ - hoãn tới FR-CART-001 (voucher_catalog) nếu cần.
- Xử lý Lazada đa quốc gia SEA (lazada.co.th/.com.my) - thêm khi mở per-country (FR-COMPLY-006); FR này nhắm lazada.vn.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Né/giả Akamai sensor (_abck/bm_sz) | lazada-no-akamai-evasion test | giòn + rủi ro ToS/pháp lý §9 | Chỉ đọc DOM render thụ động (DEC-EXT-41) |
| Hiểu "Akamai-aware" = vượt Akamai | review + thụ động test | né bot-detection chủ động | Đọc như người dùng, không bơm request (DEC-EXT-42) |
| Đọc document.cookie/token | no-secret-leak test | rò credential §5.4 | Cấm về mã (DEC-EXT-44) |
| Thu thập mật khẩu | grep password | phá cam kết lõi | Không hook form login (DEC-EXT-44) |
| Selector chính trượt | health + variant test | mất dữ liệu giỏ | Selector dự phòng (DEC-EXT-43) |
| Parse hỏng nuốt im lặng | thiếu health signal | mất tính năng âm thầm | Bắt buộc health -> FR-SCRAPE-006 (§1 #6) |
| Phân nhánh pipeline riêng | grep import shared | nhân đôi bug + bề mặt rò | Tái dùng normalize/health FR-EXT-002 (DEC-EXT-43) |
| Thử vượt Akamai challenge | challenge test | né bot chủ động, rủi ro pháp lý | Dừng lịch sự, không vượt (§1 #12) |
| Tự sửa giỏ/đặt hàng | grep mutate | đặt nhầm + nghi lạm dụng | Chỉ đọc (DEC-EXT-44) |
| Đọc khi chưa consent | consent gate test | xử lý không cơ sở pháp lý | ensureConsent("read_cart") trước (§1 #11) |

---

## §11 - Ghi chú

- Lazada đóng bộ ba sàn cho moat so sánh giá chéo - điểm khác biệt cốt lõi vs BeeCost (§5.6).
- Lazada (Alibaba) dùng Akamai Bot Manager; ranh giới cứng: KHÔNG né/giả sensor `_abck`/`bm_sz` - giòn + rủi ro ToS/pháp lý (§9, §5.5).
- "Akamai-aware" nghĩa là biết để TRÁNH gây nghi (đọc thụ động như người dùng), KHÔNG phải để vượt qua - vừa là kỹ thuật vừa là ranh giới đạo đức.
- Tái dùng khung reader/normalize/health của FR-EXT-002 giữ một điểm kiểm soát tối thiểu hóa duy nhất; chỉ selector là per-sàn (giống TikTok FR-EXT-007).
- Cam kết niềm tin (không cookie/mật khẩu, chỉ đọc) không phụ thuộc sàn - test no-secret-leak khẳng định cho cả Lazada.

---

*Hết FR-EXT-008. Status: ready_to_implement (mục tiêu audit 10/10).*
