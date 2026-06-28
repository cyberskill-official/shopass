---
id: FR-WEB-003
title: "UI biểu đồ lịch sử giá - render p95 <500ms, tiêu thụ feed FR-DEAL-003 (price_daily + overlay median90/trailing_min/verdict sale ảo/mốc ngày đôi), trả lời câu hỏi lõi sale thật hay ảo"
module: WEB
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-WEB-001, FR-DEAL-003, FR-DEAL-001, FR-PRICE-002]
depends_on: [FR-WEB-001, FR-DEAL-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.8 (NFR biểu đồ giá <500ms)"
  - "docs/... §3.7 (feed biểu đồ), §3.5(1) (verdict sale ảo SALE_AO/SALE_XIN/TAM_DUOC), §5.1 (cold-start: SKU mới chưa đủ dữ liệu)"
source_decisions:
  - "DEC-WEB-11: biểu đồ tiêu thụ DUY NHẤT feed FR-DEAL-003 (GET /v1/products/{id}/chart) - không tự gọi raw price_snapshot, không tự tính verdict/median; client chỉ vẽ"
  - "DEC-WEB-12: render p95 <500ms tính từ lúc có dữ liệu tới lúc vẽ xong - feed đã downsample theo ngày nên payload nhỏ; không tải toàn chuỗi raw"
  - "DEC-WEB-13: hiển thị nhãn verdict (SALE_AO/SALE_XIN/TAM_DUOC) đồng nhất với thẻ sản phẩm - lấy thẳng từ annotations.verdict của feed, KHÔNG suy diễn lại ở client"
  - "DEC-WEB-14: tôn trọng maturity từ feed - NEW (<14 ngày) hiển thị 'đang thu thập dữ liệu, chưa kết luận' (verdict UNKNOWN); WARMING gắn dải cờ đang tích lũy; không vẽ kết luận trên ít dữ liệu"
  - "DEC-WEB-15: bộ chọn range chỉ phơi các nút trong allowlist của feed (7d/30d/90d/180d/1y) để không gửi range lạ làm backend trả 400"

language: "TypeScript 5.x; Next.js 14 (App Router, client component cho chart); thư viện chart (Recharts/visx); gọi feed FR-DEAL-003 qua lib/api.ts (FR-WEB-001)"
service: shopass/web/
new_files:
  - web/app/(app)/products/[id]/chart/page.tsx
  - web/components/price-chart/price-chart.tsx
  - web/components/price-chart/verdict-badge.tsx
  - web/components/price-chart/range-selector.tsx
  - web/components/price-chart/maturity-notice.tsx
  - web/lib/chart/fetch-chart.ts
  - web/lib/chart/types.ts
  - web/test/price-chart.test.tsx
  - web/test/verdict-badge.test.tsx
  - web/test/range-selector.test.tsx
modified_files:
  - web/lib/api.ts                           # tái dùng apiFetch cho endpoint chart
allowed_tools:
  - file_read: web/**
  - file_write: web/**
  - bash: cd web && npm test && npx tsc --noEmit
disallowed_tools:
  - gọi raw price_snapshot / tự dựng biểu đồ từ dữ liệu thô thay vì dùng feed FR-DEAL-003 (vi phạm DEC-WEB-11, vỡ p95 <500ms)
  - tự tính lại verdict sale ảo ở client (vi phạm DEC-WEB-13, lệch nhãn với thẻ sản phẩm)
  - hiển thị verdict cho SKU NEW <14 ngày (vi phạm DEC-WEB-14, kết luận ẩu trên ít dữ liệu)
  - phơi nút range ngoài allowlist của feed (vi phạm DEC-WEB-15, backend trả 400)

effort_hours: 6
sub_tasks:
  - "0.5h: lib/chart/types.ts - mirror DTO feed FR-DEAL-003 (daily[], annotations{median90,trailing_min,verdict,accumulating,double_dates}, maturity)"
  - "1.0h: lib/chart/fetch-chart.ts - gọi GET /v1/products/{id}/chart?range= qua apiFetch; validate range qua allowlist phía client"
  - "1.5h: price-chart.tsx - vẽ thân daily (close_p) + đường tham chiếu median90/trailing_min + mốc ngày đôi"
  - "0.75h: verdict-badge.tsx - nhãn SALE_AO/SALE_XIN/TAM_DUOC/UNKNOWN từ annotations.verdict (không suy diễn)"
  - "0.5h: range-selector.tsx - nút 7d/30d/90d/180d/1y (allowlist feed), đổi range refetch"
  - "0.5h: maturity-notice.tsx - thông báo NEW (chưa kết luận) / WARMING (đang tích lũy)"
  - "0.25h: products/[id]/chart/page.tsx - client component ghép, nằm trong shell (app)"
  - "1.0h: tests - vẽ overlay đúng, badge khớp verdict feed, NEW ẩn verdict, range ngoài allowlist không gửi"

risk_if_skipped: "Đây là màn hình mà người dùng SănDeal mở để trả lời câu hỏi cốt lõi: đây có phải sale thật không. Biểu đồ lịch sử giá kèm dải verdict sale ảo, đường trailing_min và mốc ngày đôi là minh chứng trực quan cho lời hứa sản phẩm - thiếu nó thì người dùng quay lại đoán mò như trước. Nếu dựng sai bằng cách gọi raw price_snapshot và tự tính thì p95 vỡ NFR <500ms (§3.8), biểu đồ giật trên SKU lịch sử dài, và tệ hơn là verdict ở biểu đồ lệch với nhãn trên thẻ sản phẩm (client tự suy diễn) khiến người dùng mất tin - đúng thứ niềm tin mà SănDeal phải bảo vệ hậu-Honey. Nếu vẽ kết luận trên SKU mới <14 ngày thì dán nhãn dựa trên vài điểm dữ liệu, đúng kiểu sale ảo ẩu mà sản phẩm hứa loại bỏ. Đây cũng là CTA đích của landing SEO (FR-WEB-002) - hỏng nó là hỏng đầu ra funnel."
---

## §1 - Mô tả (BCP-14 normative)

UI biểu đồ lịch sử giá **MUST** là một màn hình trong khu vực đăng nhập, tiêu thụ DUY NHẤT feed của FR-DEAL-003, vẽ thân giá theo ngày kèm các chú giải tín hiệu deal, đạt render p95 <500ms, và tôn trọng độ chín dữ liệu để không kết luận ẩu. Hợp đồng:

1. UI **MUST** lấy dữ liệu DUY NHẤT từ feed `GET /v1/products/{id}/chart?range=...` của FR-DEAL-003 qua `lib/api.ts` (FR-WEB-001); KHÔNG tự gọi raw `price_snapshot`, KHÔNG tự dựng chuỗi giá từ dữ liệu thô (DEC-WEB-11).
2. Biểu đồ **MUST** vẽ phần thân từ mảng `daily` của feed (`{day, min_p, max_p, close_p}`), dùng `close_p` làm đường giá hiển thị, sắp tăng theo `day`.
3. Biểu đồ **MUST** vẽ các đường tham chiếu từ `annotations`: `median90` (đường trung vị 90 ngày) và `trailing_min` (đường đáy giá trong khoảng) - lấy thẳng từ feed, KHÔNG tự tính (DEC-WEB-11).
4. Biểu đồ **MUST** đánh dấu các mốc `annotations.double_dates` (ngày đôi 1.1...12.12 trong khoảng) trên trục thời gian, giúp người dùng đọc đúng bối cảnh lịch sale VN.
5. UI **MUST** hiển thị nhãn verdict lấy thẳng từ `annotations.verdict` (`SALE_AO`, `SALE_XIN`, `TAM_DUOC`, hoặc `UNKNOWN`) - KHÔNG suy diễn lại ở client (DEC-WEB-13). Nhãn này phải đồng nhất với nhãn trên thẻ sản phẩm (cùng nguồn FR-DEAL-001 qua feed).
6. UI **MUST** tôn trọng `maturity` của feed (DEC-WEB-14): khi `maturity == "NEW"` (<14 ngày) hiển thị thông báo "đang thu thập dữ liệu, chưa đủ để kết luận" và KHÔNG hiển thị badge verdict kết luận (verdict là `UNKNOWN`); khi `maturity == "WARMING"` (hoặc `annotations.accumulating == true`) hiển thị ghi chú "đang tích lũy dữ liệu" kèm biểu đồ.
7. Bộ chọn khoảng (range selector) **MUST** chỉ phơi các nút trong allowlist của feed `{7d, 30d, 90d, 180d, 1y}` (DEC-WEB-15); đổi range gọi lại feed với range mới. KHÔNG gửi range ngoài allowlist.
8. Render **MUST** đạt p95 <500ms tính từ khi nhận dữ liệu feed tới khi vẽ xong (DEC-WEB-12): feed đã downsample theo ngày nên payload nhỏ; không tải toàn chuỗi raw, không tính toán nặng ở client.
9. UI **MUST** xử lý trạng thái rỗng lịch sự: SKU có thật nhưng `daily` rỗng (chưa có snapshot) hiển thị "đang thu thập dữ liệu giá", không vỡ; lỗi mạng/`404` hiển thị thông báo lỗi rõ ràng.
10. Màn hình **MUST** nằm trong route group `(app)` (sau guard auth của FR-WEB-001); request feed mang JWT do `lib/api.ts` gắn.
11. Mọi giá hiển thị **MUST** được hiểu là số nguyên VND (int64 từ feed) - định dạng tiền tệ vi-VN khi render, KHÔNG dùng phép tính dấu phẩy động làm sai số.
12. Toàn bộ **MUST** vượt `npx tsc --noEmit` sạch và `npm test` xanh.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao chỉ tiêu thụ feed FR-DEAL-003 (DEC-WEB-11)?** Feed đã làm sẵn ba việc nặng ở server: downsample về bucket ngày (payload nhỏ), tính median90/trailing_min, và gắn verdict từ một nguồn (FR-DEAL-001). Client tự gọi raw rồi tính lại vừa chậm (p95 vỡ), vừa lệch (verdict khác thẻ sản phẩm). Client chỉ nên vẽ. Đây là ranh giới phân tách rõ giữa tính toán (server) và hiển thị (client).

**Vì sao p95 <500ms (DEC-WEB-12)?** §3.8 đặt ngưỡng biểu đồ <500ms. Đây là màn hình người dùng mở để quyết "mua hay đợi" - chậm là mất khoảnh khắc quyết định. Vì feed đã downsample, payload nhỏ và client chỉ vẽ, ngưỡng này đạt được mà không cần tối ưu phức tạp. Nếu lỡ tải raw thì payload phình và render giật.

**Vì sao verdict không suy diễn ở client (DEC-WEB-13)?** Nhãn sale ảo phải đồng nhất giữa thẻ sản phẩm và biểu đồ. Nếu thẻ ghi SALE_XIN mà biểu đồ tự tính ra SALE_AO thì người dùng mất tin ngay - và niềm tin là moat của SănDeal hậu-Honey. Lấy thẳng `annotations.verdict` đảm bảo một nhãn từ một nguồn.

**Vì sao tôn trọng maturity (DEC-WEB-14)?** Vẽ kết luận trên SKU mới <14 ngày là dán nhãn dựa trên vài điểm dữ liệu - đúng kiểu kết luận ẩu mà SănDeal hứa loại bỏ (§5.1 cold-start). NEW thì nói thẳng "chưa đủ dữ liệu"; WARMING thì vẽ nhưng gắn cờ đang tích lũy. Trung thực về độ chín là một phần của lời hứa niềm tin.

**Vì sao range selector chỉ allowlist (DEC-WEB-15)?** Feed chỉ nhận `{7d,30d,90d,180d,1y}`; range lạ trả 400. Phơi đúng các nút này ở UI tránh gửi range không hợp lệ và tránh lỗi hiển thị. UI và feed thống nhất một tập range.

**Vì sao tách verdict-badge, range-selector, maturity-notice riêng (§3)?** Mỗi mảnh là một đơn vị hiển thị độc lập, dễ test riêng (badge khớp verdict, selector phát range đúng, notice hiện đúng theo maturity). Ghép lại trong `price-chart` giữ component chính gọn.

---

## §3 - Hợp đồng API / DDL

### Kiểu dữ liệu mirror feed (lib/chart/types.ts)

```ts
// web/lib/chart/types.ts - mirror DTO của FR-DEAL-003
export interface DailyPoint { day: string; min_p: number; max_p: number; close_p: number } // VND int64
export type Verdict = "SALE_AO" | "SALE_XIN" | "TAM_DUOC" | "UNKNOWN";
export type Maturity = "MATURE" | "WARMING" | "NEW";

export interface Annotations {
  median90: number; trailing_min: number;       // VND
  verdict: Verdict; accumulating: boolean; double_dates: string[];
}
export interface ChartResponse {
  product_id: number; range: string; maturity: Maturity;
  daily: DailyPoint[]; annotations: Annotations;
}

export const RANGE_ALLOWLIST = ["7d", "30d", "90d", "180d", "1y"] as const; // khớp feed (DEC-WEB-15)
export type Range = typeof RANGE_ALLOWLIST[number];
```

### Lấy dữ liệu (lib/chart/fetch-chart.ts)

```ts
// web/lib/chart/fetch-chart.ts
import { apiFetch } from "@/lib/api";
import { RANGE_ALLOWLIST, type ChartResponse, type Range } from "./types";

export async function fetchChart(productId: number, range: Range): Promise<ChartResponse> {
  if (!RANGE_ALLOWLIST.includes(range)) throw new Error("range ngoài allowlist"); // DEC-WEB-15
  const res = await apiFetch(`/v1/products/${productId}/chart?range=${range}`); // feed FR-DEAL-003
  if (res.status === 404) throw new NotFoundError();
  if (!res.ok) throw new Error("không tải được biểu đồ");
  return (await res.json()) as ChartResponse; // client KHÔNG tính lại verdict/median (DEC-WEB-11/13)
}
```

### Badge verdict (verdict-badge.tsx) - hiển thị thẳng, không suy diễn

```tsx
// web/components/price-chart/verdict-badge.tsx
import type { Verdict, Maturity } from "@/lib/chart/types";

const LABEL: Record<Verdict, string> = {
  SALE_AO: "Sale ảo", SALE_XIN: "Sale xịn", TAM_DUOC: "Tạm được", UNKNOWN: "Chưa đủ dữ liệu",
};

export function VerdictBadge({ verdict, maturity }: { verdict: Verdict; maturity: Maturity }) {
  if (maturity === "NEW") return null;             // <14 ngày: KHÔNG badge kết luận (DEC-WEB-14)
  return <span data-verdict={verdict}>{LABEL[verdict]}</span>; // lấy thẳng từ feed (DEC-WEB-13)
}
```

---

## §4 - Acceptance criteria

1. `fetchChart` gọi `GET /v1/products/{id}/chart?range=` qua `apiFetch`; grep `web/components/price-chart/**` + `lib/chart/**`: KHÔNG có gọi raw `price_snapshot`/`price-history`.
2. Biểu đồ vẽ thân từ `daily` (close_p) sắp tăng theo `day`.
3. Biểu đồ vẽ đường `median90` và `trailing_min` lấy từ `annotations` (không tự tính trong client).
4. Mốc `annotations.double_dates` được đánh dấu trên trục thời gian.
5. `VerdictBadge` hiển thị nhãn từ `annotations.verdict`; với cùng verdict, nhãn khớp nhãn thẻ sản phẩm (cùng nguồn); grep: client không có hàm tính verdict.
6. `maturity == "NEW"` -> ẩn badge verdict + hiện "đang thu thập dữ liệu, chưa kết luận"; `WARMING`/`accumulating` -> hiện ghi chú đang tích lũy kèm biểu đồ.
7. Range selector chỉ render nút trong `{7d,30d,90d,180d,1y}`; `fetchChart` ném lỗi nếu nhận range ngoài allowlist (không gọi mạng).
8. SKU có thật, `daily` rỗng -> hiện "đang thu thập dữ liệu giá", không vỡ; `404`/lỗi mạng -> thông báo lỗi rõ.
9. Màn hình nằm trong group `(app)` (sau guard auth); request feed mang JWT qua `apiFetch`.
10. Giá hiển thị định dạng tiền VND (vi-VN) từ số nguyên int64; không phép dấu phẩy động làm sai số.
11. `npx tsc --noEmit` sạch; `npm test` xanh.

---

## §5 - Kiểm thử (verification)

```tsx
// web/test/price-chart.test.tsx
import { render, screen } from "@testing-library/react";
import { PriceChart } from "../components/price-chart/price-chart";

const feed = {
  product_id: 90112, range: "90d", maturity: "MATURE",
  daily: [
    { day: "2026-04-04", min_p: 119000, max_p: 149000, close_p: 119000 },
    { day: "2026-05-05", min_p: 109000, max_p: 139000, close_p: 109000 },
  ],
  annotations: { median90: 129000, trailing_min: 99000, verdict: "TAM_DUOC",
                 accumulating: false, double_dates: ["2026-04-04", "2026-05-05"] },
};

test("vẽ đường tham chiếu median90 + trailing_min từ feed (không tự tính)", () => {
  render(<PriceChart data={feed as any} />);
  expect(screen.getByTestId("ref-median90")).toHaveAttribute("data-value", "129000");
  expect(screen.getByTestId("ref-trailing-min")).toHaveAttribute("data-value", "99000");
});

test("đánh dấu mốc ngày đôi từ feed", () => {
  render(<PriceChart data={feed as any} />);
  expect(screen.getAllByTestId("double-date-marker").length).toBe(2);
});
```

```tsx
// web/test/verdict-badge.test.tsx
import { render, screen } from "@testing-library/react";
import { VerdictBadge } from "../components/price-chart/verdict-badge";

test("badge lấy thẳng verdict từ feed", () => {
  render(<VerdictBadge verdict="SALE_AO" maturity="MATURE" />);
  expect(screen.getByText("Sale ảo")).toHaveAttribute("data-verdict", "SALE_AO");
});

test("SKU NEW (<14 ngày) KHÔNG hiện badge kết luận", () => {
  const { container } = render(<VerdictBadge verdict="UNKNOWN" maturity="NEW" />);
  expect(container).toBeEmptyDOMElement();
});
```

```tsx
// web/test/range-selector.test.tsx
import { fetchChart } from "../lib/chart/fetch-chart";
import { RANGE_ALLOWLIST } from "../lib/chart/types";

test("chỉ allowlist range; range lạ ném lỗi, không gọi mạng", async () => {
  const spy = mockFetchOk();
  await expect(fetchChart(1, "5d" as any)).rejects.toThrow(/allowlist/);
  expect(spy).not.toHaveBeenCalled();
  expect(RANGE_ALLOWLIST).toEqual(["7d", "30d", "90d", "180d", "1y"]);
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `lib/chart/types.ts` (mirror DTO feed, allowlist range) -> `lib/chart/fetch-chart.ts` (gọi feed qua apiFetch, validate range) -> `verdict-badge.tsx` + `range-selector.tsx` + `maturity-notice.tsx` (mảnh hiển thị nhỏ, test riêng) -> `price-chart.tsx` (vẽ thân daily + đường tham chiếu + mốc ngày đôi, ghép các mảnh) -> `products/[id]/chart/page.tsx` (client component trong shell (app)) -> tests. Chart là client component (`"use client"`) vì cần tương tác range; trang gọi `fetchChart` phía client sau khi shell (app) qua guard auth (FR-WEB-001). Dùng thư viện chart nhẹ (Recharts/visx); không tự viết engine vẽ.

---

## §7 - Phụ thuộc

- **FR-DEAL-003** - feed `GET /v1/products/{id}/chart` là nguồn dữ liệu DUY NHẤT (daily + annotations + maturity). Hình dạng JSON của nó là hợp đồng với UI này (depends_on cứng).
- **FR-WEB-001** - scaffold + `lib/api.ts` (gắn JWT) + shell `(app)` + guard auth; màn hình chart nằm trong khu vực bảo vệ (depends_on cứng).
- **FR-DEAL-001 (qua feed)** - nguồn verdict sale ảo; UI hiển thị `annotations.verdict`, không tính lại - đảm bảo nhãn khớp thẻ sản phẩm.
- **FR-PRICE-002 (gián tiếp)** - `price_daily` nuôi feed; UI không chạm trực tiếp.
- **FR-WEB-002 (upstream funnel)** - landing SEO dẫn CTA tới màn hình này; đây là đầu ra trải nghiệm của funnel.
- Lib: `next` 14, thư viện chart (Recharts/visx), `@testing-library/react`.

---

## §8 - Payload ví dụ

### Feed nhận từ FR-DEAL-003 (rút gọn)

```json
{
  "product_id": 90112, "range": "90d", "maturity": "MATURE",
  "daily": [
    { "day": "2026-04-04", "min_p": 119000, "max_p": 149000, "close_p": 119000 },
    { "day": "2026-05-05", "min_p": 109000, "max_p": 139000, "close_p": 109000 },
    { "day": "2026-06-26", "min_p": 99000,  "max_p": 119000, "close_p": 99000 }
  ],
  "annotations": {
    "median90": 129000, "trailing_min": 99000, "verdict": "TAM_DUOC",
    "accumulating": false, "double_dates": ["2026-04-04", "2026-05-05", "2026-06-06"]
  }
}
```

### Trạng thái SKU mới (NEW) - UI không kết luận

```json
{ "product_id": 5501, "range": "90d", "maturity": "NEW",
  "daily": [ { "day": "2026-06-24", "min_p": 50000, "max_p": 60000, "close_p": 55000 } ],
  "annotations": { "median90": 0, "trailing_min": 50000, "verdict": "UNKNOWN",
                   "accumulating": false, "double_dates": [] } }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- So sánh giá chéo 3 sàn trên cùng biểu đồ (overlay nhiều sàn) - chờ FR-PRICE-004; slice này một SKU một sàn.
- Tô màu vùng theo verdict per đoạn (band per khoảng) khi feed trả dải verdict - bám DEC-DEAL của FR-DEAL-003.
- Tương tác zoom/hover chi tiết theo điểm - thêm khi UX cần, không đổi nguồn dữ liệu.
- Hiển thị giá theo currency khác khi mở SEA - bám tracked_product, giữ int64 minor unit.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Gọi raw price_snapshot tự dựng | grep + p95 metric | vỡ <500ms, biểu đồ giật | Chỉ tiêu thụ feed FR-DEAL-003 (DEC-WEB-11) |
| Client tự tính verdict | grep hàm verdict | nhãn lệch thẻ sản phẩm | Lấy thẳng annotations.verdict (DEC-WEB-13) |
| Vẽ kết luận SKU NEW | verdict-badge test | kết luận ẩu ít dữ liệu | Ẩn badge khi NEW (DEC-WEB-14) |
| Gửi range ngoài allowlist | range-selector test | backend 400 | Allowlist + validate fetchChart (DEC-WEB-15) |
| daily rỗng làm vỡ render | empty-state test | trang lỗi | "đang thu thập dữ liệu giá" lịch sự (§1 #9) |
| 404 không xử lý | fetch test | trang treo | NotFoundError + thông báo rõ |
| Sai số tiền do float | review format | giá hiển thị lệch | Định dạng từ int64 VND (§1 #11) |
| Màn hình lộ khi chưa đăng nhập | guard (app) | rò rỉ | Nằm trong (app), JWT qua apiFetch (§1 #10) |
| Payload phình do không downsample | review feed | render chậm | Feed đã bucket ngày (DEC-WEB-12) |

---

## §11 - Ghi chú

- Đây là màn hình trả lời câu hỏi lõi "sale thật hay ảo" - minh chứng trực quan cho lời hứa sản phẩm, và là đầu ra của funnel SEO (FR-WEB-002).
- Client chỉ vẽ; mọi tính toán (downsample, median90, trailing_min, verdict) nằm ở feed FR-DEAL-003 để giữ p95 <500ms và một nguồn verdict.
- Nhãn verdict đồng nhất giữa thẻ và biểu đồ là điều kiện niềm tin - lấy thẳng từ feed, không suy diễn ở client.
- Tôn trọng maturity (NEW chưa kết luận, WARMING đang tích lũy) là phần của lời hứa không kết luận ẩu (§5.1).
- Range selector và feed thống nhất một tập allowlist để tránh range lạ trả 400.
- Giá giữ int64 VND suốt từ feed tới hiển thị, chỉ định dạng khi render - không có bước float nào.

---

*Hết FR-WEB-003. Status: ready_to_implement (mục tiêu audit 10/10).*
