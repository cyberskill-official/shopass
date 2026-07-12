---
id: FR-EXT-003
title: "Pipeline tối thiểu hóa dữ liệu client - chỉ gửi productId/price/qty về backend; KHÔNG cookie/token/PII; local-first, allowlist whitelist trường + lọc lần hai"
module: EXT
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-EXT-001, FR-EXT-002, FR-EXT-005, FR-AUTH-003, FR-TRUST-002, FR-TRUST-003, FR-COMPLY-005, NFR-EXT-001]
depends_on: [FR-EXT-002]
blocks: [FR-AFFIL-004, FR-CART-002, FR-EXT-005, FR-TRUST-002, FR-TRUST-003]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (gửi backend dạng tối thiểu hóa: chỉ productId, giá, số lượng - KHÔNG gửi cookie)"
  - "docs/... §3.8 (bảo mật no-cleartext/token không rời client), §5.4 (xử lý dữ liệu tối thiểu hóa, local-first)"
source_decisions:
  - "DEC-EXT-13: backend chỉ nhận tập tối thiểu {platform, productId, price, qty} (+voucher code hiển thị); CẤM cookie/token/header/PII rời client"
  - "DEC-EXT-14: lọc theo ALLOWLIST (whitelist trường) chứ không denylist - chỉ trường được liệt kê tường minh mới được phép đi tiếp; mọi trường khác bị loại"
  - "DEC-EXT-15: pipeline là lần lọc THỨ HAI (defense in depth) sau khi content script (FR-EXT-002) đã tối thiểu hóa - không tin tầng trên, tự lọc lại"
  - "DEC-EXT-16: local-first - chuẩn hóa/khử trùng diễn ra trên client; chỉ payload đã làm sạch mới đưa cho lớp đồng bộ (FR-EXT-005)"
  - "DEC-EXT-17: schema payload có validator runtime; payload không khớp schema bị từ chối (fail-closed), không gửi 'cứ gửi đại'"

language: "TypeScript 5.x; Manifest V3 service worker; validator schema (zod hoặc kiểm tay)"
service: shopass/extension/
new_files:
  - extension/src/pipeline/minimize.ts
  - extension/src/pipeline/allowlist.ts
  - extension/src/pipeline/schema.ts
  - extension/src/pipeline/redact.ts
  - extension/test/minimize.test.ts
  - extension/test/allowlist.test.ts
  - extension/test/no-pii-leak.test.ts
modified_files:
  - extension/src/background/service-worker.ts   # nối onMessage(CART_READ) -> minimize -> queue (FR-EXT-005)
  - extension/src/shared/types.ts                # thêm OutboundPayload (tập tối thiểu)
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - dùng denylist (gửi mọi thứ trừ vài trường) thay allowlist (vi phạm DEC-EXT-14 - dễ lọt trường mới)
  - cho phép cookie/token/header/PII vào OutboundPayload (vi phạm DEC-EXT-13)
  - gửi payload không qua validator schema (vi phạm DEC-EXT-17 - fail-open nguy hiểm)
  - tin payload tầng content mà không lọc lại (vi phạm DEC-EXT-15)

effort_hours: 6
sub_tasks:
  - "0.5h: schema.ts - OutboundPayload schema (platform enum, items[{productId,price,qty}], vouchers[{code,...}])"
  - "1.0h: allowlist.ts - whitelist trường tường minh cho item/voucher/payload; loại trường ngoài danh sách"
  - "1.5h: minimize.ts - nhận CartReadMessage -> allowlist filter -> validate schema -> OutboundPayload; fail-closed nếu lệch"
  - "0.5h: redact.ts - quét chuỗi nghi credential/PII (cookie-like, email, phone) -> từ chối hoặc loại"
  - "0.5h: nối service-worker onMessage(CART_READ) -> minimize -> đẩy queue cho FR-EXT-005"
  - "1.0h: no-pii-leak.test.ts - bơm payload có cookie/email/token thừa -> bị loại; OutboundPayload sạch"
  - "1.0h: minimize.test.ts + allowlist.test.ts - trường lạ bị loại; thiếu trường bắt buộc -> reject; happy path đúng tập tối thiểu"

