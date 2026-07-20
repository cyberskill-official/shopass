---
id: TASK-WEB-004
title: "UI quản lý wishlist + alert - màn hình CRUD danh sách mong muốn (target_price) và luật cảnh báo (4 rule_type, channel push/email/sms), tiêu thụ API TASK-TRACK-002/003 qua lib/api.ts"
module: WEB
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-WEB-001, TASK-TRACK-002, TASK-TRACK-003, TASK-TRACK-004]
depends_on: [TASK-WEB-001, TASK-TRACK-002, TASK-TRACK-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §6 (catalog tính năng: wishlist + cảnh báo giảm giá)"
  - "docs/... §3.7 (API quản lý wishlist + POST /v1/alerts), §3.4 (wishlist_item.target_price, alert_rule rule_type/channel)"
source_decisions:
  - "DEC-WEB-16: UI wishlist/alert tiêu thụ API REST của TASK-TRACK-002 (/v1/wishlists...) và TASK-TRACK-003 (/v1/alerts...) qua lib/api.ts; không tự lưu trạng thái nguồn ở client, server là nguồn sự thật"
  - "DEC-WEB-17: form alert phản chiếu đúng ngữ nghĩa threshold theo rule_type của TASK-TRACK-003 - price_below cần giá VND, drop_pct cần % (1..99), real_sale/bottom_predicted KHÔNG có ô threshold; validate phía client trước khi gửi để báo sớm, server vẫn validate lần cuối"
  - "DEC-WEB-18: bộ chọn channel chỉ phơi {push, email, sms} (push mặc định) khớp enum kênh TASK-TRACK-003; không cho chọn kênh ngoài tập có dispatcher"
  - "DEC-WEB-19: target_price và threshold price_below nhập/hiển thị là số nguyên VND (int64) - không dùng dấu phẩy động; gửi nguyên số đồng VND lên API"
  - "DEC-WEB-20: UI chỉ thao tác trên tài nguyên của chính user (server đã chặn IDOR); client không cố đoán/duyệt id của user khác; lỗi 403/404 từ server hiển thị 'không tìm thấy' trung lập"

language: "TypeScript 5.x; Next.js 14 (App Router, client components form); gọi API TASK-TRACK-002/003 qua lib/api.ts (TASK-WEB-001)"
service: shopass/web/
new_files:
  - web/app/(app)/wishlist/page.tsx
  - web/app/(app)/alerts/page.tsx
  - web/components/wishlist/wishlist-panel.tsx
  - web/components/wishlist/wishlist-item-row.tsx
  - web/components/alerts/alert-form.tsx
  - web/components/alerts/alert-list.tsx
  - web/lib/wishlist/api.ts
  - web/lib/alerts/api.ts
  - web/lib/alerts/validate.ts
  - web/test/alert-form-validate.test.ts
  - web/test/wishlist-api.test.ts
  - web/test/alert-channel.test.ts
modified_files:
  - web/components/app-shell.tsx             # thêm mục điều hướng Wishlist + Cảnh báo
allowed_tools:
  - file_read: web/**
  - file_write: web/**
  - bash: cd web && npm test && npx tsc --noEmit
disallowed_tools:
  - cho chọn rule_type/channel ngoài enum của TASK-TRACK-003 (vi phạm DEC-WEB-17/18, server trả 400 hoặc luật rác)
  - nhập target_price/threshold price_below dạng float (vi phạm DEC-WEB-19, sai số tiền tệ)
  - render ô threshold cho real_sale/bottom_predicted (vi phạm DEC-WEB-17, ngữ nghĩa sai)
  - tự lưu danh sách wishlist/alert ở client như nguồn sự thật thay vì đọc từ API (vi phạm DEC-WEB-16)

effort_hours: 6
sub_tasks:
  - "0.75h: lib/wishlist/api.ts - createWishlist, listWishlists, addItem, removeItem, deleteWishlist (gọi /v1/wishlists... qua apiFetch)"
  - "0.75h: lib/alerts/api.ts - createAlert, listAlerts, toggleActive, deleteAlert, history (gọi /v1/alerts... qua apiFetch)"
  - "1.0h: lib/alerts/validate.ts - validate threshold/channel theo rule_type (mirror DEC-TRACK-22), báo lỗi sớm phía client"
  - "1.0h: wishlist-panel.tsx + wishlist-item-row.tsx - liệt kê wishlist, thêm/xóa item, sửa target_price (VND int)"
  - "1.0h: alert-form.tsx - chọn rule_type, ô threshold động theo loại, chọn channel (push/email/sms)"
  - "0.5h: alert-list.tsx - liệt kê luật, bật/tắt active, xóa"
  - "0.25h: wishlist/page.tsx + alerts/page.tsx - màn hình trong shell (app)"
  - "1.0h: tests - validate form theo rule_type, channel allowlist, target_price int64, api gọi đúng endpoint"

risk_if_skipped: "wishlist và alert là cách người dùng biến SănDeal từ công cụ tra cứu thành trợ lý chủ động: gom sản phẩm muốn mua, đặt giá mong muốn, và khai báo 'báo tôi khi nào'. Thiếu UI này thì backend TASK-TRACK-002/003 không có mặt tiền - người dùng không tạo được luật cảnh báo, và toàn bộ chuỗi giá trị 'theo dõi giá rồi nhắc đúng lúc' (engine TASK-TRACK-004 -> push TASK-NOTIF-002) không có đầu vào từ người dùng. Nếu form cho chọn rule_type/channel ngoài enum thì server trả 400 (trải nghiệm hỏng) hoặc tệ hơn tạo luật engine không hiểu. Nếu render ô threshold sai ngữ nghĩa (vd cho nhập % cho price_below, hay cho nhập ngưỡng cho real_sale) thì người dùng tạo luật không bao giờ bắn hoặc bắn loạn. Nhập target_price dạng float gây sai số trên so sánh giá. Đây là màn hình chạm trực tiếp dữ liệu cá nhân (ý định mua sắm) nên phải trung thực về lỗi 403/404 mà không lộ tài nguyên người khác."
---

## §1 - Mô tả (BCP-14 normative)

UI quản lý wishlist và alert **MUST** là các màn hình trong khu vực đăng nhập, tiêu thụ API REST của TASK-TRACK-002 (wishlist) và TASK-TRACK-003 (alert_rule) qua `lib/api.ts`, phản chiếu đúng ngữ nghĩa `rule_type`/`threshold`/`channel` của backend, và chỉ thao tác trên tài nguyên của chính user. Hợp đồng:

1. UI **MUST** lấy và ghi dữ liệu qua API TASK-TRACK-002/003 (`/v1/wishlists...`, `/v1/alerts...`) qua `apiFetch` (TASK-WEB-001); server là nguồn sự thật, client KHÔNG tự giữ danh sách như nguồn chính (DEC-WEB-16).
2. Màn hình wishlist **MUST** cho: tạo wishlist (`POST /v1/wishlists {name}`), liệt kê wishlist của caller (`GET /v1/wishlists`), thêm item (`POST /v1/wishlists/{id}/items {product_id, target_price?}`), xóa item, xóa wishlist - khớp các route TASK-TRACK-002.
3. `target_price` **MUST** nhập và hiển thị là số nguyên VND (int64), nullable (để trống = "chưa đặt giá mong muốn") (DEC-WEB-19); gửi nguyên số đồng VND lên API, KHÔNG dùng dấu phẩy động.
4. Màn hình alert **MUST** cho tạo luật (`POST /v1/alerts {product_id, rule_type, threshold?, channel[]?}`), liệt kê luật của caller (`GET /v1/alerts`), bật/tắt (`PATCH /v1/alerts/{id} {active}`), xóa (`DELETE /v1/alerts/{id}`) - khớp các route TASK-TRACK-003.
5. Bộ chọn `rule_type` **MUST** chỉ phơi enum `{price_below, drop_pct, real_sale, bottom_predicted}` của TASK-TRACK-003 (DEC-WEB-17); không cho giá trị ngoài enum.
6. Ô nhập `threshold` **MUST** đổi theo `rule_type` (DEC-WEB-17, mirror DEC-TRACK-22):
- `price_below`: hiện ô giá VND (int > 0).
- `drop_pct`: hiện ô phần trăm nguyên trong `[1, 99]`.
- `real_sale` và `bottom_predicted`: KHÔNG hiện ô threshold (tín hiệu từ engine, không có ngưỡng người dùng).
7. UI **MUST** validate phía client quan hệ `rule_type` <-> `threshold` trước khi gửi (báo lỗi sớm), nhưng KHÔNG coi đó là kiểm tra cuối: server (TASK-TRACK-003 #5) vẫn validate lần cuối; UI hiển thị lỗi `400` từ server nếu có.
8. Bộ chọn `channel` **MUST** chỉ phơi `{push, email, sms}` với `push` mặc định (DEC-WEB-18), khớp enum kênh TASK-TRACK-003; cho chọn nhiều kênh; KHÔNG cho kênh ngoài tập.
9. UI **MUST** chỉ thao tác trên tài nguyên của chính user (DEC-WEB-20): server đã chặn IDOR (TASK-TRACK-002 #6, TASK-TRACK-003), client không cố duyệt id người khác; lỗi `403`/`404` từ server hiển thị "không tìm thấy" trung lập, không lộ sự tồn tại tài nguyên người khác.
10. Các màn hình **MUST** nằm trong route group `(app)` (sau guard auth TASK-WEB-001); request mang JWT do `apiFetch` gắn.
11. UI **MUST** xử lý trạng thái rỗng (chưa có wishlist/alert) bằng hướng dẫn tạo mới, và trạng thái lỗi mạng bằng thông báo rõ; không vỡ màn hình.
12. Toàn bộ **MUST** vượt `npx tsc --noEmit` sạch và `npm test` xanh.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao server là nguồn sự thật (DEC-WEB-16)?** Wishlist và alert là dữ liệu bền thuộc người dùng, đồng bộ giữa web/extension/mobile. Client giữ bản sao như nguồn chính dẫn tới lệch trạng thái (một thiết bị sửa, thiết bị khác cũ). UI đọc/ghi qua API và phản ánh phản hồi server; cache phía client chỉ để hiển thị, làm tươi từ server.

**Vì sao form phản chiếu đúng ngữ nghĩa threshold (DEC-WEB-17)?** Backend TASK-TRACK-003 diễn giải `threshold` khác nhau theo `rule_type`: VND cho price_below, % cho drop_pct, không dùng cho real_sale/bottom_predicted. Nếu form cho nhập sai (vd % cho price_below) thì hoặc server trả 400 (trải nghiệm xấu), hoặc tạo luật vô nghĩa không bao giờ bắn. Ô threshold động theo loại giữ form khớp đúng ngữ nghĩa server.

**Vì sao validate client nhưng không tin client (§1 #7)?** Validate phía client cho phản hồi tức thì (người dùng thấy lỗi ngay khi gõ), tốt cho UX. Nhưng client có thể bị bỏ qua (gọi API trực tiếp), nên server vẫn là kiểm tra cuối. UI báo sớm + tôn trọng lỗi server là defense in depth, không thay thế.

**Vì sao channel chỉ allowlist {push,email,sms} (DEC-WEB-18)?** Mỗi kênh cần một dispatcher (FCM, email, SMS). Cho chọn kênh không có dispatcher tạo luật gửi vào hư không. Allowlist khớp đúng enum TASK-TRACK-003; push mặc định khớp mô hình chi phí push > email > sms (§3.6).

**Vì sao target_price/threshold là int VND (DEC-WEB-19)?** Tiền VN là số nguyên đồng. Nhập/lưu float gây sai số khi engine so `giá hiện tại <= target_price`. Giữ int64 suốt từ ô nhập tới API đồng nhất DEC-PRICE-05, tránh sai số.

**Vì sao lỗi 403/404 trung lập (DEC-WEB-20)?** Nếu user đoán id wishlist người khác, server chặn (IDOR). UI không nên phân biệt "không tồn tại" với "tồn tại nhưng không phải của bạn" - cả hai hiển thị "không tìm thấy" để không rò rỉ sự tồn tại tài nguyên người khác (quyền riêng tư, đụng PDPL).

---

## §3 - Hợp đồng API / DDL

### Client wishlist (lib/wishlist/api.ts)

```ts
// web/lib/wishlist/api.ts
import { apiFetch } from "@/lib/api";

export interface Wishlist { id: number; name: string; itemCount: number }
export interface WishlistItem { id: number; productId: number; targetPrice: number | null } // VND int64

export async function listWishlists(): Promise<Wishlist[]> {
  const res = await apiFetch("/v1/wishlists");           // GET của TASK-TRACK-002
  return (await res.json()) as Wishlist[];
}
export async function addItem(wishlistId: number, productId: number, targetPrice: number | null) {
  return apiFetch(`/v1/wishlists/${wishlistId}/items`, {
    method: "POST", headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ product_id: productId, target_price: targetPrice }), // int VND (DEC-WEB-19)
  });
}
```

### Validate alert theo rule_type (lib/alerts/validate.ts)

```ts
// web/lib/alerts/validate.ts - mirror DEC-TRACK-22 của TASK-TRACK-003
export type RuleType = "price_below" | "drop_pct" | "real_sale" | "bottom_predicted";
export type Channel = "push" | "email" | "sms";
export const CHANNELS: Channel[] = ["push", "email", "sms"];          // khớp enum (DEC-WEB-18)

export function needsThreshold(rt: RuleType): boolean {
  return rt === "price_below" || rt === "drop_pct";                   // còn lại KHÔNG có threshold
}

export function validateAlert(rt: RuleType, threshold: number | null, channels: Channel[]): string | null {
  if (channels.length === 0) return "Chọn ít nhất một kênh";
  if (channels.some((c) => !CHANNELS.includes(c))) return "Kênh không hợp lệ";
  if (rt === "price_below") {
    if (threshold == null || !Number.isInteger(threshold) || threshold <= 0) return "Giá phải là số nguyên dương (VND)";
  } else if (rt === "drop_pct") {
    if (threshold == null || !Number.isInteger(threshold) || threshold < 1 || threshold > 99) return "Phần trăm phải trong 1..99";
  } else {
    if (threshold != null) return "Loại này không nhận ngưỡng"; // real_sale / bottom_predicted
  }
  return null; // hợp lệ
}
```

### Form alert (ô threshold động theo loại)

```tsx
// web/components/alerts/alert-form.tsx (đoạn lõi)
const [ruleType, setRuleType] = useState<RuleType>("price_below");
const [threshold, setThreshold] = useState<number | null>(null);
const [channels, setChannels] = useState<Channel[]>(["push"]);

const err = validateAlert(ruleType, threshold, channels); // báo sớm (DEC-WEB-17), server vẫn validate

return (
  <form onSubmit={submit}>
    <select value={ruleType} onChange={(e) => setRuleType(e.target.value as RuleType)}>
      <option value="price_below">Báo khi về giá</option>
      <option value="drop_pct">Báo khi giảm %</option>
      <option value="real_sale">Báo khi sale thật</option>
      <option value="bottom_predicted">Báo khi sắp chạm đáy</option>
    </select>
    {needsThreshold(ruleType) && (
      <input type="number" step={1} value={threshold ?? ""}    // int VND hoặc % (DEC-WEB-19)
        onChange={(e) => setThreshold(e.target.value === "" ? null : parseInt(e.target.value, 10))} />
    )}
    {CHANNELS.map((c) => (/* checkbox push/email/sms */ null))}
    <button disabled={err != null}>Tạo cảnh báo</button>
  </form>
);
```

---

## §4 - Acceptance criteria

1. `lib/wishlist/api.ts` và `lib/alerts/api.ts` gọi đúng route TASK-TRACK-002/003 qua `apiFetch`; grep: client không giữ danh sách như nguồn sự thật (đọc từ API).
2. Tạo/liệt kê/thêm item/xóa wishlist khớp các route TASK-TRACK-002; `target_price` gửi là số nguyên hoặc null.
3. `target_price` nhập là số nguyên VND; grep: không có parse float cho tiền; để trống gửi `null`.
4. Tạo/liệt kê/bật-tắt/xóa alert khớp các route TASK-TRACK-003.
5. Bộ chọn `rule_type` chỉ có 4 giá trị enum; không render lựa chọn ngoài enum.
6. Ô `threshold` hiện cho `price_below` (giá) và `drop_pct` (%); ẩn cho `real_sale`/`bottom_predicted`.
7. `validateAlert` trả lỗi đúng: price_below cần int > 0; drop_pct cần 1..99; real_sale/bottom_predicted không nhận threshold; channel rỗng/lạ báo lỗi.
8. Bộ chọn `channel` chỉ `{push,email,sms}`, push mặc định; cho chọn nhiều; không kênh ngoài tập.
9. Lỗi `403`/`404` từ server hiển thị "không tìm thấy" trung lập; client không phân biệt tồn-tại-nhưng-không-phải-của-bạn.
10. Màn hình nằm trong `(app)` (sau guard auth); request mang JWT qua `apiFetch`.
11. Trạng thái rỗng có hướng dẫn tạo mới; lỗi mạng có thông báo; không vỡ.
12. `npx tsc --noEmit` sạch; `npm test` xanh.

---

## §5 - Kiểm thử (verification)

```ts
// web/test/alert-form-validate.test.ts
import { validateAlert } from "../lib/alerts/validate";

