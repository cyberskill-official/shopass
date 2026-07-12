---
id: FR-SCRAPE-003
title: "Playwright headless farm + anti-fingerprint - spoof Canvas/WebGL/AudioContext, JA3/JA4 TLS, HTTP/2 settings, hành vi giống người"
module: SCRAPE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-SCRAPE-001, FR-SCRAPE-002, FR-SCRAPE-004, FR-SCRAPE-005, FR-SCRAPE-007, FR-SCRAPE-008, FR-PRICE-002]
depends_on: [FR-SCRAPE-001]
blocks: [FR-SCRAPE-004, FR-SCRAPE-007, FR-SCRAPE-008]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.3 (headless browser farm Playwright + anti-fingerprint: Canvas/WebGL/AudioContext, JA3/JA4 TLS, HTTP/2)"
  - "docs/... §3.9 (anti-bot per-sàn: fingerprinting Canvas/WebGL device hash, Akamai TLS/HTTP)"
source_decisions:
  - "DEC-SCRAPE-10: farm dùng Playwright headless với patch chống fingerprint (spoof Canvas/WebGL/AudioContext nhất quán theo profile)"
  - "DEC-SCRAPE-11: TLS fingerprint (JA3/JA4) và HTTP/2 SETTINGS phải khớp trình duyệt thật của profile, không để lộ dấu vết automation"
  - "DEC-SCRAPE-12: một fingerprint profile gắn liền một proxy session (FR-SCRAPE-004) - không trộn fingerprint nước này với IP nước khác"
  - "DEC-SCRAPE-13: hành vi giống người (mouse move, scroll, delay đọc trang) trước khi trích dữ liệu, không request thô tức thì"
  - "DEC-SCRAPE-14: farm expose RenderPrice(ctx, job) cho adapter fallback (FR-SCRAPE-002/007/008)"

language: "Node.js 20 + Playwright (farm-worker); patch fingerprint qua addInitScript + CDP; TLS qua undici/custom client"
service: shopass/services/scrape/
new_files:
  - services/scrape/farm/src/browser.ts
  - services/scrape/farm/src/fingerprint.ts
  - services/scrape/farm/src/behavior.ts
  - services/scrape/farm/src/render.ts
  - services/scrape/farm/src/__tests__/fingerprint.test.ts
  - services/scrape/farm/src/__tests__/behavior.test.ts
  - services/scrape/farm/src/__tests__/render.test.ts
modified_files:
  - services/scrape/farm/src/config.ts                    # thêm profile pool + binding proxy session
