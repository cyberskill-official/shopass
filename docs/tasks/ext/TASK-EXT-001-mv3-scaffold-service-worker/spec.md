---
id: TASK-EXT-001
title: "Scaffold extension Manifest V3 + vòng đời service worker ephemeral - state trong chrome.storage, chrome.alarms >=30s (KHÔNG setInterval), host_permissions per-domain sàn"
module: EXT
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-EXT-002, TASK-EXT-003, TASK-EXT-004, TASK-EXT-005, TASK-EXT-006, TASK-TRUST-001, NFR-EXT-001]
depends_on: []
blocks: [TASK-EXT-002, TASK-EXT-004, TASK-EXT-006, TASK-TRUST-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (Ràng buộc Manifest V3 & cách kiến trúc vòng quanh)"
  - "docs/... §3.1 (Extension MV3 = content scripts per-sàn + service worker)"
source_decisions:
  - "DEC-EXT-01: extension dùng Manifest V3; điểm nền là một service worker (background.service_worker), KHÔNG dùng background page persistent"
  - "DEC-EXT-02: service worker ephemeral (bị kill ~30s không hoạt động, hoặc khi 1 sự kiện chạy >5 phút / fetch không phản hồi >30s) -> KHÔNG giữ state trong biến global; mọi state bền lưu chrome.storage"
  - "DEC-EXT-03: polling định kỳ dùng chrome.alarms với chu kỳ tối thiểu 30 giây (Chrome 120+); CẤM setInterval/setTimeout dài trong service worker"
  - "DEC-EXT-04: host_permissions khai báo TƯỜNG MINH từng domain sàn (shopee.vn, tiktok.com, lazada.vn...); KHÔNG dùng <all_urls>"
  - "DEC-EXT-05: tác vụ nặng/dài đẩy lên backend; extension chỉ là 'đầu đọc' nhẹ (heavy work off-device, TASK-EXT-003/005)"
  - "DEC-EXT-06: TypeScript + bundler (esbuild/vite) build ra MV3 bundle; mã nguồn mở (TASK-TRUST-001) nên ưu tiên build tái lập"

language: "TypeScript 5.x; Manifest V3; Chrome MV3 service worker (chrome.* APIs); bundler esbuild"
service: shopass/extension/
new_files:
  - extension/manifest.json
  - extension/src/background/service-worker.ts
  - extension/src/background/lifecycle.ts
  - extension/src/background/alarms.ts
  - extension/src/shared/storage.ts
  - extension/src/shared/messaging.ts
  - extension/src/shared/types.ts
  - extension/build.mjs
  - extension/tsconfig.json
  - extension/test/lifecycle.test.ts
  - extension/test/alarms.test.ts
  - extension/test/manifest.test.ts
modified_files: []
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - giữ state trong biến module-global của service worker (vi phạm DEC-EXT-02 - mất state khi SW bị kill)
  - dùng setInterval/setTimeout dài thay chrome.alarms cho polling (vi phạm DEC-EXT-03 - không sống qua chu kỳ kill)
  - khai báo host_permissions <all_urls> hoặc wildcard rộng (vi phạm DEC-EXT-04 - bị Chrome Web Store review từ chối + lộ quyền thừa)
  - dùng manifest_version 2 hoặc background page persistent (vi phạm DEC-EXT-01)

effort_hours: 8
sub_tasks:
  - "1.0h: manifest.json - manifest_version 3, background.service_worker (type module), host_permissions per-domain, permissions tối thiểu (storage, alarms)"
  - "1.0h: storage.ts - wrapper chrome.storage.local typed (get/set/remove); mọi state đi qua đây"
  - "1.0h: lifecycle.ts - onInstalled/onStartup khởi tạo state vào storage; KHÔNG dựa biến global; rehydrate khi SW wake"
  - "1.0h: alarms.ts - đăng ký chrome.alarms (periodInMinutes >=0.5), onAlarm handler đọc state từ storage"
  - "0.5h: messaging.ts - kiểu thông điệp content<->SW (typed) + helper sendMessage/onMessage"
  - "0.5h: build.mjs + tsconfig.json - esbuild bundle ra dist/, build tái lập (TASK-TRUST-001)"
  - "1.5h: lifecycle.test.ts - state survive 'kill' (mô phỏng SW restart đọc lại storage); KHÔNG có biến global mang state"
  - "1.0h: alarms.test.ts - alarm chu kỳ <30s bị từ chối/kẹp về 30s; handler chạy từ storage, không setInterval"
  - "0.5h: manifest.test.ts - manifest_version=3, không <all_urls>, có service_worker, permissions là tập tối thiểu"

risk_if_skipped: "Đây là nền của toàn bộ module EXT - mọi content script (Shopee/TikTok/Lazada), pipeline tối thiểu hóa, đồng bộ backend và UI consent đều nạp trên scaffold này. Làm sai vòng đời MV3 là lỗi phổ biến nhất khi viết extension: nếu giữ state trong biến global, service worker bị Chrome kill sau ~30s và toàn bộ trạng thái (hàng đợi, phiên đọc giỏ) biến mất giữa chừng -> đọc giỏ hàng lỗi không tái hiện được. Nếu dùng setInterval thay chrome.alarms, polling chết khi SW ngủ. Nếu khai <all_urls>, Chrome Web Store review từ chối và người dùng thấy quyền thừa (extension đọc cookie sàn vốn đã dễ bị nghi malware, §5.4). Scaffold đúng MV3 ngay từ đầu là rẻ; sửa sau khi đã chồng 5 task lên trên là đắt."
---

## §1 - Mô tả (BCP-14 normative)

Extension SănDeal **MUST** dựng trên Manifest V3 với một service worker ephemeral đúng vòng đời Chrome, không giữ state trong biến global, dùng `chrome.alarms` cho polling và khai báo `host_permissions` tường minh từng domain sàn. Hợp đồng:

1. `manifest.json` **MUST** đặt `manifest_version: 3` và khai báo `background.service_worker` (kèm `type: "module"`) - KHÔNG dùng `background.page` hay background page persistent (DEC-EXT-01).
2. Service worker **MUST** coi mình là ephemeral (DEC-EXT-02): Chrome kill SW sau ~30 giây không hoạt động, hoặc khi một sự kiện chạy quá 5 phút, hoặc khi một `fetch()` không phản hồi quá 30 giây. Mã **MUST NOT** giữ bất kỳ state có ý nghĩa nào trong biến module-global; mọi state bền **MUST** lưu vào `chrome.storage.local` (hoặc `.session` cho state nhạy cảm theo phiên).
3. Module `storage.ts` **MUST** là điểm truy cập state duy nhất: bọc `chrome.storage` thành API typed `get/set/remove`. Các module khác **MUST** đọc/ghi state qua đây, không truy cập biến toàn cục.
4. Khi SW thức dậy (wake) do một sự kiện, handler **MUST** rehydrate state từ `chrome.storage` trước khi xử lý - không giả định biến trong bộ nhớ còn sống từ lần chạy trước.
5. Polling định kỳ **MUST** dùng `chrome.alarms` với `periodInMinutes >= 0.5` (tức >=30 giây, ràng buộc Chrome 120+) (DEC-EXT-03). Mã **MUST NOT** dùng `setInterval` hay `setTimeout` chu kỳ dài trong service worker để lập lịch - chúng không sống qua chu kỳ kill.
6. `manifest.json` **MUST** khai báo `host_permissions` tường minh cho từng domain sàn được hỗ trợ (ví dụ `https://shopee.vn/*`; các sàn khác thêm khi task tương ứng tới) - KHÔNG dùng `<all_urls>` hay wildcard rộng (DEC-EXT-04).
7. `permissions` trong manifest **MUST** là tập tối thiểu cần cho slice 1: `storage`, `alarms`. Quyền khác (ví dụ `offscreen`, `declarativeNetRequest`) chỉ thêm khi task tương ứng (TASK-EXT-004) tới.
8. Tác vụ nặng hoặc dài (đọc nhiều tab, đồng bộ khối lượng lớn) **MUST** đẩy lên backend (DEC-EXT-05); extension chỉ giữ vai trò "đầu đọc" nhẹ. Service worker **MUST NOT** chạy vòng lặp tính toán dài làm vượt ngưỡng 5 phút.
9. `messaging.ts` **MUST** định nghĩa kiểu thông điệp giữa content script và service worker (typed `Message` discriminated union) và helper `sendMessage`/`onMessage` bọc `chrome.runtime` để hai bên giao tiếp an toàn kiểu.
10. Build **MUST** tái lập được (DEC-EXT-06): `build.mjs` (esbuild) sinh `dist/` xác định từ mã nguồn; phục vụ yêu cầu mã nguồn mở + reproducible build của TASK-TRUST-001.
11. Service worker **MUST** đăng ký listener ở top-level đồng bộ (`chrome.runtime.onInstalled`, `chrome.alarms.onAlarm`, `chrome.runtime.onMessage`) - KHÔNG đăng ký listener bên trong callback bất đồng bộ (MV3 yêu cầu listener đăng ký synchronously lúc SW khởi động để nhận sự kiện wake).
12. Extension **MUST** chạy trên Chrome (và Chromium MV3) tối thiểu phiên bản 120 (mốc `chrome.alarms` 30 giây); manifest **SHOULD** đặt `minimum_chrome_version: "120"`.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao service worker ephemeral, state trong chrome.storage (DEC-EXT-02)?** MV3 thay background page persistent bằng service worker bị kill chủ động để tiết kiệm tài nguyên. Đây là cái bẫy số một khi viết extension: code chạy đúng trong lúc dev (SW còn sống), rồi hỏng trong thực tế khi SW bị kill và biến global mất sạch. Quy ước "state luôn ở chrome.storage, biến global chỉ là cache nhất thời" loại bỏ cả lớp bug không tái hiện được. Mọi task sau (đọc giỏ, đồng bộ) dựa vào nền này.

**Vì sao chrome.alarms >=30s, không setInterval (DEC-EXT-03)?** `setInterval` sống trong bộ nhớ SW; khi SW ngủ, timer chết, polling dừng âm thầm. `chrome.alarms` được Chrome quản lý ngoài SW và đánh thức SW khi tới hạn - đó là cơ chế hẹn giờ duy nhất sống qua chu kỳ kill. Trần 30 giây (Chrome 120+) là ràng buộc nền tảng, không phải lựa chọn; mã phải tôn trọng nó.

**Vì sao host_permissions per-domain, không <all_urls> (DEC-EXT-04)?** Extension này đọc ngữ cảnh đăng nhập sàn (cookie phiên first-party) - quyền nhạy cảm. Khai báo từng domain (shopee.vn...) làm rõ với Chrome Web Store reviewer và với người dùng rằng extension chỉ chạm các sàn đã công bố. `<all_urls>` kích hoạt cờ review nghiêm ngặt và làm người dùng nghi malware (§5.4) - đúng điều SănDeal phải né hậu-Honey.

**Vì sao đẩy việc nặng lên backend (DEC-EXT-05)?** SW bị kill nếu một sự kiện chạy >5 phút. Mọi việc tốn thời gian (đồng bộ lớn, tính toán) phải nằm ở backend; extension chỉ trích xuất dữ liệu đã render rồi gửi đi. Ranh giới này vừa hợp ràng buộc MV3, vừa hợp triết lý "đầu đọc nhẹ + xử lý phía server".

**Vì sao listener đăng ký top-level đồng bộ (§1 #11)?** MV3 chỉ giao sự kiện wake cho các listener đã đăng ký synchronously khi SW khởi động. Đăng ký trong một callback `await` làm SW lỡ sự kiện đánh thức nó - một lỗi tinh vi khiến alarm/message thỉnh thoảng không chạy. Quy ước top-level loại bỏ lớp lỗi này.

**Vì sao build tái lập (DEC-EXT-06)?** TASK-TRUST-001 cam kết mã nguồn mở + reproducible build để chứng minh extension không gửi cookie/mật khẩu. Build xác định cho phép bên thứ ba dựng lại bundle từ nguồn và so khớp với bản trên Chrome Web Store. Đặt nền build sạch ngay từ scaffold rẻ hơn nhiều so với việc gỡ rối sau.

---

## §3 - Hợp đồng API / DDL

### manifest.json (lõi)

```jsonc
// extension/manifest.json
{
  "manifest_version": 3,
  "name": "SănDeal",
  "version": "0.1.0",
  "minimum_chrome_version": "120",
  "background": {
    "service_worker": "background/service-worker.js",
    "type": "module"
  },
  "permissions": ["storage", "alarms"],
  "host_permissions": [
    "https://shopee.vn/*"
  ],
  "action": { "default_title": "SănDeal" }
}
```

### storage.ts (điểm truy cập state duy nhất)

```ts
// extension/src/shared/storage.ts
export interface ExtState {
  schemaVersion: number;
  installedAt: number;        // epoch ms
  lastSyncAt?: number;
  pendingReads: string[];     // hàng đợi nhẹ, KHÔNG token/cookie
}

const KEY = "sandeal:state";

export async function getState(): Promise<ExtState> {
  const obj = await chrome.storage.local.get(KEY);
  return (obj[KEY] as ExtState) ?? defaultState();
}

export async function setState(next: ExtState): Promise<void> {
  await chrome.storage.local.set({ [KEY]: next });
}

export function defaultState(): ExtState {
  return { schemaVersion: 1, installedAt: Date.now(), pendingReads: [] };
}
```

### lifecycle.ts + alarms.ts (top-level, đồng bộ)

```ts
// extension/src/background/service-worker.ts  (entrypoint)
import { onInstalled, onStartup } from "./lifecycle";
import { registerAlarms, onAlarm } from "./alarms";
import { onMessage } from "../shared/messaging";

// MUST: listener đăng ký top-level đồng bộ (MV3 wake requirement)
chrome.runtime.onInstalled.addListener(onInstalled);
chrome.runtime.onStartup.addListener(onStartup);
chrome.alarms.onAlarm.addListener(onAlarm);
chrome.runtime.onMessage.addListener(onMessage);

registerAlarms(); // tạo alarm periodInMinutes >= 0.5
```

```ts
// extension/src/background/alarms.ts
export const TICK = "sandeal:tick";

export function registerAlarms(): void {
  // 0.5 phút = 30 giây, trần Chrome 120+. KHÔNG setInterval.
  chrome.alarms.create(TICK, { periodInMinutes: 0.5 });
}

export async function onAlarm(alarm: chrome.alarms.Alarm): Promise<void> {
  if (alarm.name !== TICK) return;
  const state = await getState();   // rehydrate từ storage, KHÔNG biến global
  // ... xử lý nhẹ; việc nặng đẩy backend (DEC-EXT-05)
  await setState({ ...state, lastSyncAt: Date.now() });
}
```

---

## §4 - Acceptance criteria

1. `manifest.json` có `manifest_version: 3` và `background.service_worker` (type module); KHÔNG có `background.page`.
2. `host_permissions` chỉ chứa domain sàn tường minh (`https://shopee.vn/*`); KHÔNG chứa `<all_urls>` hay `http(s)://*/*`.
3. `permissions` đúng tập tối thiểu `["storage","alarms"]` cho slice 1.
4. Đăng ký alarm với `periodInMinutes < 0.5` -> Chrome kẹp về 30s; mã KHÔNG dựa chu kỳ nhỏ hơn (test xác nhận không có alarm <30s được kỳ vọng chạy đúng chu kỳ nhỏ).
5. Mô phỏng SW bị kill rồi wake (khởi tạo lại module): state đọc lại từ `chrome.storage` khớp giá trị đã ghi - KHÔNG mất.
6. Grep mã `src/background/**`: KHÔNG có biến module-global mang state nghiệp vụ (chỉ hằng số/handler).
7. Grep mã service worker: KHÔNG có `setInterval(` và KHÔNG có `setTimeout(` với chu kỳ lập lịch dài.
8. Listener (`onInstalled`, `onAlarm`, `onMessage`) đăng ký ở top-level của entrypoint, KHÔNG bên trong callback `async`/`await`.
9. `build.mjs` chạy 2 lần trên cùng nguồn -> byte `dist/` giống nhau (build xác định/tái lập).
10. `minimum_chrome_version` = "120" (mốc alarms 30s).
11. `npm test` xanh; `tsc --noEmit` không lỗi kiểu.
12. `messaging.ts` định nghĩa `Message` là discriminated union (theo `type`) và export helper `sendMessage`/`onMessage` bọc `chrome.runtime`; một thông điệp sai kiểu (thiếu `type` hoặc payload sai) bị `tsc` bắt lúc biên dịch (§1 #9).

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/lifecycle.test.ts
import { getState, setState, defaultState } from "../src/shared/storage";
import { fakeChromeStorage } from "./helpers";

test("state survive SW kill (rehydrate từ storage)", async () => {
  globalThis.chrome = fakeChromeStorage();      // storage backing bền
  await setState({ ...defaultState(), lastSyncAt: 111 });

  // mô phỏng SW kill: vứt mọi biến module, đọc lại sạch từ storage
  jest.resetModules();
  const { getState: getState2 } = await import("../src/shared/storage");
  const s = await getState2();
  expect(s.lastSyncAt).toBe(111);               // KHÔNG mất qua "kill"
});

test("không có state trong biến global", async () => {
  const src = await readFile("src/background/service-worker.ts", "utf8");
  // entrypoint chỉ đăng ký listener + tạo alarm, không khai biến state
  expect(src).not.toMatch(/let\s+\w+State\s*=/);
});
```

```ts
// extension/test/alarms.test.ts
import { registerAlarms, TICK } from "../src/background/alarms";

test("alarm chu kỳ >= 30s, không setInterval", async () => {
  const created: any[] = [];
  globalThis.chrome = { alarms: { create: (n: string, o: any) => created.push({ n, o }) } } as any;
  registerAlarms();
  expect(created[0].n).toBe(TICK);
  expect(created[0].o.periodInMinutes).toBeGreaterThanOrEqual(0.5); // >=30s

  const swSrc = await readFile("src/background/alarms.ts", "utf8");
  expect(swSrc).not.toMatch(/setInterval\(/);
});
```

```ts
// extension/test/manifest.test.ts
test("manifest MV3, không <all_urls>, permissions tối thiểu", async () => {
  const m = JSON.parse(await readFile("manifest.json", "utf8"));
  expect(m.manifest_version).toBe(3);
  expect(m.background.service_worker).toBeTruthy();
  expect(m.background.page).toBeUndefined();
  expect(JSON.stringify(m.host_permissions)).not.toMatch(/all_urls|:\/\/\*\/\*/);
  expect(m.permissions.sort()).toEqual(["alarms", "storage"]);
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `manifest.json` + `tsconfig.json` + `build.mjs` (khung build) -> `storage.ts` (điểm state) -> `lifecycle.ts` + `alarms.ts` + `messaging.ts` -> `service-worker.ts` (entrypoint nối listener top-level) -> tests. Bundler esbuild emit ESM module cho service worker (`type: "module"`). Tránh mọi top-level `await` trong entrypoint service worker (làm trễ đăng ký listener). Mọi TASK-EXT sau nạp content script / offscreen / sync lên đúng scaffold này.

---

## §7 - Phụ thuộc

- **depends_on: []** - đây là task gốc của module EXT, không phụ thuộc task khác.
- **TASK-EXT-002 (downstream)** - content script Shopee nạp trên scaffold + messaging này.
- **TASK-EXT-003/004/005/006 (downstream)** - pipeline tối thiểu hóa, offscreen, đồng bộ backend, UI consent đều dựa scaffold.
- **TASK-TRUST-001 (downstream)** - publish mã nguồn mở + reproducible build dùng `build.mjs` ở đây.
- **NFR-EXT-001** - ràng buộc MV3 (ephemeral, alarms >=30s, no global state) mà task này hiện thực hóa.
- Nền tảng: Chrome/Chromium MV3 >=120; TypeScript 5.x; esbuild.

---

## §8 - Payload ví dụ

### Thông điệp content -> service worker (typed)

```ts
// content script gửi báo cáo đọc giỏ (KHÔNG cookie/token)
chrome.runtime.sendMessage({
  type: "CART_READ",
  platform: "shopee",
  items: [{ productId: "90112", price: 89000, qty: 1 }]
} satisfies Message);
```

### State trong chrome.storage.local (ví dụ)

```json
{
  "sandeal:state": {
    "schemaVersion": 1,
    "installedAt": 1782000000000,
    "lastSyncAt": 1782000030000,
    "pendingReads": []
  }
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Có dùng `chrome.storage.session` cho state phiên-nhạy-cảm (RAM-only) hay không - quyết ở TASK-EXT-003 khi định nghĩa dữ liệu tối thiểu hóa.
- Firefox MV3 (service worker khác Chrome) - hoãn tới khi có nhu cầu đa trình duyệt; slice 1 nhắm Chrome/Chromium.
- Cập nhật host_permissions cho tiktok.com / lazada.vn - thêm ở TASK-EXT-007 / TASK-EXT-008 (P2), không khai sớm quyền chưa dùng.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Giữ state trong biến global | test "không có state global" + lỗi runtime không tái hiện | state mất khi SW kill | Đẩy mọi state qua storage.ts (DEC-EXT-02) |
| Dùng setInterval polling | grep test + polling chết khi SW ngủ | mất nhịp đọc/đồng bộ | Thay bằng chrome.alarms >=30s (DEC-EXT-03) |
| host_permissions <all_urls> | manifest.test + Web Store review từ chối | reject + nghi malware (§5.4) | Khai từng domain sàn (DEC-EXT-04) |
| Listener đăng ký trong async callback | sự kiện wake thỉnh thoảng lỡ | alarm/message không chạy | Đăng ký top-level đồng bộ (§1 #11) |
| Sự kiện chạy >5 phút | SW bị Chrome kill giữa chừng | tác vụ dở dang | Đẩy việc nặng lên backend (DEC-EXT-05) |
| fetch không phản hồi >30s | SW kill, request treo | mất kết quả | Timeout fetch < 30s; retry phía backend |
| alarm chu kỳ <30s | Chrome kẹp âm thầm về 30s | nhịp thực != kỳ vọng | Thiết kế quanh trần 30s; không kỳ vọng nhỏ hơn |
| Build không xác định | dist khác giữa 2 lần build | TASK-TRUST-001 reproducible vỡ | Khóa version + esbuild xác định |
| Thừa permission trong manifest | manifest.test | quyền thừa, review chậm | Giữ tập tối thiểu, thêm theo task |

---

## §11 - Ghi chú

- Đây là scaffold nền của toàn module EXT; ba ràng buộc MV3 (SW ephemeral / alarms >=30s / no global state) là bất biến cho mọi task sau, được chốt cứng ở NFR-EXT-001.
- "State luôn ở chrome.storage" là quy ước rẻ nhưng loại bỏ lớp bug không tái hiện được lớn nhất của extension MV3.
- host_permissions per-domain vừa hợp lệ Web Store vừa là tín hiệu niềm tin (extension chỉ chạm sàn đã công bố) - mắt xích chiến lược hậu-Honey (§5.4).
- Việc nặng off-device là ranh giới kiến trúc xuyên suốt: extension đọc nhẹ, backend gánh nặng (TASK-EXT-003/005, TASK-SCRAPE-*).
- Reproducible build ở đây là tiền đề cho cam kết minh bạch TASK-TRUST-001.

---

*Hết TASK-EXT-001. Status: ready_to_implement (mục tiêu audit 10/10).*