test("price_below cần số nguyên dương VND", () => {
  expect(validateAlert("price_below", 0, ["push"])).toMatch(/số nguyên dương/);
  expect(validateAlert("price_below", 89000, ["push"])).toBeNull();
});

test("drop_pct cần 1..99", () => {
  expect(validateAlert("drop_pct", 0, ["push"])).toMatch(/1\.\.99/);
  expect(validateAlert("drop_pct", 150, ["push"])).toMatch(/1\.\.99/);
  expect(validateAlert("drop_pct", 30, ["push"])).toBeNull();
});

test("real_sale / bottom_predicted KHÔNG nhận threshold", () => {
  expect(validateAlert("real_sale", 100, ["push"])).toMatch(/không nhận ngưỡng/);
  expect(validateAlert("bottom_predicted", null, ["push"])).toBeNull();
});
```

```ts
// web/test/alert-channel.test.ts
import { validateAlert, CHANNELS } from "../lib/alerts/validate";

test("channel chỉ {push,email,sms}; rỗng hoặc lạ báo lỗi", () => {
  expect(CHANNELS).toEqual(["push", "email", "sms"]);
  expect(validateAlert("real_sale", null, [])).toMatch(/ít nhất một kênh/);
  expect(validateAlert("real_sale", null, ["zalo" as any])).toMatch(/không hợp lệ/);
  expect(validateAlert("real_sale", null, ["push", "email"])).toBeNull();
});
```

```ts
// web/test/wishlist-api.test.ts
import { addItem } from "../lib/wishlist/api";

