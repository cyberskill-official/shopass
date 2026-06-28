---
id: FR-EXT-005
title: "Đồng bộ extension <-> backend - auth bridge đính JWT SănDeal (KHÔNG token sàn), hàng đợi gửi OutboundPayload, WSS giữ SW sống khi cần realtime (không lạm dụng)"
module: EXT
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-EXT-001, FR-EXT-002, FR-EXT-003, FR-AUTH-002, FR-AUTH-003, FR-TRUST-002, NFR-EXT-001]
depends_on: [FR-EXT-003, FR-AUTH-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (WebSocket giữ service worker sống khi cần realtime nhưng không lạm dụng; tác vụ nặng đẩy backend)"
  - "docs/... §3.8 (token không rời client), §5.4 (local-first, không gửi cookie/mật khẩu)"
source_decisions:
  - "DEC-EXT-23: request lên backend đính JWT của SănDeal (cấp bởi FR-AUTH-002), KHÔNG BAO GIỜ token/cookie sàn; ext gắn Authorization: Bearer <jwt SănDeal> qua auth bridge"
  - "DEC-EXT-24: chỉ OutboundPayload đã qua pipeline tối thiểu hóa (FR-EXT-003) mới được đưa vào hàng đợi gửi; không có đường gửi CartReadMessage thô"
  - "DEC-EXT-25: hàng đợi gửi bền trong chrome.storage (sống qua SW kill); gửi có retry + backoff; fail-closed nếu thiếu JWT (không gửi ẩn danh dữ liệu)"
  - "DEC-EXT-26: WSS (WebSocket) CHỈ mở khi cần realtime (ví dụ nhận alert giá tức thời); mở theo nhu cầu, đóng khi xong; KHÔNG mở thường trực để 'giữ SW sống' (lạm dụng tài nguyên)"
  - "DEC-EXT-27: refresh JWT khi hết hạn qua refresh token (FR-AUTH-002); token SănDeal lưu chrome.storage.session (RAM-only), KHÔNG storage.local bền quá mức cần"
  - "DEC-EXT-28: mọi gửi lên backend qua HTTPS; KHÔNG cleartext; ext không tự xử lý nặng (giữ vai đầu đọc, DEC-EXT-05)"

language: "TypeScript 5.x; Manifest V3 service worker; fetch HTTPS + WebSocket; JWT bearer"
service: shopass/extension/
new_files:
  - extension/src/sync/auth-bridge.ts
  - extension/src/sync/queue.ts
  - extension/src/sync/sender.ts
  - extension/src/sync/ws-client.ts
  - extension/test/auth-bridge.test.ts
  - extension/test/queue-persist.test.ts
  - extension/test/sync-no-platform-token.test.ts
  - extension/test/ws-lifecycle.test.ts
modified_files:
  - extension/src/background/service-worker.ts   # nối queue -> sender; mở/đóng WSS theo nhu cầu
  - extension/src/shared/types.ts                # thêm SyncEnvelope, AuthState
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - đính token/cookie sàn vào request backend (vi phạm DEC-EXT-23 - phá cam kết token-not-on-server + niềm tin)
  - gửi CartReadMessage thô bỏ qua pipeline FR-EXT-003 (vi phạm DEC-EXT-24)
  - mở WSS thường trực để giữ SW sống (vi phạm DEC-EXT-26 - lạm dụng tài nguyên, trái ephemeral)
  - gửi dữ liệu ẩn danh khi thiếu JWT (vi phạm DEC-EXT-25 - phải fail-closed) hoặc gửi cleartext (DEC-EXT-28)

effort_hours: 6
sub_tasks:
  - "1.0h: auth-bridge.ts - lấy/refresh JWT SănDeal (FR-AUTH-002), gắn Authorization Bearer; token ở storage.session"
  - "1.0h: queue.ts - hàng đợi OutboundPayload bền trong chrome.storage; enqueue/peek/ack; sống qua SW kill"
  - "1.5h: sender.ts - gửi HTTPS có retry + backoff; fail-closed khi thiếu JWT; ack khi 2xx; xử lý 401 -> refresh"
  - "1.0h: ws-client.ts - mở WSS khi cần realtime, đóng khi xong; KHÔNG mở thường trực; reconnect có giới hạn"
  - "0.5h: nối service-worker: onMessage pipeline-done -> enqueue; alarm tick -> flush queue; mở WSS theo nhu cầu"
  - "1.0h: sync-no-platform-token.test.ts - request KHÔNG chứa cookie/token sàn; chỉ JWT SănDeal; queue-persist + ws-lifecycle test"

risk_if_skipped: "Đây là cầu nối duy nhất giữa extension và backend - nếu làm sai, hoặc dữ liệu không tới được server (mất tính năng), hoặc cầu nối trở thành kênh rò rỉ. Ranh giới sống còn: request lên backend phải đính JWT của SănDeal (danh tính người dùng trong hệ SănDeal), TUYỆT ĐỐI không phải token/cookie sàn - lẫn lộn hai loại token là đúng kiểu lỗi biến extension thành công cụ chiếm tài khoản sàn, thảm họa PDPL (§5.5, chế tài tới 5% doanh thu) + giết định vị hậu-Honey. Hàng đợi phải bền trong chrome.storage vì SW bị kill bất cứ lúc nào (NFR-EXT-001): nếu hàng đợi nằm trong biến global, dữ liệu giỏ đọc được sẽ mất khi SW ngủ giữa lúc gửi. WSS phải mở theo nhu cầu rồi đóng - mở thường trực để 'giữ SW sống' là lạm dụng tài nguyên đúng điều §3.2 cảnh báo ('không lạm dụng'). Và phải fail-closed khi thiếu JWT: thà không gửi còn hơn gửi dữ liệu mà không gắn đúng danh tính/đường bảo mật."
---

## §1 - Mô tả (BCP-14 normative)

Lớp đồng bộ **MUST** đính JWT của SănDeal (KHÔNG token sàn) cho mọi request lên backend, gửi chỉ `OutboundPayload` đã sạch qua một hàng đợi bền, và mở WSS chỉ khi cần realtime rồi đóng. Hợp đồng:

1. `auth-bridge.ts` **MUST** đính `Authorization: Bearer <jwt>` với JWT do FR-AUTH-002 cấp cho mọi request backend (DEC-EXT-23). **MUST NOT** đính, dưới bất kỳ tên header nào, token/cookie phiên của sàn (Shopee/TikTok/Lazada).
2. Chỉ `OutboundPayload` đã qua pipeline tối thiểu hóa (FR-EXT-003) **MUST** được đưa vào hàng đợi gửi (DEC-EXT-24). **MUST NOT** tồn tại đường gửi `CartReadMessage` thô bỏ qua `minimize()`.
3. `queue.ts` **MUST** lưu hàng đợi bền trong `chrome.storage` (DEC-EXT-25) để sống qua chu kỳ SW kill (NFR-EXT-001). Hàng đợi **MUST NOT** nằm trong biến module-global (mất khi SW ngủ).
4. `sender.ts` **MUST** gửi qua HTTPS với retry + backoff; chỉ `ack` (xóa khỏi hàng đợi) khi backend trả 2xx. Lỗi mạng/5xx **MUST** giữ item trong hàng đợi để thử lại - không mất dữ liệu giỏ đã đọc.
4b. Khi backend trả 401 (JWT hết hạn), `sender.ts` **MUST** gọi refresh (DEC-EXT-27) rồi thử lại; không vứt item.
5. Khi thiếu JWT (chưa đăng nhập SănDeal / refresh thất bại), gửi **MUST** fail-closed: giữ item trong hàng đợi, KHÔNG gửi dữ liệu ẩn danh không gắn danh tính (DEC-EXT-25).
6. WSS (`ws-client.ts`) **MUST** chỉ mở khi cần realtime (ví dụ nhận alert giá tức thời) và đóng khi xong (DEC-EXT-26). **MUST NOT** mở WSS thường trực chỉ để giữ SW sống - đó là lạm dụng tài nguyên (§3.2 "không lạm dụng").
7. JWT SănDeal **MUST** lưu ở `chrome.storage.session` (RAM-only, mất khi trình duyệt đóng) chứ không bền quá mức ở `storage.local` (DEC-EXT-27); refresh token theo cơ chế FR-AUTH-002.
8. Mọi giao tiếp backend **MUST** qua HTTPS/WSS (TLS); **MUST NOT** gửi cleartext (DEC-EXT-28, no-cleartext §3.8).
9. Lớp đồng bộ **MUST** giữ vai "đầu đọc nhẹ": không chạy xử lý nặng/dài (DEC-EXT-05); flush hàng đợi theo nhịp alarm (FR-EXT-001) hoặc khi có sự kiện, không vòng lặp dài làm SW vượt 5 phút.
10. Request gửi lên backend **MUST** gói trong `SyncEnvelope` typed (`payload: OutboundPayload`, không kèm bất kỳ trường credential nào); body **MUST** chỉ chứa dữ liệu đã sạch + JWT ở header (không nhúng token trong body).
11. Lớp đồng bộ **MUST** ghi metric: số payload gửi thành công, số retry, số fail-closed do thiếu JWT, số lần mở/đóng WSS - làm bằng chứng cho audit (FR-TRUST-003) và giám sát.
12. Khi người dùng đăng xuất SănDeal **MUST** xóa JWT khỏi `storage.session` và đóng WSS; hàng đợi pending xử lý theo chính sách (giữ chờ đăng nhập lại hoặc xóa theo lựa chọn người dùng), KHÔNG gửi sau khi đã đăng xuất.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao JWT SănDeal, không token sàn (DEC-EXT-23)?** Đây là ranh giới niềm tin cứng nhất của extension. Backend SănDeal chỉ cần biết "request này thuộc người dùng SănDeal nào" - đó là việc của JWT SănDeal (FR-AUTH-002). Token/cookie sàn KHÔNG BAO GIỜ được rời client (cam kết §5.4); chúng ở lại tab để content script mượn phiên qua trình duyệt (FR-EXT-002), không bao giờ đi lên server. Lẫn hai loại token là lỗi biến extension thành kênh chiếm tài khoản sàn - đúng thảm họa cần né.

**Vì sao hàng đợi bền trong chrome.storage (DEC-EXT-25)?** SW bị kill bất cứ lúc nào (NFR-EXT-001). Nếu hàng đợi nằm trong biến global, một lần SW ngủ giữa lúc gửi là mất dữ liệu giỏ đã đọc. Hàng đợi bền trong storage sống qua chu kỳ kill: SW thức dậy, đọc lại hàng đợi, gửi tiếp. Đây là hệ quả trực tiếp của quy ước "state luôn ở storage" từ FR-EXT-001.

**Vì sao fail-closed khi thiếu JWT (DEC-EXT-25)?** Thà giữ item chờ còn hơn gửi dữ liệu mà không gắn đúng danh tính/đường bảo mật. Gửi ẩn danh "cho xong" là rò rỉ: dữ liệu rời máy mà không có chủ thể rõ ràng, vừa sai về quyền riêng tư vừa khó truy vết. Fail-closed giữ tư thế đúng - không gửi khi chưa chắc.

**Vì sao WSS mở theo nhu cầu, không thường trực (DEC-EXT-26)?** WebSocket có thể giữ SW sống (đúng như §3.2 ghi), nhưng tài liệu cũng cảnh báo "không lạm dụng". Mở WSS thường trực để né cơ chế kill là chống lại tinh thần MV3 và đốt tài nguyên người dùng. Đúng cách: mở khi thực sự cần realtime (nhận alert tức thời), đóng khi xong. Polling nhẹ qua alarm lo phần còn lại.

**Vì sao JWT ở storage.session, không storage.local (DEC-EXT-27)?** Token nhạy cảm không nên bền lâu hơn cần. `storage.session` là RAM-only, mất khi trình duyệt đóng - giảm bề mặt rò token nếu máy bị truy cập. Refresh token (theo FR-AUTH-002) khôi phục phiên khi cần, nên không cần giữ JWT bền quá mức.

**Vì sao token ở header, dữ liệu sạch ở body (§1 #10)?** Tách bạch danh tính (header Authorization) khỏi nội dung (body chỉ OutboundPayload đã sạch) giữ ranh giới rõ: body không bao giờ chứa credential. Audit kiểm body thấy chỉ dữ liệu tối thiểu, kiểm header thấy chỉ JWT SănDeal - không lẫn lộn.

---

## §3 - Hợp đồng API / DDL

### types.ts (envelope + auth, typed)

```ts
// extension/src/shared/types.ts (thêm)
import type { OutboundPayload } from "../pipeline/schema";

export interface SyncEnvelope {
  payload: OutboundPayload;      // CHỈ dữ liệu đã sạch; KHÔNG credential trong body
  clientTs: number;             // epoch ms (đo độ trễ; không phải PII)
}

export interface AuthState {
  jwt?: string;                 // JWT SănDeal (storage.session, RAM-only)
  // KHÔNG có trường token/cookie sàn
}
```

### auth-bridge.ts (đính JWT SănDeal, KHÔNG token sàn)

```ts
// extension/src/sync/auth-bridge.ts
export async function authedFetch(url: string, env: SyncEnvelope): Promise<Response> {
  const jwt = await getJwt();                 // FR-AUTH-002; storage.session
  if (!jwt) throw new NoAuthError();          // fail-closed (DEC-EXT-25)
  return fetch(url, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${jwt}`,       // JWT SănDeal, KHÔNG token sàn
      "Content-Type": "application/json"
    },
    body: JSON.stringify(env)                  // body chỉ OutboundPayload đã sạch
  });
  // LƯU Ý: KHÔNG đọc document.cookie, KHÔNG đính header token sàn ở đây.
}
```

### sender.ts (retry/backoff, ack 2xx, 401 -> refresh, fail-closed)

```ts
// extension/src/sync/sender.ts
export async function flushQueue(): Promise<void> {
  for (const item of await queue.peekAll()) {
    try {
      const res = await authedFetch(SYNC_URL, item.env);
      if (res.status === 401) { await refreshJwt(); continue; } // thử lại sau refresh
      if (res.ok) { await queue.ack(item.id); metrics.sent(); }  // chỉ ack khi 2xx
      else metrics.retry();                                      // 5xx → giữ lại
    } catch (e) {
      if (e instanceof NoAuthError) { metrics.failClosed(); return; } // thiếu JWT → dừng, giữ hàng đợi
      metrics.retry();                                          // lỗi mạng → giữ lại, backoff
    }
  }
}
```

### ws-client.ts (mở theo nhu cầu, đóng khi xong)

```ts
// extension/src/sync/ws-client.ts
let ws: WebSocket | null = null;

