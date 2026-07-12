---
id: FR-EXT-004
title: "Offscreen API cho DOM parsing/clipboard ngoài service worker + declarativeNetRequest thay webRequest blocking - tài liệu offscreen vòng đời ngắn, DNR static rules tối thiểu"
module: EXT
priority: SHOULD
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-EXT-001, FR-EXT-002, FR-EXT-003, FR-EXT-005, NFR-EXT-001]
depends_on: [FR-EXT-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (Ràng buộc MV3: dùng declarativeNetRequest thay webRequest blocking; dùng Offscreen API khi cần DOM parsing/clipboard ngoài service worker)"
  - "docs/... §3.1 (extension MV3 = content scripts per-sàn + service worker)"
source_decisions:
  - "DEC-EXT-18: service worker KHÔNG có DOM (không document/DOMParser tin cậy, không clipboard) -> mọi tác vụ cần DOM/clipboard chạy trong tài liệu offscreen (chrome.offscreen) với reason tường minh"
  - "DEC-EXT-19: tài liệu offscreen có vòng đời NGẮN - tạo khi cần, đóng ngay sau khi xong (closeOffscreenDocument); KHÔNG để mở thường trực (giữ SW sống + tốn RAM)"
  - "DEC-EXT-20: dùng declarativeNetRequest (DNR) cho mọi can thiệp request; CẤM webRequest blocking (đã bỏ ở MV3) - rule khai báo tĩnh, không chạy JS chặn từng request"
  - "DEC-EXT-21: DNR rule là tập TỐI THIỂU, khai static trong manifest (rule_resources); chỉ thêm rule có lý do rõ; KHÔNG dùng để chặn/sửa request sàn theo cách lạm dụng (né cam kết niềm tin §5.4)"
  - "DEC-EXT-22: offscreen chỉ parse HTML THÔ đã có sẵn (string) hoặc thao tác clipboard do người dùng kích hoạt; KHÔNG tự fetch trang sàn từ offscreen; dữ liệu ra vẫn đi qua pipeline tối thiểu hóa (FR-EXT-003)"

language: "TypeScript 5.x; Manifest V3; chrome.offscreen + chrome.declarativeNetRequest"
service: shopass/extension/
new_files:
  - extension/src/offscreen/offscreen.html
  - extension/src/offscreen/offscreen.ts
  - extension/src/offscreen/manager.ts
  - extension/src/dnr/rules.json
  - extension/src/dnr/dnr.ts
  - extension/test/offscreen-lifecycle.test.ts
  - extension/test/dnr-rules.test.ts
  - extension/test/offscreen-no-fetch.test.ts
modified_files:
  - extension/manifest.json                 # thêm permissions offscreen + declarativeNetRequest + declarative_net_request.rule_resources
  - extension/src/shared/types.ts           # thêm ParseDomRequest, ParseDomResult
allowed_tools:
  - file_read: extension/**
  - file_write: extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - dùng webRequest blocking (chrome.webRequest onBeforeRequest blocking) thay declarativeNetRequest (vi phạm DEC-EXT-20 - đã gỡ ở MV3, Web Store từ chối)
  - để tài liệu offscreen mở thường trực (vi phạm DEC-EXT-19 - giữ SW sống trái ràng buộc ephemeral, tốn RAM)
  - fetch trang sàn trực tiếp từ offscreen bỏ qua content script + pipeline (vi phạm DEC-EXT-22)
  - thêm DNR rule rộng/chặn-sửa request sàn theo cách lạm dụng (vi phạm DEC-EXT-21 - rủi ro niềm tin §5.4)

effort_hours: 5
sub_tasks:
  - "0.5h: manifest.json - thêm permissions [offscreen, declarativeNetRequest] + declarative_net_request.rule_resources trỏ rules.json"
  - "0.5h: offscreen.html + offscreen.ts - tài liệu offscreen tối giản, nhận message parse DOM, trả kết quả"
  - "1.0h: manager.ts - tạo offscreen có reason (DOM_SCRAPING/CLIPBOARD), gọi hasDocument trước khi tạo, đóng ngay sau khi xong"
  - "0.5h: rules.json - tập DNR rule tối thiểu (khai static), mỗi rule có lý do; dnr.ts wrapper kiểm rule"
  - "1.0h: offscreen-lifecycle.test.ts - tạo->dùng->đóng; không mở 2 lần; không để mở thường trực"
  - "1.0h: dnr-rules.test.ts - rule là static tối thiểu; không webRequest blocking; offscreen-no-fetch.test.ts - offscreen không tự fetch trang sàn"
  - "0.5h: nối manager với pipeline (kết quả parse -> FR-EXT-003 minimize) + metric offscreen_open/close"

risk_if_skipped: "Service worker MV3 không có DOM thật và không có clipboard - nếu cố parse HTML phức tạp hay thao tác clipboard ngay trong SW, code chạy sai (DOMParser trong SW không đáng tin) hoặc không chạy được. Offscreen API là cơ chế đúng để có một tài liệu có DOM ngoài SW. Làm sai vòng đời offscreen (để mở thường trực) thì lại phá ràng buộc SW ephemeral của NFR-EXT-001 và tốn RAM - đúng kiểu lỗi MV3 phổ biến. Về phía mạng, webRequest blocking đã bị gỡ ở MV3; nếu vẫn dùng, Chrome Web Store từ chối review và can thiệp request không hoạt động. declarativeNetRequest là API thay thế, nhưng nó cũng là bề mặt rủi ro niềm tin: nếu thêm rule rộng chặn/sửa request của sàn, extension trông giống công cụ thao túng (đúng điều Honey bị phạt, §5.4). FR này khóa cứng cả hai ranh giới - offscreen vòng đời ngắn và DNR tối thiểu - để extension vừa làm được việc cần DOM, vừa không vi phạm MV3 lẫn cam kết niềm tin."
---

## §1 - Mô tả (BCP-14 normative)

Extension **MUST** đặt mọi tác vụ cần DOM hoặc clipboard ngoài service worker vào một tài liệu offscreen vòng đời ngắn, và **MUST** dùng `declarativeNetRequest` thay cho webRequest blocking cho mọi can thiệp request. Hợp đồng:

1. Service worker **MUST NOT** giả định có DOM: không dựa `DOMParser`/`document` trong SW cho việc parse HTML phức tạp. Mọi parse DOM nặng hoặc thao tác clipboard **MUST** chạy trong tài liệu offscreen tạo qua `chrome.offscreen.createDocument` với `reason` tường minh (`DOM_SCRAPING` hoặc `CLIPBOARD`) (DEC-EXT-18).
2. `manager.ts` **MUST** kiểm `chrome.offscreen.hasDocument()` trước khi tạo - chỉ tồn tại tối đa một tài liệu offscreen tại một thời điểm (Chrome chỉ cho phép một).
3. Tài liệu offscreen **MUST** có vòng đời ngắn (DEC-EXT-19): tạo khi cần, đóng ngay bằng `chrome.offscreen.closeOffscreenDocument()` sau khi tác vụ xong. **MUST NOT** để offscreen mở thường trực - nó giữ SW sống và tốn RAM, trái ràng buộc ephemeral (NFR-EXT-001).
4. Mọi can thiệp request (chặn, chuyển hướng, sửa header) **MUST** dùng `chrome.declarativeNetRequest` (DEC-EXT-20). Mã **MUST NOT** dùng `chrome.webRequest` ở chế độ blocking (`onBeforeRequest` blocking) - API này đã bị gỡ ở MV3, Web Store từ chối.
5. DNR rule **MUST** là tập tối thiểu, khai báo static trong manifest qua `declarative_net_request.rule_resources` trỏ `rules.json` (DEC-EXT-21). Mỗi rule **MUST** có lý do ghi rõ. **MUST NOT** thêm rule rộng để chặn/sửa request của sàn theo cách lạm dụng (rủi ro niềm tin §5.4).
6. Tài liệu offscreen **MUST** chỉ parse HTML THÔ đã có sẵn (truyền vào dạng string qua message) hoặc thao tác clipboard do người dùng kích hoạt; **MUST NOT** tự `fetch` trang sàn từ offscreen (DEC-EXT-22). Việc đọc trang sàn là của content script (FR-EXT-002).
7. Dữ liệu kết quả từ offscreen **MUST** đi tiếp qua pipeline tối thiểu hóa (FR-EXT-003) trước khi rời client - offscreen không phải đường tắt bỏ qua allowlist.
8. `manifest.json` **MUST** thêm vào `permissions`: `"offscreen"`, `"declarativeNetRequest"` - chỉ khi FR này hiện thực hóa, đúng nguyên tắc quyền tối thiểu của FR-EXT-001.
9. Message giữa SW và offscreen **MUST** typed (`ParseDomRequest` / `ParseDomResult` trong `types.ts`), định tuyến rõ `target: "offscreen"` để không lẫn với message content script.
10. Khi `createOffscreenDocument` lỗi (đã có tài liệu / quyền thiếu) **MUST** xử lý lịch sự: tái dùng tài liệu đang mở hoặc trả lỗi rõ, KHÔNG retry vô hạn làm kẹt SW.
11. `dnr.ts` **MUST** cung cấp hàm kiểm số lượng + phạm vi rule lúc khởi động (sanity check) để test khẳng định rule không vượt tập tối thiểu.
12. Toàn bộ tác vụ offscreen **MUST** đặt giới hạn thời gian (tự đóng nếu quá ngưỡng, ví dụ 10 giây) để một parse treo không giữ tài liệu mở mãi.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao offscreen cho DOM/clipboard (DEC-EXT-18)?** Service worker MV3 không phải trang - nó không có `document` đầy đủ và `DOMParser` trong SW không đáng tin cho HTML phức tạp; clipboard cũng cần ngữ cảnh tài liệu. Offscreen API tạo một tài liệu ẩn có DOM thật, đúng nơi để parse HTML nặng hoặc đọc/ghi clipboard. Đây là cơ chế Chrome cung cấp thay cho thói quen cũ (background page có DOM) đã mất ở MV3.

**Vì sao offscreen vòng đời ngắn (DEC-EXT-19)?** Một tài liệu offscreen mở giữ service worker sống và tốn RAM. Nếu để mở thường trực, ta vô tình phá chính ràng buộc ephemeral mà NFR-EXT-001 yêu cầu. Quy ước "tạo khi cần, đóng ngay" giữ offscreen như một công cụ nhất thời, không phải một background page trá hình.

**Vì sao declarativeNetRequest, không webRequest blocking (DEC-EXT-20)?** MV3 đã gỡ webRequest blocking (lý do bảo mật + hiệu năng của Chrome). Rule khai báo của DNR được trình duyệt thực thi mà không chạy JS của extension trên từng request - nhanh hơn và minh bạch hơn. Code còn dùng webRequest blocking sẽ bị Web Store từ chối và đơn giản là không hoạt động.

**Vì sao DNR tối thiểu + static (DEC-EXT-21)?** DNR có thể chặn/chuyển hướng/sửa request - đúng loại quyền lực mà một extension "thao túng" lạm dụng. SănDeal hậu-Honey phải chứng minh mình không làm vậy. Rule khai static, đếm được, mỗi rule có lý do, cho phép audit (FR-TRUST-003) xác nhận extension không can thiệp request sàn ngoài mục đích đã công bố.

**Vì sao offscreen không tự fetch trang sàn (DEC-EXT-22)?** Nếu offscreen tự `fetch` trang sàn, nó trở thành một đường đọc thứ hai vòng qua content script - phá nguyên tắc "content script là đường đọc duy nhất, đọc trong tab đã đăng nhập". Offscreen chỉ nhận HTML thô đã có (do content script lấy) để parse, rồi trả kết quả. Mọi dữ liệu vẫn qua pipeline tối thiểu hóa.

**Vì sao giới hạn thời gian tác vụ offscreen (§1 #12)?** Một parse treo sẽ giữ tài liệu offscreen mở mãi - vừa tốn RAM, vừa giữ SW sống. Đặt trần thời gian rồi tự đóng biến lỗi treo thành lỗi đóng được, không để rò tài nguyên.

---

## §3 - Hợp đồng API / DDL

### manifest.json (bổ sung)

```jsonc
// extension/manifest.json (thêm)
{
  "permissions": ["storage", "alarms", "offscreen", "declarativeNetRequest"],
  "declarative_net_request": {
    "rule_resources": [
      { "id": "sandeal_rules", "enabled": true, "path": "dnr/rules.json" }
    ]
  }
}
```

### types.ts (message offscreen, typed)

```ts
// extension/src/shared/types.ts (thêm)
export interface ParseDomRequest {
  target: "offscreen";
  type: "PARSE_DOM";
  html: string;            // HTML THÔ đã có sẵn (content script lấy), KHÔNG phải URL
  platform: "shopee" | "tiktok" | "lazada";
}
export interface ParseDomResult {
  type: "PARSE_DOM_RESULT";
  items: Array<{ productId: string; price: number; qty: number }>;
}
```

### manager.ts (vòng đời ngắn: kiểm -> tạo -> dùng -> đóng)

```ts
// extension/src/offscreen/manager.ts
const OFFSCREEN_PATH = "offscreen/offscreen.html";

export async function parseDomOffscreen(req: ParseDomRequest): Promise<ParseDomResult> {
  if (!(await chrome.offscreen.hasDocument())) {
    await chrome.offscreen.createDocument({
      url: OFFSCREEN_PATH,
      reasons: ["DOM_SCRAPING"],     // reason tường minh
      justification: "Parse HTML giỏ hàng đã render ngoài service worker"
    });
  }
  try {
    const res = await sendWithTimeout(req, 10_000); // trần 10s
    return res as ParseDomResult;
  } finally {
    await chrome.offscreen.closeOffscreenDocument(); // đóng NGAY sau khi xong
  }
}
```

### rules.json (DNR static, tối thiểu)

```json
[]
```

---

## §4 - Acceptance criteria

1. `manifest.json` có `permissions` chứa `offscreen` + `declarativeNetRequest`, và `declarative_net_request.rule_resources` trỏ `dnr/rules.json`.
2. Grep `src/**`: KHÔNG có `chrome.webRequest` ở chế độ blocking (không `onBeforeRequest` với `["blocking"]`).
3. `manager.ts` gọi `hasDocument()` trước khi `createDocument`; không tạo khi đã có tài liệu (test).
4. Sau khi parse xong, `closeOffscreenDocument` được gọi (test khẳng định offscreen được đóng, không để mở).
5. Tài liệu offscreen tạo với `reasons` tường minh (`DOM_SCRAPING`/`CLIPBOARD`) + `justification` không rỗng.
6. Grep `offscreen/**`: KHÔNG có `fetch(` tới domain sàn (offscreen không tự đọc trang sàn).
7. Kết quả parse offscreen đi qua `minimize()` (FR-EXT-003) trước khi rời client (test đường dẫn).
8. DNR rule trong `rules.json` là tập tối thiểu (test đếm + phạm vi); không có rule rộng chặn/sửa request sàn lạm dụng.
9. Tác vụ offscreen quá 10s -> tự đóng tài liệu (test timeout), không treo SW.
10. `createDocument` lỗi (đã có tài liệu) -> tái dùng/báo lỗi rõ, không retry vô hạn.
11. Message SW <-> offscreen là `ParseDomRequest`/`ParseDomResult` typed với `target: "offscreen"`; một message thiếu `target` hoặc sai kiểu bị `tsc` bắt lúc biên dịch (§1 #9).
12. `npm test` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/offscreen-lifecycle.test.ts
import { parseDomOffscreen } from "../src/offscreen/manager";

test("tạo offscreen rồi đóng ngay sau khi xong", async () => {
  const calls: string[] = [];
  globalThis.chrome = fakeOffscreen(calls); // ghi lại create/close
  await parseDomOffscreen({ target:"offscreen", type:"PARSE_DOM", html:"<div></div>", platform:"shopee" });
  expect(calls).toContain("create");
  expect(calls).toContain("close");        // đóng NGAY, không để mở thường trực
});

test("không tạo offscreen mới khi đã có tài liệu", async () => {
  const calls: string[] = [];
  globalThis.chrome = fakeOffscreen(calls, /*hasDocument*/ true);
  await parseDomOffscreen({ target:"offscreen", type:"PARSE_DOM", html:"<div></div>", platform:"shopee" });
  expect(calls.filter(c => c === "create")).toHaveLength(0); // tái dùng
});
```

```ts
// extension/test/dnr-rules.test.ts
import rules from "../src/dnr/rules.json";