risk_if_skipped: "Đây là van an toàn cuối cùng trước khi dữ liệu rời máy người dùng - cam kết PDPL + niềm tin hậu-Honey của SănDeal phụ thuộc vào nó. Tài liệu nguồn (§3.2/§5.4) nêu rõ chỉ gửi productId/giá/số lượng và KHÔNG gửi cookie. Nếu pipeline dùng denylist (gửi mọi thứ trừ vài trường), một trường mới do tầng trên vô tình thêm sẽ lọt ra backend - đúng kiểu rò rỉ âm thầm. Allowlist (chỉ cho qua trường liệt kê) là an toàn theo mặc định. Đây phải là lần lọc thứ hai độc lập: nếu content script (FR-EXT-002) có bug để lọt cookie, pipeline vẫn chặn (defense in depth). Không có van này, một thay đổi vô hại ở tầng đọc có thể biến extension thành công cụ rò rỉ dữ liệu cá nhân - phá hủy toàn bộ định vị và mời chế tài PDPL tới 5% doanh thu."
---

## §1 - Mô tả (BCP-14 normative)

Pipeline tối thiểu hóa **MUST** là van an toàn cuối: nhận dữ liệu từ content script, lọc theo allowlist trường, validate schema, và CHỈ cho tập tối thiểu `{platform, productId, price, qty}` (+ voucher code hiển thị) rời máy client. Hợp đồng:

1. Pipeline **MUST** chỉ sản sinh `OutboundPayload` chứa tập tối thiểu: `platform` (enum sàn), `items: [{ productId, price, qty }]`, `vouchers: [{ code, minSpend?, discountText? }]` (DEC-EXT-13). Không trường nào khác được phép.
2. Lọc **MUST** theo ALLOWLIST (whitelist trường), KHÔNG theo denylist (DEC-EXT-14): chỉ trường được liệt kê tường minh trong `allowlist.ts` mới đi tiếp; mọi trường khác (kể cả trường lạ tầng trên vô tình thêm) bị loại.
3. Pipeline **MUST** là lần lọc THỨ HAI độc lập (DEC-EXT-15, defense in depth): không tin payload từ content script (FR-EXT-002) đã sạch - tự lọc lại từ đầu. Nếu tầng trên có bug để lọt credential, pipeline vẫn chặn.
4. `OutboundPayload` **MUST NOT** chứa, dưới bất kỳ tên trường nào: cookie, session token, header xác thực, hay PII (email, số điện thoại, tên, địa chỉ, định danh người dùng sàn thật) (DEC-EXT-13).
5. Mọi payload **MUST** qua validator schema runtime (`schema.ts`) trước khi rời client; payload không khớp schema **MUST** bị từ chối (fail-closed) - KHÔNG có nhánh "gửi đại khi không chắc" (DEC-EXT-17).
6. `redact.ts` **MUST** quét giá trị chuỗi trong payload tìm dấu hiệu credential/PII (chuỗi giống cookie/token dài, giống email, giống số điện thoại); khi phát hiện, trường đó **MUST** bị loại hoặc cả payload bị từ chối - không để lọt qua kẽ allowlist (ví dụ một `productId` bị nhồi giá trị lạ).
7. Toàn bộ chuẩn hóa/khử trùng **MUST** diễn ra trên client (local-first, DEC-EXT-16, FR-TRUST-002); backend chỉ nhận kết quả đã làm sạch. KHÔNG gửi dữ liệu thô để "backend lọc giúp".
8. `productId` **MUST** được coi là ID sản phẩm công khai, không phải định danh người dùng; pipeline **MUST** kiểm `productId` khớp mẫu ID hợp lệ (chữ-số) và từ chối giá trị bất thường (tránh kênh rò rỉ qua trường tưởng vô hại).
9. `price`/`qty`/`minSpend` **MUST** là số nguyên không âm trong ngưỡng hợp lý; giá trị ngoài ngưỡng -> loại item đó (dữ liệu rác không nên rời client).
10. Pipeline **MUST** ghi metric đếm: số payload đi tiếp, số trường bị loại bởi allowlist, số payload bị từ chối bởi schema/redact - để FR-TRUST-003 (audit) và giám sát có bằng chứng định lượng pipeline đang lọc.
11. Service worker **MUST** nối `onMessage(CART_READ)` -> `minimize()` -> đẩy `OutboundPayload` vào hàng đợi đồng bộ (FR-EXT-005). Không có đường tắt nào gửi `CartReadMessage` thẳng ra mạng bỏ qua pipeline.
12. Pipeline **MUST** không phụ thuộc trạng thái đăng nhập/định danh: nó chỉ biến đổi dữ liệu, không gắn token. Việc gắn danh tính/JWT cho request là của lớp đồng bộ FR-EXT-005 (và token đó là JWT của SănDeal, KHÔNG phải token sàn).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao allowlist chứ không denylist (DEC-EXT-14)?** Denylist ("gửi mọi thứ trừ cookie, token...") an toàn đúng tại thời điểm viết, rồi hỏng ngay khi ai đó thêm một trường mới ở tầng trên - trường đó không nằm trong danh sách cấm nên lọt ra. Allowlist đảo mặc định: chỉ trường được kê tên mới đi tiếp; trường mới mặc nhiên bị chặn cho tới khi có người chủ động thêm vào allowlist (và review). Với dữ liệu cá nhân, mặc định an toàn là bắt buộc.