test("addItem gửi target_price là số nguyên VND (không float)", async () => {
  const spy = mockFetchOk();
  await addItem(3, 90112, 99000);
  const body = JSON.parse(spy.mock.calls[0][1].body);
  expect(body).toEqual({ product_id: 90112, target_price: 99000 });
  expect(Number.isInteger(body.target_price)).toBe(true);
});

test("addItem để trống target_price gửi null", async () => {
  const spy = mockFetchOk();
  await addItem(3, 90112, null);
  expect(JSON.parse(spy.mock.calls[0][1].body).target_price).toBeNull();
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `lib/wishlist/api.ts` + `lib/alerts/api.ts` (client gọi đúng route TASK-TRACK-002/003 qua apiFetch) -> `lib/alerts/validate.ts` (mirror ngữ nghĩa threshold/channel, test thuần) -> `wishlist-panel.tsx` + `wishlist-item-row.tsx` (CRUD wishlist + sửa target_price) -> `alert-form.tsx` (rule_type + ô threshold động + channel) + `alert-list.tsx` (bật/tắt/xóa) -> `wishlist/page.tsx` + `alerts/page.tsx` (màn hình trong shell (app)) -> thêm điều hướng vào `app-shell.tsx` -> tests. Các form là client component (`"use client"`) cho tương tác; trang gọi API phía client sau khi shell (app) qua guard auth. Validate phía client báo sớm; server (TASK-TRACK-003) vẫn là kiểm tra cuối.

---

## §7 - Phụ thuộc

- **TASK-TRACK-002** - schema + API wishlist/wishlist_item; UI này là mặt tiền của các route `/v1/wishlists...` (depends_on cứng).
- **TASK-TRACK-003** - schema + API alert_rule; UI phản chiếu đúng enum `rule_type`/`channel` và ngữ nghĩa `threshold` của nó (depends_on cứng).
- **TASK-WEB-001** - scaffold + `lib/api.ts` (JWT) + shell `(app)` + guard; màn hình nằm trong khu vực bảo vệ (depends_on cứng).
- **TASK-TRACK-004 (downstream)** - engine kích hoạt đọc các luật người dùng tạo qua UI này; UI là nơi luật được sinh ra.
- Lib: `next` 14, React 18, `@testing-library/react`.

---

## §8 - Payload ví dụ

### Tạo luật price_below (push + email)

```ts
await createAlert({
  product_id: 90112,
  rule_type: "price_below",
  threshold: 89000,            // VND int64 (DEC-WEB-19)
  channel: ["push", "email"],  // allowlist (DEC-WEB-18)
});
// POST /v1/alerts -> 201 (TASK-TRACK-003)
```

### Tạo luật real_sale (không threshold)

```json
{ "product_id": 90112, "rule_type": "real_sale", "channel": ["push"] }
```

### Thêm item vào wishlist với giá mong muốn

```json
{ "product_id": 77310, "target_price": 199000 }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Kéo-thả sắp xếp item giữa các wishlist - tinh chỉnh UX sau, không đổi API.
- Gợi ý target_price dựa median90/trailing_min (đọc feed TASK-DEAL-003) khi đặt giá mong muốn - thêm khi tích hợp biểu đồ.
- Tạo nhanh alert ngay từ màn hình biểu đồ (TASK-WEB-003) - liên kết hai màn hình ở slice sau.
- Hiển thị lịch sử bắn (alert history) inline - dùng `GET /v1/alerts/{id}/history` của TASK-TRACK-003 khi cần.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Chọn rule_type/channel ngoài enum | alert-form/channel test | server 400 hoặc luật rác | Chỉ phơi enum TASK-TRACK-003 (DEC-WEB-17/18) |
| Ô threshold sai ngữ nghĩa | alert-form-validate test | luật không bắn / bắn loạn | Ô động theo rule_type (DEC-WEB-17) |
| target_price/threshold float | wishlist-api test | sai số so giá | Int VND (DEC-WEB-19) |
| Client giữ danh sách lệch server | review | trạng thái cũ | Server là nguồn sự thật (DEC-WEB-16) |
| Lộ tài nguyên user khác qua 403/404 | review thông báo | rò rỉ quyền riêng tư | "Không tìm thấy" trung lập (DEC-WEB-20) |
| Màn hình lộ khi chưa đăng nhập | guard (app) | rò rỉ | Nằm trong (app), JWT qua apiFetch (§1 #10) |
| Channel rỗng gửi lên | validate test | luật không gửi đâu | Bắt buộc >=1 kênh (validateAlert) |
| Trạng thái rỗng làm vỡ | empty-state | UX kém | Hướng dẫn tạo mới (§1 #11) |
| Bỏ qua validate client (gọi API thẳng) | server validate | luật xấu | Server TASK-TRACK-003 validate cuối (§1 #7) |

---

## §11 - Ghi chú

- wishlist + alert biến SănDeal từ công cụ tra cứu thành trợ lý chủ động; đây là mặt tiền của backend TASK-TRACK-002/003.
- Server là nguồn sự thật; UI đọc/ghi qua API và đồng bộ giữa web/extension/mobile, không giữ bản sao làm nguồn chính.
- Form alert phản chiếu đúng ngữ nghĩa threshold theo rule_type của TASK-TRACK-003 - ô động tránh tạo luật vô nghĩa.
- Validate phía client báo sớm cho UX, nhưng server vẫn là kiểm tra cuối (defense in depth).
- channel allowlist {push,email,sms} khớp dispatcher có thật; push mặc định khớp mô hình chi phí push > email > sms.
- Lỗi 403/404 trung lập bảo vệ quyền riêng tư - không lộ sự tồn tại tài nguyên người khác, đụng đúng yêu cầu PDPL.

---

*Hết TASK-WEB-004. Status: ready_to_implement (mục tiêu audit 10/10).*
