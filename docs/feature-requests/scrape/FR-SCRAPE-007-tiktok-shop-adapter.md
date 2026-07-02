---
id: FR-SCRAPE-007
title: "TikTok Shop scraping adapter - ưu tiên đọc DOM-render, né API ký (msToken/_signature/X-Bogus), webview/SPA"
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
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (TikTok ký request msToken/_signature/X-Bogus + app attestation -> ưu tiên đọc DOM render)"
  - "docs/... §3.9 (TikTok Shop High risk; hệ ByteDance ký request; app attestation mạnh)"
source_decisions:
  - "DEC-SCRAPE-28: TikTok Shop ưu tiên đọc DOM-render qua farm Playwright (FR-SCRAPE-003), KHÔNG cố gọi/ký API nội bộ"
  - "DEC-SCRAPE-29: không tái tạo msToken/_signature/X-Bogus (app attestation mạnh, leo thang đối đầu + giòn)"
  - "DEC-SCRAPE-30: dùng proxy tier enterprise (FR-SCRAPE-004 DiffByteDance) vì rủi ro ban High"
  - "DEC-SCRAPE-31: parse giá từ JSON nhúng trong trang (__NEXT_DATA__/SIGI_STATE-like) nếu có, fallback DOM text"

language: "Node.js 20 + Playwright (adapter chạy trên farm); parse JSON nhúng / DOM text -> PriceSnapshot"
service: shopass/services/scrape/
new_files:
  - services/scrape/farm/src/adapters/tiktok/adapter.ts
  - services/scrape/farm/src/adapters/tiktok/extract.ts
  - services/scrape/farm/src/adapters/tiktok/selectors.ts
  - services/scrape/farm/src/adapters/tiktok/__tests__/extract.test.ts
  - services/scrape/farm/src/adapters/tiktok/__tests__/adapter.test.ts
  - services/scrape/farm/src/adapters/tiktok/__tests__/fixtures/pdp.html
modified_files:
  - services/scrape/internal/orchestrator/registry.go     # đăng ký TikTokAdapter theo platform_id
