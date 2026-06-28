---
id: FR-EXT-002
title: "Content script Shopee đọc giỏ hàng/voucher trong tab đã đăng nhập (session piggyback) - gọi /api/v4/cart/get cùng cookie first-party; KHÔNG thu thập mật khẩu; token KHÔNG rời client"
module: EXT
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-EXT-001, FR-EXT-003, FR-EXT-005, FR-SCRAPE-002, FR-SCRAPE-006, FR-TRUST-002, FR-TRUST-003, NFR-EXT-001]
depends_on: [FR-EXT-001]
blocks: [FR-CART-005, FR-CART-006, FR-EXT-003, FR-EXT-007, FR-EXT-008]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (Nguyên tắc cốt lõi session piggyback; mục Shopee)"
  - "docs/... §5.4 (trust: không gửi cookie/mật khẩu), §5.5 (PDPL: token phiên KHÔNG lưu/gửi server)"
source_decisions:
  - "DEC-EXT-07: content script chạy trong ngữ cảnh tab đã đăng nhập của chính người dùng; đọc DOM hoặc gọi internal endpoint /api/v4/cart/get cùng cookie phiên first-party của người dùng"
  - "DEC-EXT-08: TUYỆT ĐỐI không thu thập mật khẩu; token/cookie phiên KHÔNG rời khỏi máy client (cam kết niềm tin lõi hậu-Honey, §5.4/§5.5)"
  - "DEC-EXT-09: fetch internal endpoint dùng credentials:'include' để mượn cookie first-party của tab; KHÔNG đọc/sao chép giá trị cookie ra biến hay gửi đi"
  - "DEC-EXT-10: chỉ trích xuất dữ liệu giỏ hàng/voucher đã render (productId, giá, số lượng, mã voucher hiển thị); chuẩn hóa thành cấu trúc tối thiểu trước khi giao cho service worker (FR-EXT-003 lọc tiếp)"
  - "DEC-EXT-11: DOM giỏ hàng Shopee thay đổi theo A/B test -> parser dùng nhiều selector dự phòng + ưu tiên internal JSON khi truy cập được; phát tín hiệu cho FR-SCRAPE-006 khi parse hỏng"
  - "DEC-EXT-12: đọc là user-initiated hoặc theo nhịp nhẹ; KHÔNG tự động giao dịch, KHÔNG sửa giỏ; chỉ đọc"

language: "TypeScript 5.x; Manifest V3 content script; chrome.scripting/content_scripts; fetch credentials:'include'"
service: shopass/extension/
new_files:
  - extension/src/content/shopee/cart-reader.ts
  - extension/src/content/shopee/voucher-reader.ts
  - extension/src/content/shopee/dom-selectors.ts
  - extension/src/content/shopee/api-client.ts
  - extension/src/content/shopee/index.ts
  - extension/src/content/shared/normalize.ts
  - extension/test/shopee-cart-reader.test.ts
  - extension/test/shopee-no-secret-leak.test.ts
  - extension/test/shopee-dom-fallback.test.ts
modified_files:
  - extension/manifest.json                 # thêm content_scripts cho shopee.vn (matches)
  - extension/src/shared/types.ts           # thêm CartItem, VoucherItem, CartReadMessage
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - đọc/sao chép giá trị cookie hay session token vào biến hoặc payload (vi phạm DEC-EXT-08/09 - phá cam kết niềm tin lõi + PDPL)
  - thu thập trường mật khẩu hoặc bất kỳ input credential nào (vi phạm DEC-EXT-08)
  - tự động sửa giỏ / đặt hàng / áp voucher tự động (vi phạm DEC-EXT-12 - chỉ đọc)
  - gửi cookie/header nhạy cảm về backend (vi phạm DEC-EXT-08; FR-EXT-003 chỉ cho qua productId/price/qty)