**Vì sao lọc lần hai độc lập (DEC-EXT-15)?** Content script (FR-EXT-002) đã tối thiểu hóa, nhưng đó là cùng một codebase có thể có bug. Defense in depth: pipeline coi như input có thể bẩn và lọc lại từ đầu. Nếu một thay đổi ở tầng đọc vô tình để lọt trường nhạy cảm, van thứ hai này vẫn chặn trước khi dữ liệu rời máy.

**Vì sao fail-closed qua validator (DEC-EXT-17)?** Khi payload không khớp schema kỳ vọng, lựa chọn an toàn là từ chối, không phải "gửi đại rồi sửa sau". Dữ liệu cá nhân đã rời máy là không thu hồi được. Fail-closed đắt hơn cho dev (phải xử lý reject) nhưng là tư thế đúng cho quyền riêng tư.

**Vì sao redact quét cả giá trị, không chỉ tên trường (§1 #6)?** Allowlist chặn theo tên trường, nhưng một kẻ tấn công (hoặc bug) có thể nhồi cookie vào một trường được phép như `productId`. Quét giá trị tìm pattern credential/PII bịt kẽ hở đó - kiểm cả "trường nào" lẫn "nội dung gì".

**Vì sao local-first (DEC-EXT-16)?** "Backend lọc giúp" nghĩa là dữ liệu thô đã rời máy - đã quá muộn. Triết lý SănDeal (§5.4) là xử lý tối thiểu hóa ngay trên client; chỉ kết quả sạch mới truyền đi. Đây vừa là cam kết kỹ thuật vừa là điểm bán niềm tin.

**Vì sao tách việc gắn token sang FR-EXT-005 (§1 #12)?** Pipeline chỉ làm sạch dữ liệu, không biết gì về danh tính. Việc đính JWT (của SănDeal, không phải sàn) cho request là trách nhiệm của lớp đồng bộ. Tách bạch giữ pipeline đơn giản, dễ audit, và tránh lẫn token SănDeal với token sàn.

---

## §3 - Hợp đồng API / DDL

### schema.ts (OutboundPayload + validator)

```ts
// extension/src/pipeline/schema.ts
export interface OutboundItem  { productId: string; price: number; qty: number; }
export interface OutboundVoucher { code: string; minSpend?: number; discountText?: string; }
export interface OutboundPayload {
  platform: "shopee" | "tiktok" | "lazada";
  items: OutboundItem[];
  vouchers: OutboundVoucher[];
  // KHÔNG trường nào khác. Mọi mở rộng phải sửa schema + allowlist + review.
}

const ID_RE = /^[A-Za-z0-9._-]{1,64}$/;

export function validatePayload(p: unknown): p is OutboundPayload {
  const o = p as OutboundPayload;
  if (!o || !["shopee", "tiktok", "lazada"].includes(o.platform)) return false;
  if (!Array.isArray(o.items) || !Array.isArray(o.vouchers)) return false;
  return o.items.every(it =>
    ID_RE.test(it.productId) &&
    Number.isInteger(it.price) && it.price >= 0 && it.price < 1e12 &&
    Number.isInteger(it.qty) && it.qty >= 0 && it.qty < 1e6);
}
```

### allowlist.ts (whitelist trường)

```ts
// extension/src/pipeline/allowlist.ts
const ITEM_FIELDS    = ["productId", "price", "qty"] as const;
const VOUCHER_FIELDS = ["code", "minSpend", "discountText"] as const;

export function pickItem(raw: Record<string, unknown>): OutboundItem {
  return pick(raw, ITEM_FIELDS) as OutboundItem;   // CHỈ trường được kê
}
function pick<T extends string>(o: Record<string, unknown>, keys: readonly T[]) {
  const out: Record<string, unknown> = {};
  for (const k of keys) if (k in o) out[k] = o[k];
  return out;
}
```

### minimize.ts (allowlist -> redact -> validate -> fail-closed)

```ts
// extension/src/pipeline/minimize.ts
export function minimize(msg: CartReadMessage): OutboundPayload | null {
  const items    = msg.items.map(i => pickItem(i as any)).filter(isCleanItem);
  const vouchers = msg.vouchers.map(v => pickVoucher(v as any)).filter(noCredentialLike);
  const payload: OutboundPayload = { platform: msg.platform, items, vouchers };
  if (!validatePayload(payload)) { metrics.rejected("schema"); return null; } // fail-closed
  if (containsPiiOrCredential(payload)) { metrics.rejected("redact"); return null; }
  metrics.passed(items.length);
  return payload;
}
```

---

## §4 - Acceptance criteria

1. `OutboundPayload` chỉ có khóa `platform`, `items`, `vouchers`; `items` chỉ `{productId,price,qty}`; `vouchers` chỉ `{code,minSpend?,discountText?}` (test introspection).
2. Bơm `CartReadMessage` có trường thừa (ví dụ `cookie`, `userId`, `email` nhồi vào item) -> allowlist loại; `OutboundPayload` không chứa chúng.
3. Lọc là allowlist: thêm một khóa lạ mới vào input -> mặc nhiên bị loại, không cần sửa code lọc (test khẳng định khóa lạ biến mất).
4. Payload không khớp schema (thiếu `productId`, `price` âm, `platform` lạ) -> `minimize` trả `null` (fail-closed), metric `rejected("schema")` tăng.
5. Giá trị nghi credential/PII nhồi vào trường được phép (ví dụ `productId` = chuỗi cookie dài / email) -> bị `redact`/`validate` loại; metric `rejected("redact")` hoặc item bị lọc.
6. `price`/`qty` ngoài ngưỡng (âm, quá lớn) -> item bị loại, không rời client.
7. Grep `pipeline/**`: KHÔNG có denylist kiểu "gửi mọi thứ trừ"; chỉ pick theo danh sách trường.
8. Không tồn tại đường gửi `CartReadMessage` ra mạng bỏ qua `minimize` (grep: mọi outbound đi qua pipeline -> queue FR-EXT-005).
9. Metric đếm passed / allowlist-dropped / rejected hoạt động.
10. Happy path: input giỏ hợp lệ -> `OutboundPayload` đúng tập tối thiểu, validate pass.
11. `npm test` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/no-pii-leak.test.ts
import { minimize } from "../src/pipeline/minimize";

test("trường thừa (cookie/email/userId) bị loại khỏi OutboundPayload", () => {
  const msg: any = {
    type: "CART_READ", platform: "shopee",
    items: [{ productId: "90112", price: 89000, qty: 1, cookie: "SPC_=abc", userId: 7, email: "a@b.vn" }],
    vouchers: []
  };
  const out = minimize(msg)!;
  const flat = JSON.stringify(out).toLowerCase();
  expect(flat).not.toMatch(/cookie|userid|email|@/);
  expect(out.items[0]).toEqual({ productId: "90112", price: 89000, qty: 1 });
});

test("credential nhồi vào productId bị từ chối (fail-closed)", () => {
  const msg: any = { type: "CART_READ", platform: "shopee",
    items: [{ productId: "SPC_SESSION_eyJhbGciOi...verylong", price: 1, qty: 1 }], vouchers: [] };
  expect(minimize(msg)).toBeNull();   // ID_RE/redact loại
});
```

```ts
// extension/test/allowlist.test.ts
test("khóa lạ mới mặc nhiên bị loại (allowlist, không cần sửa code)", () => {
  const out = minimize({ type:"CART_READ", platform:"shopee",
    items:[{ productId:"1", price:1, qty:1, brandNewField:"x" } as any], vouchers:[] })!;
  expect("brandNewField" in out.items[0]).toBe(false);
});
```

```ts
// extension/test/minimize.test.ts
test("schema lỗi → null (fail-closed) + metric", () => {
  expect(minimize({ type:"CART_READ", platform:"vng" as any, items:[], vouchers:[] })).toBeNull();
});

test("happy path đúng tập tối thiểu", () => {
  const out = minimize({ type:"CART_READ", platform:"shopee",
    items:[{ productId:"90112", price:89000, qty:1 }], vouchers:[{ code:"FREESHIP" }] })!;
  expect(out).toEqual({ platform:"shopee",
    items:[{ productId:"90112", price:89000, qty:1 }], vouchers:[{ code:"FREESHIP" }] });
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `schema.ts` (OutboundPayload + validatePayload) -> `allowlist.ts` (pick theo danh sách trường) -> `redact.ts` (quét giá trị credential/PII) -> `minimize.ts` (allowlist -> redact -> validate, fail-closed) -> nối `service-worker.ts` `onMessage(CART_READ) -> minimize -> queue` -> tests. Lớp đồng bộ (FR-EXT-005) nhận `OutboundPayload` đã sạch từ hàng đợi và đính JWT SănDeal. Mọi mở rộng trường trong tương lai phải sửa cả `schema.ts` lẫn `allowlist.ts` và đi qua review - đó là điểm kiểm soát duy nhất.

---

## §7 - Phụ thuộc

- **FR-EXT-002** - content script Shopee cung cấp `CartReadMessage` (đã tối thiểu hóa lần một).
- **FR-EXT-001** - service worker + messaging + storage làm khung nối pipeline.
- **FR-EXT-005 (downstream)** - lớp đồng bộ nhận `OutboundPayload` sạch, đính JWT, gửi backend.
- **FR-TRUST-002 (downstream)** - chính sách local-first + tối thiểu hóa được hiện thực ở đây.
- **FR-TRUST-003 (downstream)** - audit độc lập dùng metric + test pipeline làm bằng chứng không gửi cookie/PII.
- **FR-COMPLY-005** - cưỡng chế no-cleartext + token-not-on-server; pipeline là điểm thực thi phía client.
- **FR-CART-002 / FR-AFFIL-004 (downstream, P2)** - cart_snapshot + guardrails affiliate dựa dữ liệu sạch từ pipeline.

---

## §8 - Payload ví dụ

### Vào (CartReadMessage, có thể "bẩn") -> ra (OutboundPayload, sạch)

```ts
// VÀO (giả định tầng trên lỡ thêm trường thừa)
{ type:"CART_READ", platform:"shopee",
  items:[{ productId:"90112", price:89000, qty:1, cookie:"SPC_=x", shopId:55 }],
  vouchers:[{ code:"FREESHIPXTRA", minSpend:0, discountText:"đến 15k", internalId:9 }] }

// RA (sau minimize: chỉ trường allowlist, đã validate)
{ platform:"shopee",
  items:[{ productId:"90112", price:89000, qty:1 }],
  vouchers:[{ code:"FREESHIPXTRA", minSpend:0, discountText:"đến 15k" }] }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Có hash `productId` thêm một lớp trước khi gửi hay không - không cần (productId là công khai); xét lại nếu B2B (FR-B2B-001) yêu cầu ẩn danh mạnh hơn.
- Nén/gộp payload nhiều lần đọc thành batch - tối ưu ở FR-EXT-005 (đồng bộ), không thuộc pipeline lọc.
- Mở rộng allowlist cho trường mới (ví dụ rating) - chỉ khi có FR tiêu thụ; thêm phải qua review.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Dùng denylist thay allowlist | grep test (§4 #7) | trường mới lọt ra backend | Allowlist pick-by-name (DEC-EXT-14) |
| Cookie/token/PII trong payload | no-pii-leak test | rò rỉ dữ liệu cá nhân, chế tài PDPL | Allowlist + redact + validate (§1 #4/#6) |
| Credential nhồi vào trường được phép | redact test | rò qua kẽ allowlist | Quét giá trị, không chỉ tên trường (§1 #6) |
| Payload lệch schema vẫn gửi (fail-open) | minimize trả null test | gửi dữ liệu rác/nhạy cảm | Fail-closed qua validator (DEC-EXT-17) |
| Tin tầng content không lọc lại | review + defense-in-depth test | bug tầng trên lọt thẳng | Lọc lần hai độc lập (DEC-EXT-15) |
| Đường tắt gửi CartReadMessage thô | grep outbound (§4 #8) | bỏ qua van an toàn | Mọi outbound qua minimize -> queue (§1 #11) |
| Gửi thô để "backend lọc" | review | dữ liệu thô đã rời máy | Local-first, lọc trên client (DEC-EXT-16) |
| productId bất thường (kênh ẩn) | ID_RE validate | rò qua trường vô hại | Kiểm mẫu ID + redact (§1 #8) |

---

## §11 - Ghi chú

- Pipeline là van an toàn cuối trước khi dữ liệu rời máy - allowlist (mặc định an toàn) + lọc lần hai + fail-closed là ba trụ của nó.
- Quét giá trị (không chỉ tên trường) bịt kẽ hở "nhồi credential vào trường được phép" - kiểm cả tên lẫn nội dung.
- Local-first nghĩa là dữ liệu thô KHÔNG BAO GIỜ rời máy; chỉ kết quả sạch truyền đi - điểm bán niềm tin của §5.4.
- Mọi mở rộng trường tương lai phải sửa schema + allowlist và qua review: đây là điểm kiểm soát quyền riêng tư duy nhất, cố ý làm "khó nới lỏng".
- Metric đếm passed/dropped/rejected cung cấp bằng chứng định lượng cho audit độc lập FR-TRUST-003.
- Token gắn cho request là JWT của SănDeal (FR-EXT-005), KHÔNG phải token sàn - pipeline không chạm danh tính.

---

*Hết FR-EXT-003. Status: ready_to_implement (mục tiêu audit 10/10).*