allowed_tools:
  - file_read: services/scrape/**
  - file_write: services/scrape/**
  - bash: cd services/scrape/farm && npm test
disallowed_tools:
  - cố tái tạo/ký msToken/_signature/X-Bogus (vi phạm DEC-SCRAPE-29, giòn + leo thang)
  - dùng proxy budget/datacenter cho TikTok (vi phạm DEC-SCRAPE-30, risk High)
  - lưu giá float (vi phạm DEC-PRICE-05 của FR-PRICE-002)

effort_hours: 10
sub_tasks:
  - "2.0h: selectors.ts - định vị giá/list_price/stock/flash trong DOM TikTok Shop PDP + JSON nhúng"
  - "2.5h: extract.ts - parse JSON nhúng (__NEXT_DATA__-like) ưu tiên, fallback DOM text -> giá VND số nguyên"
  - "2.0h: adapter.ts - render qua farm (FR-SCRAPE-003), chạy behavior, gọi extract -> PriceSnapshot"
  - "1.5h: fixtures/pdp.html + extract.test.ts - trang thật ẩn danh -> assert price/list_price/flash"
  - "1.0h: adapter.test.ts - challenge -> ChallengedError; JSON nhúng path + DOM fallback path"
  - "1.0h: registry.go đăng ký + integration orchestrator (proxy enterprise, tier)"

risk_if_skipped: "TikTok Shop chiếm 41,31% GMV và là sàn đang tăng nhanh nhất - bỏ qua là bỏ một nửa thị trường live-commerce. Nhưng TikTok có rủi ro ban High (§3.9): ký request msToken/_signature/X-Bogus + app attestation mạnh. Nếu cố ký API là leo thang đối đầu với ByteDance và giòn (thuật toán ký đổi là hỏng). Nếu dùng proxy budget/datacenter là bị chặn ngay. Nếu lưu giá float là sai số trên so sánh chéo sàn (FR-PRICE-004). Adapter này phải đi đường an toàn: đọc DOM-render qua farm, proxy enterprise, không đụng vào API ký."
---

## §1 - Mô tả (BCP-14 normative)

Adapter TikTok Shop **MUST** lấy giá bằng đọc DOM-render qua farm Playwright, né hoàn toàn API ký, dùng proxy enterprise. Hợp đồng:

1. **MUST** triển khai interface adapter của FR-SCRAPE-001 và lấy giá qua farm `RenderPrice` của FR-SCRAPE-003 (DEC-SCRAPE-28) - TikTok buộc DOM-render, không có đường JSON-không-login như Shopee.
2. **MUST NOT** tái tạo hay ký `msToken`/`_signature`/`X-Bogus` (DEC-SCRAPE-29): adapter không gọi API nội bộ cần chữ ký; chỉ đọc trang đã render trong ngữ cảnh trình duyệt thật của farm.
3. **MUST** dùng proxy tier `enterprise` (DEC-SCRAPE-30): khi cấp proxy cho job TikTok, dùng `SelectTier(DiffByteDance)` của FR-SCRAPE-004 (= enterprise) vì rủi ro ban High.
4. **MUST** trích giá ưu tiên từ JSON nhúng trong trang (DEC-SCRAPE-31): nếu trang có state JSON (kiểu `__NEXT_DATA__`/`SIGI_STATE`), parse từ đó (ổn định hơn DOM text); fallback đọc DOM text khi không có hoặc cấu trúc đổi.
5. **MUST** map về `price.PriceSnapshot` với `price`/`list_price` BIGINT VND số nguyên (đồng bộ DEC-PRICE-05 của FR-PRICE-002), `flash_sale` true khi trang báo đang flash/live-deal.
6. **MUST** chạy hành vi giống người của farm (FR-SCRAPE-003 #6) trước khi trích - TikTok có hệ hành vi mạnh, đọc tức thì là cờ đỏ.
7. **MUST** phát hiện challenge/verify và ném `ChallengedError` (như FR-SCRAPE-003 #8) để FR-SCRAPE-005 / orchestrator xử lý, KHÔNG trả snapshot rỗng.
8. **MUST** báo outcome (success/parse_fail/challenge) cho monitor của FR-SCRAPE-006 - TikTok là webview/SPA dễ đổi cấu trúc, cần giám sát drift sát.
9. **MUST** xử lý SPA: chờ nội dung giá render xong (wait cho selector/state) thay vì đọc DOM lúc trang chưa hydrate.
10. **MUST** tối thiểu hóa: chỉ trích trường giá, không lưu trang thô (§5.4).
11. **SHOULD** phát OTel metric: `tiktok_extract_source_total{source=json|dom}` (counter), `tiktok_challenge_total` (counter), `tiktok_render_duration_ms` (histogram).
12. **MUST** quy đổi giá bằng số nguyên (không float trung gian) khi đơn vị TikTok khác VND nguyên.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao đọc DOM-render, né API ký (DEC-SCRAPE-28/29)?** Tài liệu nói rõ (§3.2): TikTok ký request ở tầng API nội bộ (`msToken`/`_signature`/`X-Bogus`) cùng app attestation mạnh. Tái tạo chữ ký là cuộc đua vũ trang: ByteDance đổi thuật toán, adapter hỏng, sửa lại, lặp mãi - giòn và tốn công vô tận. Đọc trang đã render trong trình duyệt thật của farm để chính TikTok tự tạo chữ ký hợp lệ, ta chỉ đọc kết quả.

**Vì sao proxy enterprise (DEC-SCRAPE-30)?** TikTok là rủi ro ban High (§3.9) - cao nhất trong ba sàn. Proxy budget/datacenter bị chặn ngay. `SelectTier(DiffByteDance)` của FR-SCRAPE-004 trả enterprise đúng cho trường hợp này: trả tiền cao để đổi lấy độ tin cậy, vì TikTok không tha IP chất lượng thấp.

**Vì sao ưu tiên JSON nhúng (DEC-SCRAPE-31)?** SPA thường nhúng state JSON trong trang để hydrate. Đọc giá từ JSON này ổn định hơn nhiều so với cào DOM text (text đổi theo layout/A/B, JSON state ít đổi hơn). Fallback DOM text giữ độ phủ khi cấu trúc JSON đổi.

**Vì sao chờ hydrate SPA (§1 #9)?** TikTok Shop là webview/SPA: HTML ban đầu rỗng, giá render bằng JS sau. Đọc DOM ngay lúc `domcontentloaded` sẽ thấy trang trống. Phải chờ selector/state giá xuất hiện trước khi trích, nếu không sẽ parse_fail giả.

**Vì sao giám sát drift sát (§1 #8)?** TikTok là SPA, đổi cấu trúc thường xuyên hơn trang truyền thống. Báo outcome đều cho monitor (FR-SCRAPE-006) để bắt sớm khi TikTok đổi state shape/selectors - adapter SPA hỏng âm thầm là rủi ro lớn hơn adapter JSON-API.

**Vì sao vẫn dùng farm thay adapter riêng (§1 #1)?** Farm (FR-SCRAPE-003) đã giải bài fingerprint + TLS + behavior. TikTok adapter chỉ thêm phần "đọc giá ở đâu trên trang TikTok". Tái dùng farm tránh nhân bản lớp chống ban; adapter chỉ là selectors + extract logic riêng của TikTok.

---

## §3 - Hợp đồng API / DDL

### Selectors + extract (TypeScript)

```typescript
// services/scrape/farm/src/adapters/tiktok/selectors.ts
export const tiktokSelectors = {
  embeddedState: 'script#__NEXT_DATA__, script#SIGI_STATE',
  priceText:     '[data-e2e="product-price"]',
  listPriceText: '[data-e2e="product-origin-price"]',
  flashBadge:    '[data-e2e="flash-sale-badge"], [data-e2e="live-deal"]',
  readyAnchor:   '[data-e2e="product-price"]', // chờ selector này trước khi trích (SPA hydrate)
};
```

```typescript
// services/scrape/farm/src/adapters/tiktok/extract.ts
const VND_UNIT = 1; // TikTok hiển thị VND nguyên; chỉnh nếu phát hiện micro-đơn-vị

// extractTikTok ưu tiên JSON nhúng, fallback DOM text; trả giá VND số nguyên.
export function extractTikTok(page: PageView): RawPrice {
  const state = page.embeddedJSON(tiktokSelectors.embeddedState);
  if (state) {
    const p = pickPriceFromState(state); // đọc price/origin/flash từ state JSON
    if (p) return p;                      // source = json
  }
  // fallback DOM text
  return {
    price:     parseVNDInt(page.text(tiktokSelectors.priceText)),       // bỏ ký tự '₫', '.', số nguyên
    listPrice: parseVNDInt(page.text(tiktokSelectors.listPriceText)),
    flashSale: page.exists(tiktokSelectors.flashBadge),
    source:    'dom',
  };
}
```

### Adapter (TypeScript)

```typescript
// services/scrape/farm/src/adapters/tiktok/adapter.ts
export class TikTokAdapter implements PlatformAdapter {
  platformID() { return PLATFORM_TIKTOK; }

  async fetch(ctx: BrowserContext, job: ScrapeJob): Promise<PriceSnapshot> {
    const page = await ctx.newPage();
    try {
      await page.goto(job.url, { waitUntil: 'domcontentloaded' });
      if (await detectChallenge(page)) throw new ChallengedError('tiktok verify'); // không ký API
      await humanize(page);                                  // FR-SCRAPE-003 behavior
      await page.waitForSelector(tiktokSelectors.readyAnchor); // chờ SPA hydrate
      const raw = extractTikTok(viewOf(page));
      metrics.tiktokSource(raw.source);
      return toSnapshot(job.productId, raw, VND_UNIT);       // VND số nguyên, set ts
    } finally {
      await page.close();
    }
  }
}
```

---

## §4 - Acceptance criteria

1. `TikTokAdapter` thỏa interface `PlatformAdapter`; `platformID()` trả mã TikTok.
2. Job TikTok được cấp proxy tier `enterprise` (`SelectTier(DiffByteDance)` = enterprise).
3. Adapter KHÔNG gọi API nội bộ cần `msToken`/`_signature`/`X-Bogus` (kiểm không có request tới endpoint ký trong test).
4. Trang có JSON nhúng -> `extractTikTok` trả giá từ state (`source='json'`).
5. Trang không có JSON nhúng -> fallback DOM text (`source='dom'`), giá VND đúng (số nguyên, đã bỏ ' VND'/'.').
6. `flash_sale` = true khi trang có badge flash/live-deal.
7. Giá quy đổi bằng số nguyên; với text " VND89.000" -> `price=89000` chính xác.
8. Trang verify/challenge -> `fetch` ném `ChallengedError`, không trả snapshot.
9. SPA chưa hydrate -> adapter chờ `readyAnchor` trước khi trích (không parse trang trống).
10. Adapter báo outcome (success/parse_fail/challenge) cho monitor FR-SCRAPE-006.
11. Page đóng sau render; chỉ trường giá được giữ (không lưu trang thô).
12. Metric `tiktok_extract_source_total{source}`, `tiktok_challenge_total` thay đổi đúng.

---

## §5 - Kiểm thử (verification)

```typescript
// services/scrape/farm/src/adapters/tiktok/__tests__/extract.test.ts
test('ưu tiên JSON nhúng', () => {
  const page = pageWithState({ product: { price: 89000, originPrice: 149000, flash: true } });
  const raw = extractTikTok(page);
  expect(raw.source).toBe('json');
  expect(raw.price).toBe(89000);
  expect(raw.listPrice).toBe(149000);
  expect(raw.flashSale).toBe(true);
});

test('fallback DOM text khi không có JSON', () => {
  const page = pageDomOnly({ price: '₫89.000', origin: '₫149.000', flash: false });
  const raw = extractTikTok(page);
  expect(raw.source).toBe('dom');
  expect(raw.price).toBe(89000);   // bỏ '₫' và '.', số nguyên
});

test('parse VND số nguyên không float', () => {
  expect(parseVNDInt('₫1.234.567')).toBe(1234567);
});
```

```typescript
// services/scrape/farm/src/adapters/tiktok/__tests__/adapter.test.ts
test('không gọi endpoint ký', async () => {
  const calls: string[] = [];
  const ctx = ctxRecordingRequests(calls);
  await new TikTokAdapter().fetch(ctx, job('https://shop.tiktok.com/x'));
  expect(calls.some(u => /msToken|_signature|X-Bogus/.test(u))).toBe(false);
});

test('challenge ném ChallengedError', async () => {
  const ctx = ctxWithChallenge();
  await expect(new TikTokAdapter().fetch(ctx, job('u'))).rejects.toBeInstanceOf(ChallengedError);
});

test('chờ hydrate trước khi trích', async () => {
  const ctx = ctxSlowHydrate(tiktokSelectors.readyAnchor);
  const snap = await new TikTokAdapter().fetch(ctx, job('u'));
  expect(snap.price).toBeGreaterThan(0); // không parse trang trống
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: selectors.ts -> extract.ts (JSON-ưu-tiên + DOM-fallback, test trước với fixture) -> adapter.ts (render qua farm + chờ hydrate + challenge) -> registry.go (đăng ký, gắn proxy enterprise) -> tests. Adapter chạy trên farm Playwright của FR-SCRAPE-003 (tái dùng fingerprint/TLS/behavior). Proxy lấy qua `SelectTier(DiffByteDance)` của FR-SCRAPE-004. Outcome báo về monitor FR-SCRAPE-006.

---

## §7 - Phụ thuộc

- **FR-SCRAPE-003** - farm Playwright (`RenderPrice`, fingerprint, behavior, ChallengedError).
- **FR-SCRAPE-004** - proxy tier enterprise cho TikTok (DiffByteDance).
- **FR-SCRAPE-001** - interface adapter; orchestrator điều phối + gọi sink.
- **FR-SCRAPE-005** - pacing + xử lý ChallengedError; ghi qua delta-only.
- **FR-SCRAPE-006** - monitor nhận outcome (SPA drift cần giám sát sát).
- **FR-PRICE-002 (downstream)** - snapshot ghi qua InsertSnapshot delta-only.

---

## §8 - Payload ví dụ

### State JSON nhúng (rút gọn, ẩn danh)

```json
{
  "product": { "price": 89000, "originPrice": 149000, "flash": true, "stock": 52 }
}
```

### Snapshot adapter trả về

```typescript
{
  productId: 77231,
  ts: nowISO,
  price: 89000,            // VND số nguyên
  listPrice: 149000,
  flashSale: true,
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Hình dạng state JSON chính xác của TikTok Shop VN - xác minh trên fixture thật; `pickPriceFromState` tập trung một chỗ để chỉnh.
- Đơn vị giá (VND nguyên hay micro) - hằng `VND_UNIT` một chỗ, khóa bằng test.
- Live-commerce (giá trong livestream động theo phút) - tier hot + re-tier xử lý tần suất; đọc giá live cụ thể để sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Cố ký API msToken/X-Bogus | grep request | giòn + leo thang | Đọc DOM-render, né API ký (§1 #2) |
| Proxy budget cho TikTok | tỷ lệ block | ban (risk High) | Enterprise tier (§1 #3) |
| Đọc DOM lúc chưa hydrate | parse trang trống | parse_fail giả | Chờ readyAnchor (§1 #9) |
| State JSON đổi shape | `tiktok_extract_source` lệch | json fail | Fallback DOM + monitor (§1 #4,#8) |
| Float trong quy đổi | extract test | sai số tiền | parseVNDInt số nguyên (§1 #12) |
| Challenge coi như "không giá" | ChallengedError | dữ liệu bẩn | Ném lỗi, chuyển FR-SCRAPE-005 (§1 #7) |
| Đọc tức thì như bot | behavior spy | dấu vết | humanize trước trích (§1 #6) |
| SPA đổi cấu trúc | parse_fail spike | dữ liệu chết | Monitor drift sát (§1 #8) |
| Lưu trang thô | review code | rủi ro dữ liệu | Chỉ trích trường giá (§1 #10) |
| Rò page farm | mem profiler | RAM phình | Đóng page finally (§1 #11) |

---

## §11 - Ghi chú

- TikTok Shop (41,31% GMV) là sàn tăng nhanh nhất; bỏ qua là bỏ một nửa thị trường live-commerce.
- Đường an toàn là đọc DOM-render qua farm, né API ký - để chính TikTok tự tạo chữ ký hợp lệ, ta chỉ đọc kết quả, tránh cuộc đua vũ trang với ByteDance.
- Proxy enterprise là bắt buộc cho rủi ro ban High (§3.9) - đây là sàn không tha IP chất lượng thấp.
- JSON nhúng ưu tiên hơn DOM text vì ổn định hơn trước A/B layout; fallback DOM giữ độ phủ.
- TikTok là SPA dễ đổi cấu trúc -> giám sát drift (FR-SCRAPE-006) quan trọng hơn so với adapter JSON-API; báo outcome đều.

---

*Hết FR-SCRAPE-007. Status: ready_to_implement (mục tiêu audit 10/10).*