effort_hours: 10
sub_tasks:
  - "1.0h: manifest content_scripts matches https://shopee.vn/* + run_at document_idle"
  - "1.5h: dom-selectors.ts - bảng selector nhiều phương án cho item/giá/qty/voucher (A/B resilient)"
  - "2.0h: cart-reader.ts - đọc DOM giỏ; ưu tiên api-client nếu internal JSON truy cập được"
  - "1.5h: api-client.ts - fetch('/api/v4/cart/get', {credentials:'include'}); KHÔNG chạm cookie; parse JSON ra item tối thiểu"
  - "1.0h: voucher-reader.ts - đọc voucher/freeship hiển thị (mã + điều kiện), không áp dụng"
  - "1.0h: normalize.ts - chuẩn hóa ra CartItem{productId, price, qty} tối thiểu; bỏ mọi trường thừa"
  - "1.5h: shopee-no-secret-leak.test.ts - khẳng định payload KHÔNG chứa cookie/token; không đọc document.cookie"
  - "1.5h: shopee-cart-reader.test.ts + shopee-dom-fallback.test.ts - parse DOM mẫu; selector chính hỏng -> fallback; parse hỏng hẳn -> phát tín hiệu health"

risk_if_skipped: "Đây là bề mặt khả dụng tối thiểu của extension - lý do người dùng cài SănDeal (đọc giỏ hàng để tối ưu voucher/giá). Đồng thời đây là FR rủi ro niềm tin cao nhất: extension đọc ngữ cảnh đăng nhập Shopee dễ bị nghi là scam (§5.4). Nếu lỡ đọc/gửi cookie hay token phiên, một sự cố là chiếm tài khoản sàn hàng loạt người dùng - thảm họa pháp lý PDPL (chế tài tới 5% doanh thu) và giết chết toàn bộ định vị 'không phải Honey'. Session piggyback đúng nghĩa là: mượn cookie first-party của tab qua credentials:'include' nhưng KHÔNG BAO GIỜ đọc giá trị cookie ra; chỉ lấy dữ liệu giỏ đã render. Ranh giới này phải được test khẳng định, không chỉ quy ước. DOM Shopee đổi theo A/B test nên parser không resilient sẽ vỡ thầm lặng và mất tính năng."
---

## §1 - Mô tả (BCP-14 normative)

Content script Shopee **MUST** chạy trong tab đã đăng nhập của chính người dùng, đọc dữ liệu giỏ hàng/voucher đã render theo nguyên tắc session piggyback, và TUYỆT ĐỐI không thu thập mật khẩu hay để token/cookie phiên rời máy client. Hợp đồng:

1. `manifest.json` **MUST** khai báo `content_scripts` với `matches: ["https://shopee.vn/*"]`, `run_at: "document_idle"` - chạy đúng trên domain Shopee đã khai trong `host_permissions` (FR-EXT-001 §1 #6).
2. Content script **MUST** đọc dữ liệu giỏ hàng theo một trong hai đường, ưu tiên (a): (a) gọi internal endpoint `/api/v4/cart/get` bằng `fetch(..., { credentials: "include" })` để mượn cookie phiên first-party của tab; (b) parse DOM giỏ hàng đã render khi endpoint không truy cập được (DEC-EXT-07).
3. Content script **MUST NOT** đọc, sao chép, hay ghi log giá trị cookie/session token: cấm `document.cookie`, cấm đọc header `Set-Cookie`/`Authorization`, cấm đưa bất kỳ giá trị credential nào vào biến hoặc payload (DEC-EXT-08, DEC-EXT-09). `credentials: "include"` chỉ để trình duyệt tự đính cookie vào request - extension không bao giờ thấy giá trị cookie.
4. Content script **MUST NOT** thu thập mật khẩu hay bất kỳ input credential nào (không đọc `input[type=password]`, không hook form đăng nhập) (DEC-EXT-08).
5. Dữ liệu trích xuất **MUST** giới hạn ở giỏ hàng/voucher đã hiển thị: với mỗi item lấy `productId`, `price`, `qty`; với voucher lấy mã + điều kiện hiển thị. Mọi trường khác **MUST** bị loại trước khi giao cho service worker (DEC-EXT-10). FR-EXT-003 lọc lần hai (defense in depth).
6. `dom-selectors.ts` **MUST** cung cấp nhiều selector dự phòng cho mỗi trường (item, giá, qty, voucher) vì DOM Shopee đổi theo A/B test (DEC-EXT-11). Khi selector chính trượt, parser **MUST** thử selector dự phòng trước khi coi là hỏng.
7. Khi cả internal JSON lẫn mọi selector DOM đều hỏng, content script **MUST** phát một tín hiệu sức khỏe (health signal) về service worker để FR-SCRAPE-006 (DOM-change monitoring) ghi nhận - KHÔNG nuốt lỗi im lặng (DEC-EXT-11).
8. Content script **MUST** chỉ đọc: KHÔNG tự sửa giỏ, KHÔNG đặt hàng, KHÔNG tự áp voucher (DEC-EXT-12). Áp voucher tự động (nếu có) thuộc FR-CART-005 với ràng buộc user-initiated riêng - không nằm ở đây.
9. Kết quả đọc **MUST** gửi cho service worker qua `chrome.runtime.sendMessage` dạng `CartReadMessage` typed (định nghĩa ở `types.ts`) chứa CHỈ `CartItem[]` + `VoucherItem[]` tối thiểu - KHÔNG cookie/header/token.
10. `api-client.ts` **MUST** đặt timeout cho `fetch` < 30 giây (ràng buộc MV3: fetch treo >30s làm SW liên đới bị kill; cũng tránh treo UI tab).
11. Content script **MUST** xử lý trường hợp người dùng CHƯA đăng nhập (endpoint trả lỗi auth / DOM giỏ rỗng) một cách lịch sự: không lỗi, báo trạng thái "chưa có dữ liệu giỏ", không thử đọc credential để "tự đăng nhập".
12. Toàn bộ đường dữ liệu **MUST** tuân local-first (FR-TRUST-002): trích xuất và chuẩn hóa diễn ra trên máy client; chỉ dữ liệu tối thiểu hóa đi tiếp (FR-EXT-003 quyết định gửi gì về backend).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao session piggyback, không lưu token (DEC-EXT-07/08)?** Định vị lõi của SănDeal hậu-Honey là "extension không phải malware": nó đọc giỏ của chính bạn, trong tab của chính bạn, mà không bao giờ chạm credential. `credentials: "include"` để trình duyệt tự gắn cookie first-party vào request `/api/v4/cart/get` - extension nhận được JSON giỏ hàng nhưng KHÔNG thấy giá trị cookie. Đây là khác biệt tinh tế nhưng sống còn: mượn phiên (qua trình duyệt) khác hoàn toàn với đánh cắp phiên (đọc cookie ra). Test phải khẳng định ranh giới này, vì một lần lỡ là thảm họa PDPL + niềm tin.

**Vì sao cấm document.cookie tuyệt đối (§1 #3)?** Chỉ cần một dòng `document.cookie` lọt vào là extension "có khả năng" lấy token - đủ để một audit độc lập (FR-TRUST-003) đánh trượt và đủ để người dùng nghi ngờ. Cấm về mã (grep test), không chỉ về quy ước, để cam kết là kiểm chứng được.

**Vì sao nhiều selector dự phòng (DEC-EXT-11)?** Shopee chạy A/B test liên tục; DOM giỏ hàng đổi class/cấu trúc mà không báo. Parser một-selector vỡ thầm lặng và người dùng mất tính năng mà ta không biết. Bảng selector dự phòng + ưu tiên internal JSON (ổn định hơn DOM) + phát health signal khi hỏng cho phép FR-SCRAPE-006 phát hiện và vá nhanh.

**Vì sao ưu tiên internal JSON hơn DOM (§1 #2)?** Endpoint `/api/v4/cart/get` trả cấu trúc ổn định hơn DOM render và rẻ hơn để parse. DOM là đường lui khi endpoint đổi/khóa. Hai đường cho cùng kết quả tối thiểu hóa.

**Vì sao chỉ đọc, không sửa giỏ (DEC-EXT-12)?** Tự sửa giỏ/đặt hàng là hành vi rủi ro cao (sai một bước là đặt nhầm đơn của người dùng) và dễ bị nghi tự động hóa lạm dụng - đúng loại hành vi Honey bị phạt. SănDeal chỉ đọc; mọi hành động (áp voucher) phải user-initiated và tách sang FR-CART-005 với ràng buộc riêng.

**Vì sao xử lý chưa-đăng-nhập lịch sự (§1 #11)?** Người dùng có thể bật extension khi chưa đăng nhập Shopee. Đúng đắn là báo "chưa có dữ liệu giỏ" - tuyệt đối không được thử đọc credential để "giúp đăng nhập". Hành vi này vừa hợp triết lý, vừa tránh lỗi.

---

## §3 - Hợp đồng API / DDL

### Kiểu dữ liệu (types.ts bổ sung)

```ts
// extension/src/shared/types.ts (thêm)
export interface CartItem {
  productId: string;   // chỉ ID hiển thị, KHÔNG kèm thông tin định danh người dùng
  price: number;       // VND, số nguyên
  qty: number;
}

export interface VoucherItem {
  code: string;        // mã voucher hiển thị
  minSpend?: number;   // điều kiện hiển thị (nếu có)
  discountText?: string;
}

export interface CartReadMessage {
  type: "CART_READ";
  platform: "shopee";
  items: CartItem[];
  vouchers: VoucherItem[];
  // KHÔNG có trường cookie/token/header
}
```

### api-client.ts (mượn cookie, KHÔNG đọc cookie)

```ts
// extension/src/content/shopee/api-client.ts
const CART_ENDPOINT = "https://shopee.vn/api/v4/cart/get";

export async function fetchCartViaApi(): Promise<CartItem[] | null> {
  const ctrl = new AbortController();
  const t = setTimeout(() => ctrl.abort(), 25_000); // <30s (MV3 ràng buộc)
  try {
    const res = await fetch(CART_ENDPOINT, {
      method: "GET",
      credentials: "include",   // trình duyệt tự gắn cookie first-party
      signal: ctrl.signal
    });
    if (!res.ok) return null;   // chưa đăng nhập / endpoint đổi → fallback DOM
    const json = await res.json();
    return mapCart(json);       // chỉ rút productId/price/qty
  } catch {
    return null;                // lỗi → fallback DOM, không nuốt im lặng ở caller
  } finally {
    clearTimeout(t);
  }
  // LƯU Ý: KHÔNG có document.cookie, KHÔNG đọc Set-Cookie/Authorization ở đây.
}
```

### cart-reader.ts (orchestrate: JSON trước, DOM sau, health signal)

```ts
// extension/src/content/shopee/cart-reader.ts
export async function readCart(): Promise<CartReadMessage> {
  let items = await fetchCartViaApi();          // (a) ưu tiên internal JSON
  let source: "api" | "dom" = "api";
  if (items === null) {
    items = readCartFromDom();                  // (b) fallback DOM nhiều selector
    source = "dom";
  }
  if (items === null) {
    reportHealth({ platform: "shopee", broke: "cart", source }); // FR-SCRAPE-006
    items = [];
  }
  const vouchers = readVouchersFromDom();
  return { type: "CART_READ", platform: "shopee", items, vouchers };
}
```

---

## §4 - Acceptance criteria

1. `manifest.json` có `content_scripts` match `https://shopee.vn/*`, `run_at: document_idle`.
2. Grep toàn bộ `src/content/shopee/**` + `normalize.ts`: KHÔNG có `document.cookie`, KHÔNG đọc header `Set-Cookie`/`Authorization`, KHÔNG đọc `input[type=password]`.
3. `CartReadMessage` chỉ chứa `items: CartItem[]` + `vouchers: VoucherItem[]`; KHÔNG có khóa nào tên cookie/token/session/auth (test introspection payload).
4. `fetchCartViaApi` dùng `credentials: "include"` và có timeout < 30s (AbortController).
5. Khi internal JSON trả 200 hợp lệ -> đọc đúng item tối thiểu (productId/price/qty) từ JSON mẫu.
6. Khi internal JSON lỗi (non-200) -> fallback đọc DOM; parse DOM mẫu trả đúng item.
7. Khi selector DOM chính trượt -> dùng selector dự phòng; vẫn parse được item từ DOM "biến thể A/B".
8. Khi cả JSON lẫn mọi selector hỏng -> phát health signal (broke: "cart") VÀ trả `items: []` (không ném lỗi vỡ tab).
9. Khi người dùng chưa đăng nhập (JSON auth-fail + DOM giỏ rỗng) -> trả message rỗng lịch sự, không lỗi, không thử đọc credential.
10. Content script KHÔNG gọi API sửa giỏ/đặt hàng/áp voucher (grep: không có endpoint mutate).
11. `npm test` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/shopee-no-secret-leak.test.ts
import { readCart } from "../src/content/shopee/cart-reader";
import { readFile } from "fs/promises";

test("payload KHÔNG chứa cookie/token", async () => {
  const msg = await readCart();
  const flat = JSON.stringify(msg).toLowerCase();
  for (const banned of ["cookie", "token", "session", "authorization", "password"]) {
    expect(flat).not.toContain(banned);
  }
});

test("mã nguồn KHÔNG đọc document.cookie / password input", async () => {
  for (const f of ["cart-reader", "voucher-reader", "api-client", "dom-selectors", "index"]) {
    const src = await readFile(`src/content/shopee/${f}.ts`, "utf8");
    expect(src).not.toMatch(/document\.cookie/);
    expect(src).not.toMatch(/Set-Cookie|Authorization/i);
    expect(src).not.toMatch(/input\[type=["']?password/i);
  }
});
```

```ts
// extension/test/shopee-cart-reader.test.ts
test("đọc giỏ từ internal JSON khi 200", async () => {
  mockFetchOk(cartJsonFixture);                 // 3 item
  const msg = await readCart();
  expect(msg.items).toHaveLength(3);
  expect(msg.items[0]).toMatchObject({ productId: "90112", price: 89000, qty: 1 });
});

test("fallback DOM khi JSON non-200", async () => {
  mockFetch(403);
  document.body.innerHTML = shopeeCartDomFixtureMain;
  const msg = await readCart();
  expect(msg.items.length).toBeGreaterThan(0);
});
```

```ts
// extension/test/shopee-dom-fallback.test.ts
test("selector chính trượt → dùng selector dự phòng (A/B variant)", () => {
  document.body.innerHTML = shopeeCartDomFixtureVariantB; // class đổi
  const items = readCartFromDom();
  expect(items && items.length).toBeGreaterThan(0);
});

test("parse hỏng hẳn → health signal + items rỗng, không ném lỗi", async () => {
  mockFetch(500);
  document.body.innerHTML = "<div>không khớp gì</div>";
  const spy = jest.fn();
  setHealthReporter(spy);
  const msg = await readCart();
  expect(spy).toHaveBeenCalledWith(expect.objectContaining({ broke: "cart" }));
  expect(msg.items).toEqual([]);
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: cập nhật `manifest.json` (content_scripts shopee) + `types.ts` (CartItem/VoucherItem/CartReadMessage) -> `dom-selectors.ts` (bảng selector dự phòng) -> `api-client.ts` (fetch credentials include, no cookie read) -> `cart-reader.ts` + `voucher-reader.ts` (orchestrate + health signal) -> `normalize.ts` (tối thiểu hóa) -> `index.ts` (entry content script) -> tests. Content script gửi `CartReadMessage` cho service worker (FR-EXT-001 `onMessage`); service worker chuyển cho pipeline tối thiểu hóa (FR-EXT-003). Health signal nối vào FR-SCRAPE-006.

---

## §7 - Phụ thuộc

- **FR-EXT-001** - scaffold MV3 + messaging + storage phải có trước (content script gửi message cho SW).
- **FR-EXT-003 (downstream)** - pipeline tối thiểu hóa nhận `CartReadMessage`, lọc lần hai trước khi gửi backend.
- **FR-EXT-007 / FR-EXT-008 (downstream, P2)** - content script TikTok Shop / Lazada tái dùng khung reader + normalize + health signal này.
- **FR-CART-002 / FR-CART-005 / FR-CART-006 (downstream, P2)** - cart_snapshot, auto-test mã, checklist xu dựa dữ liệu giỏ đọc ở đây.
- **FR-SCRAPE-006** - DOM-change monitoring nhận health signal khi parser hỏng.
- **FR-SCRAPE-002** - adapter Shopee backend (`/api/v4/...`, `is_login:false`) song song; reader này là phía client.
- **FR-TRUST-002 / FR-TRUST-003** - local-first + audit độc lập chứng minh không gửi cookie/mật khẩu.

---

## §8 - Payload ví dụ

### Message content -> service worker (tối thiểu, KHÔNG credential)

```json
{
  "type": "CART_READ",
  "platform": "shopee",
  "items": [
    { "productId": "90112", "price": 89000, "qty": 1 },
    { "productId": "77310", "price": 245000, "qty": 2 }
  ],
  "vouchers": [
    { "code": "FREESHIPXTRA", "minSpend": 0, "discountText": "Freeship đến 15k" }
  ]
}
```

### Health signal khi parser hỏng (-> FR-SCRAPE-006)

```json
{ "type": "PARSER_HEALTH", "platform": "shopee", "broke": "cart", "source": "dom", "ts": 1782000000000 }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Có đọc thêm trang sản phẩm (PDP) ngoài giỏ hàng hay không - slice sau; slice 1 nhắm giỏ + voucher.
- Ngưỡng nhịp đọc tự động (debounce theo thao tác giỏ) - tinh chỉnh ở FR-EXT-005 khi gắn đồng bộ realtime.
- Chuẩn hóa voucher điều kiện phức (bậc thang) - hoãn tới FR-CART-001 (voucher_catalog) ở P2.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Đọc document.cookie / token | test no-secret-leak (grep + payload) | rò rỉ credential -> thảm họa PDPL/niềm tin | Cấm về mã; chỉ credentials:'include' (DEC-EXT-08/09) |
| Thu thập mật khẩu | grep input[type=password] | vi phạm cam kết lõi | Không hook form login (DEC-EXT-08) |
| DOM A/B đổi, selector chính trượt | health signal + dom-fallback test | mất dữ liệu giỏ | Selector dự phòng + ưu tiên JSON (DEC-EXT-11) |
| Parse hỏng hẳn nuốt im lặng | thiếu health signal | mất tính năng âm thầm | Bắt buộc phát health -> FR-SCRAPE-006 (§1 #7) |
| fetch internal endpoint treo >30s | timeout AbortController | SW liên đới bị kill | Timeout 25s (§1 #10) |
| Người dùng chưa đăng nhập | giỏ rỗng/auth-fail | nhầm là lỗi | Báo "chưa có dữ liệu giỏ" lịch sự (§1 #11) |
| Tự sửa giỏ/đặt hàng | grep endpoint mutate | đặt nhầm đơn + nghi lạm dụng | Chỉ đọc (DEC-EXT-12); mutate tách FR-CART-005 |
| Gửi trường thừa về backend | introspection payload | lộ dữ liệu ngoài tối thiểu | normalize.ts cắt + FR-EXT-003 lọc lần hai |
| Shopee đổi đường dẫn /api/v4 | api trả non-200 | mất đường JSON | Fallback DOM + cập nhật endpoint qua FR-SCRAPE-006 |

---

## §11 - Ghi chú

- Đây là bề mặt khả dụng tối thiểu của extension và là FR rủi ro niềm tin cao nhất: session piggyback đúng nghĩa = mượn cookie qua trình duyệt (`credentials:'include'`), KHÔNG đọc cookie ra.
- Cấm `document.cookie` về mã (grep test) biến cam kết "không chạm credential" thành kiểm chứng được - tiền đề cho audit độc lập FR-TRUST-003.
- Ưu tiên internal JSON hơn DOM cho ổn định; DOM là đường lui; health signal nối FR-SCRAPE-006 để vá nhanh khi A/B đổi.
- Chỉ-đọc là ranh giới cứng; mọi hành động (áp voucher) phải user-initiated, tách sang FR-CART-005.
- Khung reader + normalize + health signal ở đây được TikTok Shop (FR-EXT-007) và Lazada (FR-EXT-008) tái dùng - chỉ thay lớp selector/đường đọc per-sàn.

---

*Hết FR-EXT-002. Status: ready_to_implement (mục tiêu audit 10/10).*