allowed_tools:
  - file_read: services/scrape/farm/**
  - file_write: services/scrape/farm/**
  - bash: cd services/scrape/farm && npm test
disallowed_tools:
  - chạy Playwright mặc định không patch fingerprint (vi phạm DEC-SCRAPE-10, lộ headless ngay)
  - trộn fingerprint profile với proxy khác nước (vi phạm DEC-SCRAPE-12, mâu thuẫn locale/timezone/IP)
  - request thô tức thì không có hành vi người (vi phạm DEC-SCRAPE-13)

effort_hours: 12
sub_tasks:
  - "2.0h: fingerprint.ts - sinh profile nhất quán (UA, platform, languages, timezone, screen, WebGL vendor/renderer)"
  - "2.5h: browser.ts - khởi tạo Playwright context + addInitScript spoof Canvas/WebGL/AudioContext readback"
  - "2.0h: TLS/HTTP2 - client khớp JA3/JA4 + HTTP/2 SETTINGS theo profile (undici/boringssl tuning)"
  - "1.5h: behavior.ts - mouse path, scroll, dwell time ngẫu nhiên trước khi đọc DOM"
  - "1.5h: render.ts - RenderPrice(job): mở trang, chạy behavior, trích giá DOM -> PriceSnapshot"
  - "1.0h: fingerprint.test.ts - profile nội nhất quán (timezone<->locale<->UA), readback Canvas/WebGL bị spoof"
  - "1.0h: behavior.test.ts - có mouse/scroll/dwell trước trích; render.test.ts - DOM giả -> snapshot đúng"
  - "0.5h: OTel metric farm_render_total{platform,outcome} + farm_detected_total"

risk_if_skipped: "Anti-bot 3 sàn dựa trên fingerprinting (§3.9): Shopee hash thiết bị từ Canvas/WebGL/hành vi; Akamai của Lazada đọc TLS/HTTP. Playwright mặc định lộ headless tức thì (navigator.webdriver, Canvas readback đặc trưng, JA3 automation) -> bị ban ngay, đốt proxy. Không có farm patch thì fallback của FR-SCRAPE-002 vô dụng, TikTok (FR-SCRAPE-007) và Lazada (FR-SCRAPE-008) không quét được. Trộn fingerprint nước này với IP nước khác tạo mâu thuẫn locale/timezone/IP - cờ đỏ phát hiện bot. Đây là lớp chống ban tốn công nhất và là điều kiện để 2/3 sàn có dữ liệu."
---

## §1 - Mô tả (BCP-14 normative)

Farm headless **MUST** chạy Playwright với fingerprint nhất quán và hành vi giống người, expose `RenderPrice` cho adapter fallback. Hợp đồng:

1. **MUST** sinh fingerprint profile nhất quán nội bộ (DEC-SCRAPE-10): `userAgent`, `platform`, `languages`, `timezone`, `screen` (width/height/dpr), `hardwareConcurrency`, `deviceMemory`, WebGL `vendor`/`renderer` - mọi trường phải đồng bộ (ví dụ timezone `Asia/Ho_Chi_Minh` <-> `languages` chứa `vi-VN`).
2. **MUST** spoof readback fingerprint qua `addInitScript`/CDP để cùng một profile cho cùng giá trị Canvas hash, WebGL parameters, AudioContext fingerprint giữa các lần render (không random mỗi request - bất nhất cũng là cờ đỏ).
3. **MUST** ẩn dấu vết automation: `navigator.webdriver=false`, không để lộ biến CDP/Playwright, các thuộc tính `window.chrome` hợp lý cho profile.
4. **MUST** khớp TLS fingerprint (JA3/JA4) và HTTP/2 SETTINGS với trình duyệt thật của profile (DEC-SCRAPE-11) - cipher order, ALPN, HTTP/2 pseudo-header order, SETTINGS frame phải giống Chrome/Firefox thật, không để lộ dấu Go/Node default client. Đặc biệt cần cho Akamai của Lazada (§3.9).
5. **MUST** gắn một fingerprint profile với một proxy session (DEC-SCRAPE-12): profile có timezone/locale VN phải đi cùng IP residential VN của FR-SCRAPE-004; KHÔNG trộn timezone nước này với IP nước khác.
6. **MUST** chèn hành vi giống người trước khi trích dữ liệu (DEC-SCRAPE-13): di chuyển chuột theo đường cong, scroll, thời gian dừng đọc (dwell) ngẫu nhiên trong khoảng người-thật, không bắn request và trích DOM tức thì.
7. **MUST** expose `RenderPrice(ctx, job ScrapeJob) (PriceSnapshot, error)` (DEC-SCRAPE-14): mở trang sản phẩm, chạy behavior, trích giá từ DOM render, trả `PriceSnapshot` cho adapter gọi fallback (FR-SCRAPE-002/007/008).
8. **MUST** phát hiện khi bị challenge (CAPTCHA/slider/verify page) và trả error phân loại `ErrChallenged` để FR-SCRAPE-005 xử lý CAPTCHA hoặc orchestrator backoff, KHÔNG trả snapshot rỗng coi như giá hợp lệ.
9. **MUST** giới hạn vòng đời context/page và dọn tài nguyên (đóng context sau render) để tránh rò bộ nhớ trên farm dài hạn.
10. **SHOULD** xoay vòng profile pool: chọn profile ít dùng gần đây, không tái dùng cùng một (fingerprint + IP) liên tục cho cùng một sàn trong cửa sổ ngắn.
11. **SHOULD** phát OTel metric: `farm_render_total{platform, outcome}` (counter), `farm_detected_total{platform, kind}` (counter, kind=captcha|block|tls), `farm_render_duration_ms{platform}` (histogram).
12. **MUST** không lưu trữ trang thô; chỉ trích các trường giá rồi đóng trang (tối thiểu hóa dữ liệu, §5.4).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao fingerprint nhất quán nội bộ (DEC-SCRAPE-10)?** Anti-bot không chỉ chặn "headless"; nó chấm điểm mâu thuẫn. Một trình duyệt khai timezone `Asia/Ho_Chi_Minh` nhưng `languages=en-US`, screen lạ, WebGL vendor không khớp UA là cờ đỏ. Profile phải tự nhất quán như một thiết bị thật, không phải một mớ trường random.

**Vì sao spoof readback ổn định (§1 #2)?** Fingerprinting đọc Canvas/WebGL/AudioContext để tạo hash thiết bị (§3.9 nói rõ Shopee làm vậy). Nếu hash đổi mỗi request, server thấy "một người dùng" có thiết bị biến hình từng giây - bất thường rõ rệt. Spoof phải ổn định cho mỗi profile, đổi khi đổi profile.

**Vì sao khớp TLS/HTTP2 (DEC-SCRAPE-11)?** Akamai (Lazada) và nhiều WAF đọc JA3/JA4 ở tầng bắt tay TLS, trước cả khi JS chạy. Client Go/Node mặc định có chữ ký TLS khác hẳn Chrome thật - bị nhận diện ngay dù JS có patch hoàn hảo. Phải khớp cipher order, ALPN, HTTP/2 SETTINGS để qua tầng này.

**Vì sao gắn fingerprint với proxy session (DEC-SCRAPE-12)?** IP residential mang theo geolocation. Một profile VN (timezone, ngôn ngữ VN) đi qua IP Mỹ là mâu thuẫn không thể giải thích. Buộc cặp (profile <-> proxy cùng nước) giữ câu chuyện nhất quán từ tầng mạng tới tầng trình duyệt.

**Vì sao hành vi giống người (DEC-SCRAPE-13)?** Hệ thống hành vi (behavioral) đo nhịp tương tác. Tải trang rồi đọc giá trong 50ms không phải hành vi người. Mouse move, scroll, dwell time đưa farm vào phân bố hành vi người-thật, phối hợp với pacing/jitter của FR-SCRAPE-005.

**Vì sao tách ErrChallenged (§1 #8)?** Khi bị challenge, trả "không có giá" làm bẩn dữ liệu và làm orchestrator coi như parse fail. Phân loại challenge rõ ràng để FR-SCRAPE-005 quyết định giải CAPTCHA hay lùi, và để metric `farm_detected_total` cảnh báo sớm khi tỷ lệ challenge tăng.

---

## §3 - Hợp đồng API / DDL

### Fingerprint profile (TypeScript)

```typescript
// services/scrape/farm/src/fingerprint.ts
export interface FingerprintProfile {
  userAgent: string;
  platform: string;            // 'Win32' | 'Linux x86_64' | ...
  languages: string[];         // ['vi-VN','vi','en-US']
  timezone: string;            // 'Asia/Ho_Chi_Minh'
  screen: { width: number; height: number; dpr: number };
  hardwareConcurrency: number;
  deviceMemory: number;
  webgl: { vendor: string; renderer: string };
  canvasNoiseSeed: number;     // seed ổn định cho readback spoof của profile
}

// Tạo profile nhất quán: timezone, languages, UA, WebGL khớp nhau theo nước.
export function makeProfile(country: 'VN' | 'ID' | 'TH', seed: number): FingerprintProfile;

// Kiểm tra nội nhất quán (dùng trong test + guard runtime).
export function isCoherent(p: FingerprintProfile): boolean;
```

### Patch trình duyệt (TypeScript)

```typescript
// services/scrape/farm/src/browser.ts
export async function newPatchedContext(
  browser: Browser, p: FingerprintProfile, proxy: ProxySession,
): Promise<BrowserContext> {
  const ctx = await browser.newContext({
    userAgent: p.userAgent,
    locale: p.languages[0],
    timezoneId: p.timezone,
    viewport: { width: p.screen.width, height: p.screen.height },
    deviceScaleFactor: p.screen.dpr,
    proxy: { server: proxy.url, username: proxy.user, password: proxy.pass },
  });
  await ctx.addInitScript(spoofScript(p)); // navigator.webdriver=false, Canvas/WebGL/AudioContext spoof
  return ctx;
}
```

### Render fallback (TypeScript)

```typescript
// services/scrape/farm/src/render.ts
export class ChallengedError extends Error {}

export async function renderPrice(
  ctx: BrowserContext, job: ScrapeJob, sel: PriceSelectors,
): Promise<PriceSnapshot> {
  const page = await ctx.newPage();
  try {
    await page.goto(job.url, { waitUntil: 'domcontentloaded' });
    if (await detectChallenge(page)) throw new ChallengedError('captcha/slider/verify');
    await humanize(page);                       // behavior.ts: mouse/scroll/dwell
    const raw = await extractDom(page, sel);    // đọc price/list_price/stock/flash
    return toSnapshot(job.productId, raw);      // quy đổi VND số nguyên, set ts
  } finally {
    await page.close();                         // dọn tài nguyên (§1 #9)
  }
}
```

---

## §4 - Acceptance criteria

1. `makeProfile('VN', seed)` trả profile có `timezone='Asia/Ho_Chi_Minh'` và `languages` chứa `vi-VN`; `isCoherent` = true.
2. `isCoherent` trả false khi timezone và locale mâu thuẫn (ví dụ timezone VN + `languages=['en-US']` không kèm vi).
3. Cùng một profile (cùng seed) -> Canvas readback hash giống nhau giữa 2 lần render; profile khác -> hash khác.
4. Context đã patch: `navigator.webdriver === false` và không lộ biến automation (kiểm qua page evaluate).
5. WebGL `UNMASKED_VENDOR/RENDERER` trả đúng giá trị của profile, không phải giá trị mặc định của môi trường CI.
6. TLS client của farm tạo JA3/JA4 khớp profile mục tiêu (so với chữ ký Chrome thật trong fixture), khác chữ ký Node/Go default.
7. Profile VN luôn được ghép với proxy session nước VN; ghép sai nước bị từ chối (guard `bindProfileProxy`).
8. `renderPrice` chạy `humanize` (có mouse move/scroll/dwell) TRƯỚC khi `extractDom` (kiểm thứ tự qua spy).
9. Trang challenge (fixture có slider/verify) -> `renderPrice` ném `ChallengedError`, không trả snapshot.
10. DOM giá hợp lệ -> `renderPrice` trả `PriceSnapshot` với giá VND đúng (số nguyên) và `flash_sale` đúng.
11. Page được đóng sau mỗi render (kiểm không rò context/page).
12. Metric `farm_render_total{outcome}` và `farm_detected_total{kind}` tăng đúng theo nhánh.

---

## §5 - Kiểm thử (verification)

```typescript
// services/scrape/farm/src/__tests__/fingerprint.test.ts
test('profile VN nhất quán timezone/locale', () => {
  const p = makeProfile('VN', 42);
  expect(p.timezone).toBe('Asia/Ho_Chi_Minh');
  expect(p.languages).toContain('vi-VN');
  expect(isCoherent(p)).toBe(true);
});

test('mâu thuẫn timezone vs locale bị bắt', () => {
  const p = { ...makeProfile('VN', 1), languages: ['en-US'] };
  expect(isCoherent(p)).toBe(false);
});

test('Canvas readback ổn định theo seed', async () => {
  const a = await canvasHash(makeProfile('VN', 7));
  const b = await canvasHash(makeProfile('VN', 7));
  const c = await canvasHash(makeProfile('VN', 8));
  expect(a).toBe(b);          // cùng profile -> cùng hash
  expect(a).not.toBe(c);      // khác profile -> khác hash
});
```

```typescript
// services/scrape/farm/src/__tests__/behavior.test.ts
test('humanize chạy trước extract', async () => {
  const calls: string[] = [];
  const page = fakePage({
    onHumanize: () => calls.push('humanize'),
    onExtract: () => calls.push('extract'),
  });
  await renderPrice(ctxFor(page), job('https://shopee.vn/x'), selectors);
  expect(calls).toEqual(['humanize', 'extract']); // không trích thô tức thì
});

// services/scrape/farm/src/__tests__/render.test.ts
test('trang challenge ném ChallengedError', async () => {
  const page = fakePage({ challenge: true });
  await expect(renderPrice(ctxFor(page), job('u'), selectors))
    .rejects.toBeInstanceOf(ChallengedError);
});

test('webdriver ẩn sau patch', async () => {
  const ctx = await newPatchedContext(browser, makeProfile('VN', 1), proxyVN);
  const page = await ctx.newPage();
  const wd = await page.evaluate(() => (navigator as any).webdriver);
  expect(wd).toBe(false);
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: fingerprint.ts (profile + isCoherent, test trước) -> browser.ts (context + spoof script) -> tuning TLS/HTTP2 -> behavior.ts (humanize) -> render.ts (RenderPrice + detectChallenge) -> tests. Spoof script chèn qua `addInitScript` chạy trước script trang. Profile pool và binding proxy session ở config.ts; mỗi (profile <-> proxy) cố định theo nước. Selectors giá per-sàn đến từ adapter gọi (Shopee/TikTok/Lazada).

---

## §7 - Phụ thuộc

- **FR-SCRAPE-001** - orchestrator gọi adapter; adapter gọi farm khi fallback; kiểu `ScrapeJob`.
- **FR-SCRAPE-002 / 007 / 008** - các adapter gọi `RenderPrice` của farm khi gặp challenge / sàn buộc DOM-render.
- **FR-SCRAPE-004** - proxy session residential gắn theo profile (cùng nước).
- **FR-SCRAPE-005 (downstream)** - `ChallengedError` chuyển sang xử lý CAPTCHA + pacing.
- **FR-PRICE-002 (downstream)** - snapshot từ farm ghi qua `InsertSnapshot` delta-only.
- Lib: Playwright, undici/boringssl tuning cho TLS, CDP cho spoof readback.

---

## §8 - Payload ví dụ

### Profile VN sinh ra (rút gọn)

```json
{
  "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 ... Chrome/...",
  "platform": "Win32",
  "languages": ["vi-VN", "vi", "en-US"],
  "timezone": "Asia/Ho_Chi_Minh",
  "screen": { "width": 1920, "height": 1080, "dpr": 1 },
  "hardwareConcurrency": 8,
  "deviceMemory": 8,
  "webgl": { "vendor": "Google Inc. (NVIDIA)", "renderer": "ANGLE (NVIDIA ...)" }
}
```

### Adapter gọi farm khi fallback (nội bộ)

```typescript
// trong ShopeeAdapter khi nhận HTML challenge (FR-SCRAPE-002 #6)
const snap = await farm.renderPrice(ctx, job, shopeeSelectors);
// farm tự chọn profile + proxy VN, chạy humanize, trích giá, trả PriceSnapshot
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Thư viện anti-detect cụ thể (patch tự viết vs giải pháp thương mại) - đánh giá khi đo `farm_detected_total` thực tế.
- Mức độ giống người của behavior (mô hình hóa quỹ đạo chuột) - bắt đầu đơn giản, nâng theo tỷ lệ phát hiện.
- Pool profile lớn tới đâu để không tái dùng cặp (fp+IP) quá dày - tinh chỉnh theo quy mô farm.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Playwright mặc định lộ headless | `farm_detected_total` cao | ban tức thì | Patch fingerprint bắt buộc (§1 #1-#3) |
| Hash Canvas đổi mỗi request | property test | bất nhất -> cờ đỏ | Spoof ổn định theo seed (§1 #2) |
| JA3/JA4 lộ Node/Go client | test TLS | Akamai chặn (Lazada) | Khớp TLS/HTTP2 profile (§1 #4) |
| Profile VN + IP nước khác | guard bindProfileProxy | mâu thuẫn geo | Cặp (fp <-> proxy) cùng nước (§1 #5) |
| Trích DOM tức thì không behavior | spy thứ tự | hành vi bot | humanize trước extract (§1 #6) |
| Challenge coi như "không giá" | `ChallengedError` | dữ liệu bẩn | Phân loại + chuyển FR-SCRAPE-005 (§1 #8) |
| Rò context/page | mem profiler | farm phình RAM | Đóng page trong finally (§1 #9) |
| Tái dùng cặp fp+IP quá dày | tần suất render | cụm bị ban chùm | Xoay profile pool (§1 #10) |
| Lưu nhầm trang thô | review code | rủi ro dữ liệu | Chỉ trích trường giá (§1 #12) |
| Selector DOM đổi (A/B) | render fail tăng | parse hụt | FR-SCRAPE-006 giám sát + cập nhật selectors |

---

## §11 - Ghi chú

- Farm là lớp chống ban tốn công nhất và là điều kiện để Lazada/TikTok có dữ liệu (cả hai buộc DOM-render, §3.2).
- Anti-bot chấm điểm mâu thuẫn, không chỉ chặn headless: nhất quán (timezone<->locale<->WebGL<->IP) quan trọng hơn từng trick lẻ.
- TLS/HTTP2 fingerprint nằm dưới JS - phải khớp trước khi script trang chạy, đặc biệt với Akamai của Lazada.
- Farm là đường đắt; adapter ưu tiên JSON (FR-SCRAPE-002) và chỉ gọi farm khi buộc, để giữ unit economics (§4.1).
- Behavior người-thật ở đây phối hợp với pacing/jitter của FR-SCRAPE-005 - hai lớp chống hành vi-bot bổ trợ nhau.

---

*Hết FR-SCRAPE-003. Status: ready_to_implement (mục tiêu audit 10/10).*