export function openRealtime(): void {
  if (ws) return;                              // không mở trùng
  ws = new WebSocket(WSS_URL);                 // CHỈ khi cần realtime
  ws.onclose = () => { ws = null; };
}
export function closeRealtime(): void {
  ws?.close();                                 // đóng khi xong; KHÔNG mở thường trực
  ws = null;
}
```

---

## §4 - Acceptance criteria

1. Mọi request backend đính `Authorization: Bearer <jwt SănDeal>`; grep `sync/**`: KHÔNG đọc `document.cookie`, KHÔNG header token/cookie sàn.
2. Body request (`SyncEnvelope`) chỉ chứa `OutboundPayload` + `clientTs`; KHÔNG khóa cookie/token/session/auth trong body (test introspection).
3. Không tồn tại đường gửi `CartReadMessage` thô bỏ qua `minimize()` -> queue (grep: mọi enqueue là OutboundPayload).
4. Hàng đợi bền: enqueue, mô phỏng SW kill (reset module), đọc lại hàng đợi từ `chrome.storage` còn nguyên item.
5. Gửi 2xx -> item bị `ack` (rời hàng đợi); gửi 5xx/lỗi mạng -> item ở lại để retry.
6. Backend 401 -> gọi `refreshJwt` rồi thử lại; item không bị vứt.
7. Thiếu JWT -> fail-closed: `flushQueue` dừng, metric `failClosed` tăng, item ở lại; KHÔNG gửi ẩn danh.
8. WSS chỉ mở khi gọi `openRealtime`; `closeRealtime` đóng; không có chỗ mở WSS top-level thường trực (grep).
9. JWT lưu ở `chrome.storage.session` (RAM-only), không `storage.local` (test/grep).
10. Đăng xuất -> JWT bị xóa khỏi session + WSS đóng; không gửi sau đăng xuất.
11. `npm test` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/sync-no-platform-token.test.ts
import { authedFetch } from "../src/sync/auth-bridge";

test("request đính JWT SănDeal, KHÔNG token/cookie sàn", async () => {
  setJwt("eyJ.sandeal.jwt");
  const captured = captureFetch();
  await authedFetch(SYNC_URL, { payload: { platform:"shopee", items:[], vouchers:[] }, clientTs: 1 });
  const h = captured.headers;
  expect(h["Authorization"]).toBe("Bearer eyJ.sandeal.jwt");
  const flat = JSON.stringify(captured).toLowerCase();
  expect(flat).not.toMatch(/cookie|spc_|set-cookie|x-bogus|mstoken/); // không token sàn
});

test("mã nguồn sync KHÔNG đọc document.cookie / header token sàn", async () => {
  for (const f of ["auth-bridge", "sender", "ws-client", "queue"]) {
    const src = await readFile(`src/sync/${f}.ts`, "utf8");
    expect(src).not.toMatch(/document\.cookie/);
  }
});
```

```ts
// extension/test/queue-persist.test.ts
test("hàng đợi sống qua SW kill (đọc lại từ storage)", async () => {
  globalThis.chrome = fakeChromeStorage();
  await queue.enqueue({ payload: { platform:"shopee", items:[{productId:"1",price:1,qty:1}], vouchers:[] }, clientTs: 1 });
  jest.resetModules();                                  // mô phỏng SW kill
  const { queue: q2 } = await import("../src/sync/queue");
  expect((await q2.peekAll()).length).toBe(1);          // KHÔNG mất
});

test("thiếu JWT → fail-closed, item ở lại", async () => {
  setJwt(undefined);
  await queue.enqueue(anyEnvelope());
  await flushQueue();
  expect((await queue.peekAll()).length).toBe(1);       // không gửi ẩn danh
});
```

```ts
// extension/test/ws-lifecycle.test.ts
test("WSS mở theo nhu cầu rồi đóng, không mở thường trực", async () => {
  openRealtime();
  expect(isWsOpen()).toBe(true);
  closeRealtime();
  expect(isWsOpen()).toBe(false);
  const swSrc = await readFile("src/background/service-worker.ts", "utf8");
  expect(swSrc).not.toMatch(/new WebSocket\(/);          // không mở WSS top-level thường trực
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `auth-bridge.ts` (lấy/refresh JWT SănDeal, đính Bearer, fail-closed thiếu JWT) -> `queue.ts` (hàng đợi bền trong chrome.storage) -> `sender.ts` (flush có retry/backoff, ack 2xx, 401->refresh) -> `ws-client.ts` (mở/đóng theo nhu cầu) -> nối `service-worker.ts` (pipeline-done -> enqueue; alarm tick -> flush; mở WSS khi cần realtime) -> tests. JWT ở storage.session; token sàn KHÔNG BAO GIỜ rời client (vẫn ở tab cho FR-EXT-002 mượn phiên). Body chỉ OutboundPayload đã sạch (FR-EXT-003), danh tính ở header.

---

## §7 - Phụ thuộc

- **FR-EXT-003** - cung cấp `OutboundPayload` đã sạch; lớp đồng bộ chỉ gửi cái này, không gửi thô.
- **FR-AUTH-002** - cấp JWT + refresh token SănDeal mà auth bridge đính vào request.
- **FR-EXT-001** - service worker + chrome.storage + alarm làm khung hàng đợi + flush; messaging nối pipeline.
- **FR-AUTH-003** - liên kết platform_account ẩn danh (ext_user_ref), KHÔNG token sàn - đồng bộ với ranh giới ở đây.
- **FR-TRUST-002 (downstream)** - chính sách local-first; lớp đồng bộ là điểm dữ liệu sạch rời máy.
- **NFR-EXT-001** - hàng đợi bền + WSS theo nhu cầu để không phá ràng buộc SW ephemeral.
- **FR-CART-002 / FR-AFFIL-004 (downstream, P2)** - tiêu thụ dữ liệu giỏ đồng bộ về backend.

---

## §8 - Payload ví dụ

### Request đồng bộ lên backend (header JWT SănDeal, body sạch)

```http
POST /v1/ext/sync HTTP/1.1
Host: api.sandeal.vn
Authorization: Bearer eyJ.<jwt SănDeal>...
Content-Type: application/json

{
  "payload": {
    "platform": "shopee",
    "items": [ { "productId": "90112", "price": 89000, "qty": 1 } ],
    "vouchers": [ { "code": "FREESHIPXTRA", "minSpend": 0, "discountText": "đến 15k" } ]
  },
  "clientTs": 1782000030000
}
```

### Item hàng đợi bền trong chrome.storage (ví dụ)

```json
{
  "sandeal:queue": [
    { "id": "q1", "env": { "payload": { "platform": "shopee", "items": [], "vouchers": [] }, "clientTs": 1782000030000 } }
  ]
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Gộp nhiều OutboundPayload thành batch trước khi gửi (giảm số request) - tối ưu ở slice sau khi đo lưu lượng.
- Kênh realtime cụ thể (WSS topic cho alert giá) - định nghĩa chi tiết khi FR-TRACK-004 (engine alert) tới.
- Chính sách hàng đợi pending lúc đăng xuất (giữ chờ vs xóa) - chốt cùng FR-COMPLY-003 (DSAR/xóa dữ liệu).

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Đính token/cookie sàn lên backend | sync-no-platform-token test | rò credential sàn -> chiếm tài khoản, PDPL | Chỉ JWT SănDeal ở header (DEC-EXT-23) |
| Gửi CartReadMessage thô | grep enqueue | bỏ qua tối thiểu hóa | Chỉ OutboundPayload vào queue (DEC-EXT-24) |
| Hàng đợi trong biến global | queue-persist test | mất dữ liệu khi SW kill | Hàng đợi bền chrome.storage (DEC-EXT-25) |
| Mất dữ liệu khi gửi lỗi | ack-only-2xx test | dữ liệu giỏ biến mất | Retry/backoff, ack chỉ khi 2xx (§1 #4) |
| Gửi ẩn danh khi thiếu JWT | fail-closed test | dữ liệu không gắn danh tính | Fail-closed, giữ hàng đợi (DEC-EXT-25) |
| WSS mở thường trực | ws-lifecycle test | lạm dụng tài nguyên, trái ephemeral | Mở theo nhu cầu, đóng khi xong (DEC-EXT-26) |
| JWT bền quá mức (storage.local) | grep/test | bề mặt rò token | Lưu storage.session RAM-only (DEC-EXT-27) |
| Gửi cleartext | review/test scheme | nghe lén dữ liệu | HTTPS/WSS bắt buộc (DEC-EXT-28) |
| 401 vứt item | sender 401 test | mất dữ liệu khi token hết hạn | Refresh rồi thử lại (§1 #4b) |
| Gửi sau đăng xuất | logout test | dữ liệu rời máy sau khi rút consent | Xóa JWT + đóng WSS + chặn gửi (§1 #12) |

---

## §11 - Ghi chú

- Ranh giới cứng nhất: request lên backend đính JWT SănDeal, token/cookie sàn KHÔNG BAO GIỜ rời client (ở lại tab cho FR-EXT-002 mượn phiên qua trình duyệt).
- Hàng đợi bền trong chrome.storage là hệ quả trực tiếp của SW ephemeral (NFR-EXT-001) - dữ liệu giỏ đọc được không mất khi SW ngủ giữa lúc gửi.
- Fail-closed khi thiếu JWT: thà giữ hàng đợi chờ còn hơn gửi dữ liệu ẩn danh không gắn danh tính/đường bảo mật.
- WSS mở theo nhu cầu rồi đóng - tài liệu nguồn cho phép WSS giữ SW sống nhưng "không lạm dụng" (§3.2); mở thường trực là chống tinh thần MV3.
- Body chỉ OutboundPayload đã sạch, danh tính ở header Authorization - tách bạch để audit (FR-TRUST-003) kiểm body thấy không credential.

---

*Hết FR-EXT-005. Status: ready_to_implement (mục tiêu audit 10/10).*
