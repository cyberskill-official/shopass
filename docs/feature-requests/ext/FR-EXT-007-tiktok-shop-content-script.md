---
id: FR-EXT-007
title: "Content script TikTok Shop - đọc DOM giỏ trong webview/SPA, tránh API ký msToken/_signature/X-Bogus + app attestation; tái dùng khung reader/normalize/health của FR-EXT-002"
module: EXT
priority: MUST
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-EXT-002, FR-EXT-003, FR-EXT-006, FR-SCRAPE-007, FR-SCRAPE-006, FR-TRUST-002, NFR-EXT-001]
depends_on: [FR-EXT-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (TikTok Shop: cart/checkout trong webview/SPA -> content script đọc DOM; cơ chế ký request msToken/_signature/X-Bogus + app attestation mạnh -> ưu tiên đọc DOM render thay vì gọi API ký)"
  - "docs/... §5.4 (token không rời client), §6.1 (TikTok Shop 41,31% GMV - sàn ưu tiên P2)"
source_decisions:
  - "DEC-EXT-35: TikTok Shop có cart/checkout nằm trong webview/SPA -> content script đọc DOM giỏ đã render; KHÔNG gọi internal API như Shopee vì API TikTok ký bằng msToken/_signature/X-Bogus + app attestation"
  - "DEC-EXT-36: TUYỆT ĐỐI KHÔNG tự sinh/ký msToken/_signature/X-Bogus (ngược-kỹ-nghệ chữ ký) - vừa giòn (đổi liên tục) vừa rủi ro ToS/pháp lý cao; chỉ đọc DOM đã render trong tab đăng nhập"
  - "DEC-EXT-37: tái dùng khung reader + normalize + health signal của FR-EXT-002; chỉ thay lớp selector/đường đọc đặc thù TikTok (SPA route đổi -> quan sát DOM thay đổi, không dựa internal JSON)"
  - "DEC-EXT-38: SPA TikTok điều hướng client-side (route đổi không tải lại trang) -> reader theo dõi DOM mutation/route change để biết khi giỏ render xong; selector nhiều phương án (A/B + đổi cấu trúc)"
  - "DEC-EXT-39: giữ nguyên cam kết niềm tin FR-EXT-002: KHÔNG mật khẩu, token/cookie KHÔNG rời client; chỉ trích productId/giá/qty đã render; chỉ đọc, KHÔNG sửa giỏ/đặt hàng"

language: "TypeScript 5.x; Manifest V3 content script; MutationObserver cho SPA; tái dùng shared/normalize"
service: shopass/extension/
new_files:
  - extension/src/content/tiktok/cart-reader.ts
  - extension/src/content/tiktok/dom-selectors.ts
  - extension/src/content/tiktok/spa-observer.ts
  - extension/src/content/tiktok/index.ts
  - extension/test/tiktok-cart-reader.test.ts
  - extension/test/tiktok-no-api-sign.test.ts
  - extension/test/tiktok-spa-observer.test.ts
modified_files:
  - extension/manifest.json                 # thêm content_scripts cho tiktok.com + host_permissions
  - extension/src/shared/types.ts           # tái dùng CartItem/VoucherItem; thêm platform "tiktok" nơi cần
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - tự sinh/ký msToken/_signature/X-Bogus hay gọi internal API ký của TikTok (vi phạm DEC-EXT-36 - giòn + rủi ro ToS/pháp lý)
  - đọc/sao chép cookie/session token vào biến hay payload (vi phạm DEC-EXT-39 - phá cam kết niềm tin FR-EXT-002)
  - thu thập mật khẩu hoặc tự sửa giỏ/đặt hàng (vi phạm DEC-EXT-39 - chỉ đọc)
  - phân nhánh khung reader riêng thay vì tái dùng normalize/health của FR-EXT-002 (vi phạm DEC-EXT-37)

effort_hours: 10
sub_tasks:
  - "1.0h: manifest content_scripts matches https://*.tiktok.com/* (trang shop) + host_permissions; run_at document_idle"
  - "2.0h: dom-selectors.ts - bảng selector nhiều phương án cho item/giá/qty giỏ TikTok Shop (A/B + đổi cấu trúc)"
  - "2.5h: spa-observer.ts - MutationObserver + theo dõi route SPA; phát sự kiện khi giỏ render xong"
  - "2.0h: cart-reader.ts - đọc DOM giỏ (KHÔNG gọi API ký); orchestrate qua khung FR-EXT-002; health signal khi hỏng"
  - "0.5h: index.ts - entry content script TikTok; gửi CartReadMessage(platform:tiktok) cho SW"
  - "1.0h: tiktok-no-api-sign.test.ts - KHÔNG msToken/_signature/X-Bogus; KHÔNG document.cookie"
  - "1.0h: tiktok-cart-reader.test.ts + tiktok-spa-observer.test.ts - parse DOM mẫu; SPA route đổi -> đọc lại; parse hỏng -> health"

risk_if_skipped: "TikTok Shop chiếm 41,31% GMV TMĐT VN (§6.1) - bỏ sàn này là bỏ gần nửa thị trường, moat so sánh chéo 3 sàn không thành. Nhưng TikTok là sàn khó nhất về kỹ thuật: cart nằm trong webview/SPA và API nội bộ ký bằng msToken/_signature/X-Bogus cùng app attestation mạnh (§3.2). Cám dỗ kỹ thuật là ngược-kỹ-nghệ chữ ký để gọi API - đây là cái bẫy: chữ ký đổi liên tục (giòn, vỡ mỗi lần TikTok cập nhật) VÀ rủi ro ToS/pháp lý cao (phơi bày §5.5). Đường đúng theo tài liệu là đọc DOM đã render trong tab đăng nhập của chính người dùng - giống session piggyback Shopee nhưng không có đường internal JSON. Vì TikTok là SPA, reader phải theo dõi DOM mutation/route change (trang không tải lại khi đổi route) nếu không sẽ đọc nhầm trạng thái cũ. Nếu phân nhánh khung reader riêng thay vì tái dùng normalize/health của FR-EXT-002, ta nhân đôi bug và mất tính nhất quán tối thiểu hóa dữ liệu. FR này mở moat đa sàn mà vẫn giữ cam kết niềm tin và né rủi ro chữ ký."
---

## §1 - Mô tả (BCP-14 normative)

Content script TikTok Shop **MUST** đọc giỏ hàng qua DOM đã render trong webview/SPA của tab đăng nhập, tránh tuyệt đối API ký (msToken/_signature/X-Bogus), và tái dùng khung reader/normalize/health của FR-EXT-002. Hợp đồng:

1. `manifest.json` **MUST** khai `content_scripts` với `matches` cho domain TikTok Shop (ví dụ `https://*.tiktok.com/*` giới hạn trang shop), `run_at: "document_idle"`, và thêm `host_permissions` tương ứng (per-domain, không `<all_urls>`, theo FR-EXT-001).
2. Content script **MUST** đọc dữ liệu giỏ bằng cách parse DOM đã render (DEC-EXT-35) - TikTok không có đường internal JSON tiện như Shopee vì cart nằm trong webview/SPA và API ký mạnh.
3. Content script **MUST NOT** tự sinh, ký, hay gửi `msToken`/`_signature`/`X-Bogus` hay gọi internal API ký của TikTok (DEC-EXT-36). Ngược-kỹ-nghệ chữ ký bị cấm: vừa giòn (đổi liên tục) vừa rủi ro ToS/pháp lý.
4. Vì TikTok là SPA (điều hướng client-side, route đổi không tải lại trang), `spa-observer.ts` **MUST** dùng `MutationObserver` và/hoặc theo dõi route change để biết khi giỏ render xong, rồi mới đọc (DEC-EXT-38) - không đọc một lần lúc load rồi bỏ.
5. `dom-selectors.ts` **MUST** cung cấp nhiều selector dự phòng cho mỗi trường (item/giá/qty) vì DOM TikTok đổi theo A/B test + cấu trúc SPA (DEC-EXT-38). Selector chính trượt -> thử dự phòng trước khi coi là hỏng.
6. Khi mọi selector hỏng, content script **MUST** phát health signal về service worker để FR-SCRAPE-006 (DOM-change monitoring) ghi nhận - KHÔNG nuốt lỗi im lặng (tái dùng cơ chế FR-EXT-002).
7. Content script **MUST** tái dùng khung reader + `normalize.ts` + health signal của FR-EXT-002 (DEC-EXT-37); chỉ thay lớp selector/đường đọc đặc thù TikTok. **MUST NOT** phân nhánh một pipeline tối thiểu hóa riêng.
8. Content script **MUST** giữ nguyên cam kết niềm tin FR-EXT-002 (DEC-EXT-39): KHÔNG đọc `document.cookie`/token, KHÔNG thu mật khẩu, token/cookie KHÔNG rời client; chỉ trích `productId`/giá/qty đã render.
9. Content script **MUST** chỉ đọc: KHÔNG tự sửa giỏ, KHÔNG đặt hàng, KHÔNG áp voucher tự động (DEC-EXT-39).
10. Kết quả **MUST** gửi cho service worker dạng `CartReadMessage` typed với `platform: "tiktok"`, chứa CHỈ `CartItem[]` + `VoucherItem[]` tối thiểu - đi tiếp qua pipeline FR-EXT-003.
11. Content script **MUST** chỉ chạy khi consent "read_cart" (FR-EXT-006) bật; gọi `ensureConsent` trước khi đọc.
12. Content script **MUST** xử lý chưa-đăng-nhập / giỏ rỗng lịch sự (không lỗi, báo "chưa có dữ liệu giỏ"); và xử lý SPA chưa render xong (chờ observer) thay vì đọc DOM trống rồi báo hỏng nhầm.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao đọc DOM, không gọi API ký (DEC-EXT-35/36)?** TikTok ký request nội bộ bằng msToken/_signature/X-Bogus cùng app attestation - một hệ chống bot tinh vi. Ngược-kỹ-nghệ chữ ký để gọi API là cái bẫy kép: chữ ký đổi mỗi lần TikTok cập nhật (code vỡ liên tục, chi phí bảo trì khổng lồ), VÀ ranh giới ToS/pháp lý nguy hiểm (§5.5). Đọc DOM đã render trong tab đăng nhập của chính người dùng là đường an toàn và bền hơn - đúng tinh thần session piggyback, chỉ khác là không có đường internal JSON.

**Vì sao MutationObserver cho SPA (DEC-EXT-38)?** TikTok là single-page app: bấm vào giỏ đổi route nhưng trang không tải lại, DOM được vẽ lại bằng JS. Reader đọc một lần lúc `document_idle` sẽ bắt nhầm trạng thái cũ hoặc DOM trống. `MutationObserver` + theo dõi route change cho reader biết "giỏ đã render xong, đọc bây giờ" - đây là khác biệt cốt lõi so với content script Shopee (trang truyền thống hơn).

**Vì sao tái dùng khung FR-EXT-002 (DEC-EXT-37)?** FR-EXT-002 đã định nghĩa khung reader + normalize (tối thiểu hóa) + health signal và đã được test cho cam kết "không rò cookie". Phân nhánh một pipeline riêng cho TikTok nghĩa là nhân đôi bug và mở một bề mặt mới có thể lọt credential. Tái dùng giữ một điểm kiểm soát tối thiểu hóa duy nhất; chỉ lớp selector/đường đọc là per-sàn.

**Vì sao giữ nguyên cam kết niềm tin (DEC-EXT-39)?** Ranh giới "không cookie, không mật khẩu, chỉ đọc" không phụ thuộc sàn - nó là định vị hậu-Honey của toàn extension (§5.4). TikTok không phải ngoại lệ. Test no-secret-leak tương tự FR-EXT-002 khẳng định ranh giới này cho cả TikTok, không chỉ tin quy ước.

**Vì sao chờ render thay vì báo hỏng nhầm (§1 #12)?** Trong SPA, DOM trống ngay sau khi đổi route là bình thường (chưa vẽ xong), không phải lỗi parser. Nếu reader báo health "hỏng" mỗi lần gặp DOM trống tạm thời, FR-SCRAPE-006 ngập tín hiệu nhiễu. Chờ observer báo render xong rồi mới đọc/đánh giá hỏng phân biệt đúng "chưa xong" với "thật sự hỏng".

---

## §3 - Hợp đồng API / DDL

### manifest.json (bổ sung content script TikTok)

```jsonc
// extension/manifest.json (thêm)
{
  "host_permissions": ["https://shopee.vn/*", "https://*.tiktok.com/*"],
  "content_scripts": [
    {
      "matches": ["https://*.tiktok.com/*"],
      "js": ["content/tiktok/index.js"],
      "run_at": "document_idle"
    }
  ]
}
```

### spa-observer.ts (chờ giỏ render xong trong SPA)

```ts
// extension/src/content/tiktok/spa-observer.ts
export function onCartRendered(cb: () => void): () => void {
  const obs = new MutationObserver(() => {
    if (document.querySelector(CART_ROOT_SELECTORS.find(s => document.querySelector(s)) ?? "x")) {
      cb();   // giỏ đã render → đọc bây giờ
    }
  });
  obs.observe(document.body, { childList: true, subtree: true });
  return () => obs.disconnect();   // dọn observer
}
```

### cart-reader.ts (đọc DOM, KHÔNG API ký; tái dùng normalize + health FR-EXT-002)

```ts
// extension/src/content/tiktok/cart-reader.ts
import { normalizeCart } from "../shared/normalize";     // tái dùng FR-EXT-002
import { reportHealth } from "../shared/health";         // tái dùng FR-EXT-002

export async function readTiktokCart(): Promise<CartReadMessage> {
  if (!(await ensureConsent("read_cart"))) {             // FR-EXT-006 gate
    return { type: "CART_READ", platform: "tiktok", items: [], vouchers: [] };
  }
  let raw = readCartFromDom();                            // CHỈ DOM; KHÔNG msToken/_signature/X-Bogus
  if (raw === null) {
    reportHealth({ platform: "tiktok", broke: "cart", source: "dom" }); // FR-SCRAPE-006
    raw = [];
  }
  const items = normalizeCart(raw);                      // tối thiểu hóa chung
  return { type: "CART_READ", platform: "tiktok", items, vouchers: readVouchersFromDom() };
  // LƯU Ý: KHÔNG document.cookie, KHÔNG gọi API ký ở đây.
}
```

---

## §4 - Acceptance criteria

1. `manifest.json` có `content_scripts` match domain TikTok Shop, `run_at: document_idle`; `host_permissions` thêm `https://*.tiktok.com/*`, KHÔNG `<all_urls>`.
2. Grep `src/content/tiktok/**`: KHÔNG có `msToken`, `_signature`, `X-Bogus`, KHÔNG gọi internal API ký TikTok.
3. Grep `src/content/tiktok/**`: KHÔNG có `document.cookie`, KHÔNG đọc `input[type=password]`.
4. Reader đọc giỏ từ DOM mẫu TikTok -> trả đúng item tối thiểu (productId/price/qty).
5. Selector chính trượt -> dùng selector dự phòng; vẫn parse được item từ DOM "biến thể".
6. SPA route đổi (giỏ render lại) -> observer kích hoạt đọc lại; reader không kẹt ở trạng thái cũ.
7. Mọi selector hỏng -> phát health signal (broke: "cart") VÀ trả `items: []`, không ném lỗi.
8. Reader tái dùng `normalize.ts` + health của FR-EXT-002 (grep import từ `content/shared/`), không phân nhánh riêng.
9. `CartReadMessage` có `platform: "tiktok"`, chỉ `items` + `vouchers`; không cookie/token (introspection payload).
10. Reader chỉ chạy khi consent "read_cart" bật (gọi `ensureConsent`); chưa bật -> trả rỗng.
11. `npm test` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/tiktok-no-api-sign.test.ts
test("KHÔNG ký/gửi msToken/_signature/X-Bogus, KHÔNG đọc cookie", async () => {
  for (const f of ["cart-reader", "dom-selectors", "spa-observer", "index"]) {
    const src = await readFile(`src/content/tiktok/${f}.ts`, "utf8");
    expect(src).not.toMatch(/msToken|_signature|X-Bogus/i);   // không API ký
    expect(src).not.toMatch(/document\.cookie/);              // không cookie
    expect(src).not.toMatch(/input\[type=["']?password/i);    // không mật khẩu
  }
});

test("payload TikTok KHÔNG chứa cookie/token", async () => {
  setConsent(["read_cart"]);
  document.body.innerHTML = tiktokCartFixtureMain;
  const msg = await readTiktokCart();
  const flat = JSON.stringify(msg).toLowerCase();
  for (const b of ["cookie", "token", "mstoken", "signature", "x-bogus", "password"]) {
    expect(flat).not.toContain(b);
  }
});
```

```ts
// extension/test/tiktok-cart-reader.test.ts
test("đọc giỏ từ DOM TikTok đã render", async () => {
  setConsent(["read_cart"]);
  document.body.innerHTML = tiktokCartFixtureMain;          // 2 item
  const msg = await readTiktokCart();
  expect(msg.platform).toBe("tiktok");
  expect(msg.items.length).toBe(2);
});

test("selector chính trượt → dự phòng (biến thể)", () => {
  document.body.innerHTML = tiktokCartFixtureVariant;       // class đổi
  const items = readCartFromDom();
  expect(items && items.length).toBeGreaterThan(0);
});

test("reader tái dùng normalize/health FR-EXT-002, không phân nhánh", async () => {
  const src = await readFile("src/content/tiktok/cart-reader.ts", "utf8");
  expect(src).toMatch(/from ["']\.\.\/shared\/normalize["']/);
  expect(src).toMatch(/from ["']\.\.\/shared\/health["']/);
});
```

```ts
// extension/test/tiktok-spa-observer.test.ts
test("SPA route đổi → observer kích hoạt đọc lại", async () => {
  const spy = jest.fn();
  const stop = onCartRendered(spy);
  document.body.innerHTML = tiktokCartFixtureMain;          // mô phỏng render
  await flushMutations();
  expect(spy).toHaveBeenCalled();
  stop();
});

test("parse hỏng hẳn → health + items rỗng", async () => {
  setConsent(["read_cart"]);
  document.body.innerHTML = "<div>không khớp</div>";
  const sp = jest.fn(); setHealthReporter(sp);
  const msg = await readTiktokCart();
  expect(sp).toHaveBeenCalledWith(expect.objectContaining({ broke: "cart", platform: "tiktok" }));
  expect(msg.items).toEqual([]);
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: cập nhật `manifest.json` (content_scripts tiktok + host_permissions) -> `dom-selectors.ts` (bảng selector dự phòng TikTok) -> `spa-observer.ts` (MutationObserver/route change báo render xong) -> `cart-reader.ts` (đọc DOM, tái dùng normalize + health FR-EXT-002, gate consent) -> `index.ts` (entry, gửi CartReadMessage platform tiktok) -> tests. KHÔNG chạm msToken/_signature/X-Bogus. Khung reader/normalize/health dùng chung với Shopee (FR-EXT-002) và Lazada (FR-EXT-008); chỉ lớp selector + observer SPA là đặc thù TikTok. Kết quả qua pipeline FR-EXT-003.

---

## §7 - Phụ thuộc

- **FR-EXT-002** - cung cấp khung reader + normalize + health signal tái dùng; ranh giới niềm tin chung.
- **FR-EXT-003** - pipeline tối thiểu hóa nhận `CartReadMessage(platform:tiktok)`, lọc lần hai.
- **FR-EXT-006** - consent "read_cart" gate trước khi đọc.
- **FR-SCRAPE-006** - DOM-change monitoring nhận health signal khi parser TikTok hỏng.
- **FR-SCRAPE-007 (song song, P2)** - adapter TikTok backend (ưu tiên DOM-render, né API ký); reader này là phía client.
- **FR-TRUST-002** - local-first + tối thiểu hóa; reader giữ cam kết.
- **NFR-EXT-001** - content script gửi message cho SW ephemeral; không giữ state global.

---

## §8 - Payload ví dụ

### Message content TikTok -> service worker (tối thiểu, KHÔNG credential)

```json
{
  "type": "CART_READ",
  "platform": "tiktok",
  "items": [
    { "productId": "T-883201", "price": 159000, "qty": 1 },
    { "productId": "T-771045", "price": 49000, "qty": 3 }
  ],
  "vouchers": [
    { "code": "TIKTOKFREESHIP", "minSpend": 0, "discountText": "Freeship 20k" }
  ]
}
```

### Health signal khi parser TikTok hỏng (-> FR-SCRAPE-006)

```json
{ "type": "PARSER_HEALTH", "platform": "tiktok", "broke": "cart", "source": "dom", "ts": 1790000000000 }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đọc trang sản phẩm TikTok (PDP) ngoài giỏ - slice sau; P2 slice 1 nhắm giỏ + voucher.
- Xử lý TikTok in-app webview (ngoài trình duyệt desktop) - hoãn tới mobile (FR-MOBILE-002); FR này nhắm web desktop.
- Ngưỡng debounce observer cho SPA route đổi nhanh liên tiếp - tinh chỉnh khi đo thấy đọc lại quá dày.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Ngược-kỹ-nghệ ký msToken/_signature/X-Bogus | tiktok-no-api-sign test | giòn + rủi ro ToS/pháp lý | Chỉ đọc DOM render (DEC-EXT-36) |
| Đọc document.cookie/token | no-secret-leak test | rò credential, §5.4 | Cấm về mã (DEC-EXT-39) |
| Thu thập mật khẩu | grep password | phá cam kết lõi | Không hook form login (DEC-EXT-39) |
| Đọc DOM trống lúc SPA chưa render | spa-observer test | parse hỏng nhầm | Chờ observer báo render xong (DEC-EXT-38) |
| Selector chính trượt | health + variant test | mất dữ liệu giỏ | Selector dự phòng (DEC-EXT-38) |
| Parse hỏng nuốt im lặng | thiếu health signal | mất tính năng âm thầm | Bắt buộc health -> FR-SCRAPE-006 (§1 #6) |
| Phân nhánh pipeline riêng | grep import shared | nhân đôi bug + bề mặt rò | Tái dùng normalize/health FR-EXT-002 (DEC-EXT-37) |
| Tự sửa giỏ/đặt hàng | grep mutate | đặt nhầm + nghi lạm dụng | Chỉ đọc (DEC-EXT-39) |
| Đọc khi chưa consent | consent gate test | xử lý không cơ sở pháp lý | ensureConsent("read_cart") trước (§1 #11) |
| host_permissions <all_urls> | manifest test | Web Store reject | Per-domain tiktok.com (§1 #1) |

---

## §11 - Ghi chú

- TikTok Shop (41,31% GMV §6.1) là sàn bắt buộc cho moat đa sàn, nhưng khó nhất: cart trong webview/SPA + API ký mạnh.
- Ranh giới cứng: KHÔNG ngược-kỹ-nghệ msToken/_signature/X-Bogus - giòn + rủi ro ToS/pháp lý (§5.5); chỉ đọc DOM render trong tab đăng nhập.
- SPA cần MutationObserver/route tracking - khác biệt cốt lõi so với content script Shopee; đọc một lần lúc load là sai.
- Tái dùng khung reader/normalize/health của FR-EXT-002 giữ một điểm kiểm soát tối thiểu hóa duy nhất; chỉ selector + observer là per-sàn.
- Cam kết niềm tin (không cookie/mật khẩu, chỉ đọc) không phụ thuộc sàn - test no-secret-leak khẳng định cho cả TikTok.

---

*Hết FR-EXT-007. Status: ready_to_implement (mục tiêu audit 10/10).*
