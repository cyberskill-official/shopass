---
id: FR-SCRAPE-008
title: "Lazada scraping adapter - xử lý Akamai (Alibaba), residential bắt buộc, đọc DOM-render"
module: SCRAPE
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-SCRAPE-001, FR-SCRAPE-002, FR-SCRAPE-003, FR-SCRAPE-004, FR-SCRAPE-005, FR-SCRAPE-006, FR-PRICE-002]
depends_on: [FR-SCRAPE-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (Lazada (Alibaba) thường dùng Akamai; ưu tiên đọc DOM đã render)"
  - "docs/... §3.9 (Lazada Medium-High; Akamai (Alibaba); fingerprinting TLS/HTTP; CAPTCHA)"
source_decisions:
  - "DEC-SCRAPE-32: Lazada ưu tiên đọc DOM-render qua farm Playwright (FR-SCRAPE-003); API Lazada thường ký"
  - "DEC-SCRAPE-33: Akamai đọc fingerprint ở tầng TLS/HTTP2 -> bắt buộc TLS match của farm (FR-SCRAPE-003 #4) + residential"
  - "DEC-SCRAPE-34: residential bắt buộc (DiffAkamai -> enterprise của FR-SCRAPE-004); datacenter vô dụng với Akamai"
  - "DEC-SCRAPE-35: parse giá từ JSON nhúng (window state) nếu có, fallback DOM text; giá VND số nguyên"

language: "Node.js 20 + Playwright (adapter chạy trên farm); parse JSON nhúng / DOM text -> PriceSnapshot"
service: shopass/services/scrape/
new_files:
  - services/scrape/farm/src/adapters/lazada/adapter.ts
  - services/scrape/farm/src/adapters/lazada/extract.ts
  - services/scrape/farm/src/adapters/lazada/selectors.ts
  - services/scrape/farm/src/adapters/lazada/__tests__/extract.test.ts
  - services/scrape/farm/src/adapters/lazada/__tests__/adapter.test.ts
  - services/scrape/farm/src/adapters/lazada/__tests__/fixtures/pdp.html
modified_files:
  - services/scrape/internal/orchestrator/registry.go     # đăng ký LazadaAdapter theo platform_id
allowed_tools:
  - file_read: services/scrape/**
  - file_write: services/scrape/**
  - bash: cd services/scrape/farm && npm test
disallowed_tools:
  - dùng datacenter proxy cho Lazada (vi phạm DEC-SCRAPE-34, Akamai chặn chắc)
  - bỏ TLS/HTTP2 match khi qua Akamai (vi phạm DEC-SCRAPE-33, lộ client trước cả JS)
  - lưu giá float (vi phạm DEC-PRICE-05 của FR-PRICE-002)

effort_hours: 8
sub_tasks:
  - "1.5h: selectors.ts - định vị giá/list_price/stock/flash trong DOM Lazada PDP + JSON nhúng (window state)"
  - "2.0h: extract.ts - parse JSON nhúng ưu tiên, fallback DOM text -> giá VND số nguyên"
  - "2.0h: adapter.ts - render qua farm (TLS match cho Akamai), behavior, gọi extract -> PriceSnapshot"
  - "1.5h: fixtures/pdp.html + extract.test.ts - trang thật ẩn danh -> assert price/list_price/flash"
  - "0.5h: adapter.test.ts - Akamai challenge -> ChallengedError; JSON path + DOM fallback"
  - "0.5h: registry.go đăng ký + integration orchestrator (proxy enterprise, TLS match)"

risk_if_skipped: "Lazada là sàn lớn trong khu vực và là chân thứ ba của moat so sánh chéo 3 sàn (FR-PRICE-004) - thiếu nó thì so sánh giá chỉ còn 2 sàn, yếu hơn hẳn. Lazada (Alibaba) dùng Akamai (§3.2) đọc fingerprint TLS/HTTP2 ở tầng bắt tay, trước cả khi JS chạy. Datacenter proxy vô dụng với Akamai (§3.3) - dùng nó là bị chặn ngay. Bỏ TLS/HTTP2 match là lộ client Node/Go trước cả JS patch. Lưu giá float là sai số trên so sánh chéo sàn. Adapter này phải qua được Akamai bằng farm TLS-match + residential enterprise."
---

## §1 - Mô tả (BCP-14 normative)

Adapter Lazada **MUST** lấy giá qua đọc DOM-render trên farm có TLS match (qua Akamai), residential enterprise. Hợp đồng:

1. **MUST** triển khai interface adapter của FR-SCRAPE-001 và lấy giá qua farm `RenderPrice` của FR-SCRAPE-003 (DEC-SCRAPE-32) - Lazada API thường ký, ưu tiên đọc DOM đã render.
2. **MUST** dựa vào TLS/HTTP2 fingerprint match của farm (FR-SCRAPE-003 #4) để qua Akamai (DEC-SCRAPE-33): Akamai đọc JA3/JA4 + HTTP/2 SETTINGS ở tầng bắt tay, trước khi JS chạy; adapter KHÔNG được dùng client thô bỏ qua lớp này.
3. **MUST** dùng proxy residential tier `enterprise` (DEC-SCRAPE-34): job Lazada dùng `SelectTier(DiffAkamai)` của FR-SCRAPE-004 (= enterprise); datacenter bị từ chối (Akamai chặn chắc).
4. **MUST** trích giá ưu tiên từ JSON nhúng trong trang (DEC-SCRAPE-35): Lazada thường nhúng window state JSON; parse từ đó, fallback DOM text khi không có hoặc cấu trúc đổi.
5. **MUST** map về `price.PriceSnapshot` với `price`/`list_price` BIGINT VND số nguyên (đồng bộ DEC-PRICE-05 của FR-PRICE-002), `flash_sale` true khi trang báo đang flash sale.
6. **MUST** chạy hành vi giống người của farm (FR-SCRAPE-003 #6) trước khi trích.
7. **MUST** phát hiện Akamai challenge (sensor/verify page) và ném `ChallengedError` (FR-SCRAPE-003 #8) để FR-SCRAPE-005 / orchestrator xử lý, KHÔNG trả snapshot rỗng.
8. **MUST** báo outcome (success/parse_fail/challenge) cho monitor của FR-SCRAPE-006.
9. **MUST** chờ nội dung giá render xong trước khi trích (Lazada có phần render bằng JS); không đọc DOM lúc trang chưa sẵn.
10. **MUST** tối thiểu hóa: chỉ trích trường giá, không lưu trang thô (§5.4).
11. **SHOULD** phát OTel metric: `lazada_extract_source_total{source=json|dom}` (counter), `lazada_akamai_challenge_total` (counter), `lazada_render_duration_ms` (histogram).
12. **MUST** quy đổi giá bằng số nguyên (không float trung gian).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao đọc DOM-render (DEC-SCRAPE-32)?** Lazada (Alibaba) có API thường ký và đứng sau Akamai. Như TikTok, cố gọi/ký API là cuộc đua giòn. Đọc trang đã render trong trình duyệt thật của farm để Akamai thấy một phiên trình duyệt hợp lệ, ta chỉ đọc giá đã hiển thị.

**Vì sao TLS/HTTP2 match là then chốt với Akamai (DEC-SCRAPE-33)?** Đây là điểm khác biệt cốt lõi của Lazada so với Shopee. Akamai chấm điểm ở tầng bắt tay TLS (JA3/JA4) và HTTP/2 SETTINGS - trước khi một dòng JS nào chạy. Client Node/Go mặc định có chữ ký TLS khác Chrome thật, bị Akamai nhận diện ngay dù JS patch hoàn hảo. Adapter phải đi qua farm có TLS match (FR-SCRAPE-003 #4), không được dùng client thô.

**Vì sao residential enterprise bắt buộc (DEC-SCRAPE-34)?** Tài liệu nói thẳng: datacenter vô dụng với Akamai (§3.3). Akamai chặn dải IP datacenter theo lô. `SelectTier(DiffAkamai)` của FR-SCRAPE-004 trả enterprise - residential chất lượng cao nhất - vì Lazada Medium-High không qua được bằng IP rẻ.

**Vì sao ưu tiên JSON nhúng (DEC-SCRAPE-35)?** Như TikTok: window state JSON ổn định hơn cào DOM text trước A/B layout. Lazada thường có state nhúng để hydrate; đọc từ đó trước, fallback DOM giữ độ phủ.

**Vì sao chờ render (§1 #9)?** Lazada PDP có phần giá render bằng JS. Đọc DOM lúc trang chưa sẵn cho parse_fail giả. Chờ selector/state giá xuất hiện trước khi trích.

**Vì sao vẫn tái dùng farm (§1 #1, #2)?** Farm (FR-SCRAPE-003) đã giải fingerprint + TLS + behavior - đúng những gì Akamai kiểm. Lazada adapter chỉ thêm "đọc giá ở đâu trên trang Lazada". Tái dùng farm tránh nhân bản lớp chống Akamai; adapter chỉ là selectors + extract riêng.

---

## §3 - Hợp đồng API / DDL

### Selectors + extract (TypeScript)

```typescript
// services/scrape/farm/src/adapters/lazada/selectors.ts
export const lazadaSelectors = {
  embeddedState: 'script:has-text("window.__moduleData__"), script:has-text("pdpTrackingData")',
  priceText:     '.pdp-price--current, [data-price]',
  listPriceText: '.pdp-price--deleted, .origin-block .price',
  flashBadge:    '.pdp-mod-flash-sale, [data-spm="flashsale"]',
  readyAnchor:   '.pdp-price--current', // chờ giá render xong (§1 #9)
};
```

```typescript
// services/scrape/farm/src/adapters/lazada/extract.ts
const VND_UNIT = 1; // Lazada VN hiển thị VND nguyên; chỉnh nếu phát hiện khác

// extractLazada ưu tiên JSON nhúng (window state), fallback DOM text; giá VND số nguyên.
export function extractLazada(page: PageView): RawPrice {
  const state = page.embeddedJSON(lazadaSelectors.embeddedState);
  if (state) {
    const p = pickPriceFromModuleData(state); // đọc price/origin/flash từ __moduleData__
    if (p) return p;                           // source = json
  }
  return {
    price:     parseVNDInt(page.text(lazadaSelectors.priceText)),
    listPrice: parseVNDInt(page.text(lazadaSelectors.listPriceText)),
    flashSale: page.exists(lazadaSelectors.flashBadge),
    source:    'dom',
  };
}
```

### Adapter (TypeScript)

```typescript
// services/scrape/farm/src/adapters/lazada/adapter.ts
export class LazadaAdapter implements PlatformAdapter {
  platformID() { return PLATFORM_LAZADA; }

  async fetch(ctx: BrowserContext, job: ScrapeJob): Promise<PriceSnapshot> {
    // ctx do farm tạo với TLS match (FR-SCRAPE-003 #4) + proxy residential enterprise
    const page = await ctx.newPage();
    try {
      await page.goto(job.url, { waitUntil: 'domcontentloaded' });
      if (await detectAkamaiChallenge(page)) throw new ChallengedError('akamai sensor/verify');
      await humanize(page);                                  // FR-SCRAPE-003 behavior
      await page.waitForSelector(lazadaSelectors.readyAnchor);
      const raw = extractLazada(viewOf(page));
      metrics.lazadaSource(raw.source);
      return toSnapshot(job.productId, raw, VND_UNIT);       // VND số nguyên, set ts
    } finally {
      await page.close();
    }
  }
}
```

---

## §4 - Acceptance criteria

1. `LazadaAdapter` thỏa interface `PlatformAdapter`; `platformID()` trả mã Lazada.
2. Job Lazada được cấp proxy residential tier `enterprise` (`SelectTier(DiffAkamai)` = enterprise); datacenter bị từ chối.
3. Adapter chạy trên context farm có TLS/HTTP2 match (không dùng client thô bỏ qua TLS).
4. Trang có window state JSON -> `extractLazada` trả giá từ state (`source='json'`).
5. Trang không có JSON nhúng -> fallback DOM text (`source='dom'`), giá VND đúng (số nguyên).
6. `flash_sale` = true khi trang có badge flash sale Lazada.
7. Giá quy đổi bằng số nguyên; với text " VND89.000" -> `price=89000` chính xác.
8. Trang Akamai challenge (sensor/verify) -> `fetch` ném `ChallengedError`, không trả snapshot.
9. Giá chưa render -> adapter chờ `readyAnchor` trước khi trích (không parse trang chưa sẵn).
10. Adapter báo outcome (success/parse_fail/challenge) cho monitor FR-SCRAPE-006.
11. Page đóng sau render; chỉ trường giá được giữ (không lưu trang thô).
12. Metric `lazada_extract_source_total{source}`, `lazada_akamai_challenge_total` thay đổi đúng.

---

## §5 - Kiểm thử (verification)

```typescript
// services/scrape/farm/src/adapters/lazada/__tests__/extract.test.ts
test('ưu tiên window state JSON', () => {
  const page = pageWithModuleData({ price: 89000, originPrice: 149000, flash: true });
  const raw = extractLazada(page);
  expect(raw.source).toBe('json');
  expect(raw.price).toBe(89000);
  expect(raw.listPrice).toBe(149000);
  expect(raw.flashSale).toBe(true);
});

test('fallback DOM text', () => {
  const page = pageDomOnly({ price: '₫89.000', origin: '₫149.000', flash: false });
  const raw = extractLazada(page);
  expect(raw.source).toBe('dom');
  expect(raw.price).toBe(89000);
});

test('parse VND số nguyên', () => {
  expect(parseVNDInt('₫1.234.567')).toBe(1234567);
});
```

```typescript
// services/scrape/farm/src/adapters/lazada/__tests__/adapter.test.ts
test('Akamai challenge ném ChallengedError', async () => {
  const ctx = ctxWithAkamaiSensor();
  await expect(new LazadaAdapter().fetch(ctx, job('https://lazada.vn/x')))
    .rejects.toBeInstanceOf(ChallengedError);
});

test('chờ render trước khi trích', async () => {
  const ctx = ctxSlowRender(lazadaSelectors.readyAnchor);
  const snap = await new LazadaAdapter().fetch(ctx, job('u'));
  expect(snap.price).toBeGreaterThan(0); // không parse trang chưa sẵn
});

test('proxy datacenter bị từ chối cho Lazada', () => {
  expect(() => assignProxy(job('lazada'), 'datacenter')).toThrow();
  expect(selectTierFor('lazada')).toBe('enterprise');
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: selectors.ts -> extract.ts (JSON-ưu-tiên + DOM-fallback, test trước với fixture) -> adapter.ts (render qua farm TLS-match + chờ render + Akamai challenge) -> registry.go (đăng ký, gắn proxy enterprise) -> tests. Adapter chạy trên farm Playwright của FR-SCRAPE-003 (đặc biệt cần TLS/HTTP2 match cho Akamai). Proxy lấy qua `SelectTier(DiffAkamai)` của FR-SCRAPE-004. Outcome báo về monitor FR-SCRAPE-006.

---

## §7 - Phụ thuộc

- **FR-SCRAPE-003** - farm Playwright với TLS/HTTP2 match (then chốt cho Akamai), behavior, ChallengedError.
- **FR-SCRAPE-004** - proxy residential tier enterprise (DiffAkamai); datacenter bị từ chối.
- **FR-SCRAPE-001** - interface adapter; orchestrator điều phối + gọi sink.
- **FR-SCRAPE-005** - pacing + xử lý ChallengedError; ghi qua delta-only.
- **FR-SCRAPE-006** - monitor nhận outcome.
- **FR-PRICE-002 / FR-PRICE-004 (downstream)** - snapshot ghi qua InsertSnapshot; Lazada là chân thứ ba của so sánh chéo.

---

## §8 - Payload ví dụ

### window state nhúng (rút gọn, ẩn danh)

```json
{
  "__moduleData__": {
    "data": { "root": { "fields": {
      "price": { "value": 89000, "originalPrice": 149000 },
      "flashSale": { "active": true }
    } } }
  }
}
```

### Snapshot adapter trả về

```typescript
{
  productId: 55310,
  ts: nowISO,
  price: 89000,            // VND số nguyên
  listPrice: 149000,
  flashSale: true,
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Hình dạng `__moduleData__` chính xác của Lazada VN - xác minh trên fixture thật; `pickPriceFromModuleData` một chỗ để chỉnh.
- Mức độ Akamai siết theo thời gian (sensor data v2/v3) - dựa farm TLS-match; nâng khi Akamai đổi.
- Đơn vị giá - hằng `VND_UNIT` một chỗ, khóa bằng test.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Datacenter proxy cho Lazada | tỷ lệ block | Akamai chặn chắc | Residential enterprise (§1 #3) |
| Client thô bỏ TLS match | block ở bắt tay | lộ trước cả JS | Farm TLS/HTTP2 match (§1 #2) |
| Đọc DOM chưa render | parse trang chưa sẵn | parse_fail giả | Chờ readyAnchor (§1 #9) |
| window state đổi shape | `lazada_extract_source` lệch | json fail | Fallback DOM + monitor (§1 #4,#8) |
| Float trong quy đổi | extract test | sai số tiền | parseVNDInt số nguyên (§1 #12) |
| Akamai challenge coi như "không giá" | ChallengedError | dữ liệu bẩn | Ném lỗi, chuyển FR-SCRAPE-005 (§1 #7) |
| Đọc tức thì như bot | behavior | dấu vết | humanize trước trích (§1 #6) |
| Cấu trúc Lazada đổi | parse_fail spike | dữ liệu chết | Monitor drift (§1 #8) |
| Lưu trang thô | review code | rủi ro dữ liệu | Chỉ trích trường giá (§1 #10) |
| Rò page farm | mem profiler | RAM phình | Đóng page finally (§1 #11) |

---

## §11 - Ghi chú

- Lazada là chân thứ ba của moat so sánh chéo 3 sàn (FR-PRICE-004) - thiếu nó thì so sánh giá yếu hẳn.
- Khác biệt cốt lõi của Lazada là Akamai: kiểm fingerprint ở tầng TLS/HTTP2 trước khi JS chạy - TLS match của farm (FR-SCRAPE-003 #4) là điều kiện sống còn, không phải tùy chọn.
- Residential enterprise bắt buộc vì datacenter vô dụng với Akamai (§3.3) - đây không phải lựa chọn tiết kiệm.
- Cùng khuôn với TikTok (FR-SCRAPE-007): đọc DOM-render, JSON nhúng ưu tiên, chờ render, né API ký; khác ở lớp Akamai TLS.
- Giá VND số nguyên đồng bộ FR-PRICE-002 - sai số float làm hỏng so sánh chéo sàn vốn là điểm mạnh của Lazada trong moat.

---

*Hết FR-SCRAPE-008. Status: ready_to_implement (mục tiêu audit 10/10).*