test("DNR rule là tập tối thiểu, không webRequest blocking", async () => {
  expect(Array.isArray(rules)).toBe(true);
  expect(rules.length).toBeLessThanOrEqual(5);          // tối thiểu
  const src = await readFile("src/dnr/dnr.ts", "utf8");
  expect(src).not.toMatch(/webRequest/);                // không webRequest blocking
});
```

```ts
// extension/test/offscreen-no-fetch.test.ts
test("offscreen KHÔNG tự fetch trang sàn", async () => {
  for (const f of ["offscreen", "manager"]) {
    const src = await readFile(`src/offscreen/${f}.ts`, "utf8");
    expect(src).not.toMatch(/fetch\(["'`]https:\/\/(shopee|tiktok|lazada)/);
  }
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: cập nhật `manifest.json` (permissions offscreen + DNR + rule_resources) + `types.ts` (ParseDomRequest/Result) -> `offscreen.html` + `offscreen.ts` (tài liệu nhận HTML thô, parse, trả) -> `manager.ts` (kiểm hasDocument -> create reason -> sendWithTimeout -> close) -> `rules.json` + `dnr.ts` (rule static tối thiểu + sanity check) -> nối kết quả parse vào `minimize()` (FR-EXT-003) -> tests. Offscreen chỉ là công cụ nhất thời cho DOM/clipboard; content script vẫn là đường đọc trang sàn duy nhất. DNR rule khai static để Web Store và audit kiểm được.

---

## §7 - Phụ thuộc

- **FR-EXT-001** - scaffold MV3 + service worker + manifest + messaging làm khung cho offscreen/DNR.
- **FR-EXT-002** - content script lấy HTML thô trang sàn rồi giao cho offscreen parse (offscreen không tự đọc).
- **FR-EXT-003 (downstream)** - kết quả parse offscreen đi qua pipeline tối thiểu hóa trước khi rời client.
- **FR-EXT-005 (downstream)** - lớp đồng bộ gửi payload đã sạch; DNR/offscreen không phải đường tắt ra mạng.
- **NFR-EXT-001** - offscreen vòng đời ngắn để không phá ràng buộc SW ephemeral.
- Nền tảng: Chrome/Chromium MV3 >=120 (chrome.offscreen, chrome.declarativeNetRequest).

---

## §8 - Payload ví dụ

### SW yêu cầu offscreen parse HTML thô (typed)

```json
{
  "target": "offscreen",
  "type": "PARSE_DOM",
  "platform": "shopee",
  "html": "<div class='cart-item'>...HTML đã render do content script lấy...</div>"
}
```

### Offscreen trả kết quả (đi tiếp qua minimize)

```json
{
  "type": "PARSE_DOM_RESULT",
  "items": [ { "productId": "90112", "price": 89000, "qty": 1 } ]
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Có cần offscreen cho audio/iframe ngoài DOM_SCRAPING/CLIPBOARD hay không - chỉ thêm reason khi có FR tiêu thụ.
- DNR rule cụ thể (nếu có nhu cầu chặn tracker bên thứ ba của chính extension) - thêm khi có yêu cầu rõ + review niềm tin.
- Tái dùng một offscreen cho nhiều parse liên tiếp (gộp) thay vì tạo/đóng mỗi lần - tối ưu sau nếu đo thấy chi phí tạo/đóng đáng kể.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Parse DOM nặng trong SW | kết quả sai không tái hiện | parse hỏng thầm lặng | Đẩy vào offscreen có DOM thật (DEC-EXT-18) |
| Offscreen để mở thường trực | lifecycle test (thiếu close) | giữ SW sống + tốn RAM | Đóng ngay sau khi xong (DEC-EXT-19) |
| Tạo 2 tài liệu offscreen | createDocument lỗi | exception | Kiểm hasDocument trước (§1 #2) |
| webRequest blocking | grep test + Web Store reject | can thiệp request không chạy | Dùng declarativeNetRequest (DEC-EXT-20) |
| DNR rule rộng chặn/sửa request sàn | dnr-rules test + review | nghi thao túng (§5.4) | Rule static tối thiểu, có lý do (DEC-EXT-21) |
| Offscreen tự fetch trang sàn | offscreen-no-fetch test | đường đọc thứ hai vòng pipeline | Chỉ nhận HTML thô, không fetch (DEC-EXT-22) |
| Tác vụ offscreen treo | timeout test | tài liệu mở mãi, rò RAM | Trần 10s rồi tự đóng (§1 #12) |
| Kết quả parse bỏ qua minimize | đường-dẫn test | rò trường thừa | Mọi kết quả qua FR-EXT-003 (§1 #7) |
| Quyền offscreen/DNR thiếu trong manifest | manifest check | API không chạy | Thêm permissions tối thiểu (§1 #8) |

---

## §11 - Ghi chú

- Offscreen là cách MV3 thay background-page-có-DOM: dùng cho parse HTML nặng và clipboard, nhưng phải vòng đời ngắn để không phá ephemeral.
- declarativeNetRequest thay webRequest blocking là bắt buộc ở MV3; DNR rule khai static + tối thiểu để Web Store và audit (FR-TRUST-003) kiểm được.
- Offscreen KHÔNG tự đọc trang sàn - content script (FR-EXT-002) là đường đọc duy nhất; offscreen chỉ parse HTML thô đã có.
- Mọi dữ liệu ra từ offscreen vẫn qua pipeline tối thiểu hóa (FR-EXT-003) - không có đường tắt bỏ qua allowlist.
- Khả năng can thiệp request (DNR) là quyền lực nhạy cảm hậu-Honey; giữ nó tối thiểu và minh bạch là mắt xích niềm tin (§5.4).

---

*Hết FR-EXT-004. Status: ready_to_implement (mục tiêu audit 10/10).*
