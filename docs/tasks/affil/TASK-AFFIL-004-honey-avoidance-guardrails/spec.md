---
id: TASK-AFFIL-004
title: "Guardrails né Honey - KHÔNG cookie-stuffing/dropping/pop-under/auto-redirect/forced-install; bắt buộc hành động người dùng + disclosure; tuân Chrome Web Store policy (cập nhật 3/2025, thực thi 10/06/2025)"
module: AFFIL
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-AFFIL-001, TASK-AFFIL-002, TASK-EXT-003, TASK-TRUST-001, TASK-TRUST-002, NFR-AFFIL-001]
depends_on: [TASK-AFFIL-002, TASK-EXT-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §4.2 (bài học PayPal Honey; Chrome Web Store policy 3/2025 thực thi 10/06/2025; cấm cookie dropping/pop-under/auto-redirect/forced install)"
  - "docs/... §5.4 (trust: minh bạch, local-first), §5.2 (rủi ro Chrome gỡ extension kiểu Honey)"
source_decisions:
  - "DEC-AFFIL-16: guardrail là một module kiểm tra (lint + runtime assert) khẳng định extension/affil KHÔNG có cookie-stuffing, cookie-dropping, pop-under, auto-redirect, forced-install"
  - "DEC-AFFIL-17: mọi điều hướng affiliate phải bắt nguồn từ user gesture (bấm) + đi qua TASK-AFFIL-002 với user_initiated=true + có disclosure hiển thị"
  - "DEC-AFFIL-18: extension KHÔNG khai báo permission/host cho phép sửa cookie domain sàn; manifest bị kiểm để chặn webRequest sửa cookie/redirect nền"
  - "DEC-AFFIL-19: guardrail chạy như test bắt buộc (CI gate) - vi phạm làm đỏ build; đối chiếu checklist Chrome Web Store policy (cập nhật 3/2025, thực thi 10/06/2025)"
  - "DEC-AFFIL-20: không có chế độ 'tự gắn affiliate' ẩn; mọi affiliate đi qua đúng một cửa (TASK-AFFIL-002) - guardrail khẳng định không có cửa thứ hai"

language: "TypeScript 5.x (extension guardrail) + Go 1.22 (affil-svc assertion); test CI gate"
service: shopass/extension/ + shopass/services/affil/
new_files:
  - extension/src/guardrails/no-cookie-stuffing.ts
  - extension/src/guardrails/user-gesture.ts
  - extension/src/guardrails/manifest-audit.ts
  - extension/test/guardrails/no-cookie-stuffing.test.ts
  - extension/test/guardrails/manifest-audit.test.ts
  - extension/test/guardrails/single-affiliate-path.test.ts
  - services/affil/internal/affil/guardrail_assert.go
  - services/affil/internal/affil/guardrail_assert_test.go
  - docs/compliance/CHROME-WEBSTORE-AFFILIATE-CHECKLIST.md
modified_files: []
allowed_tools:
  - file_read: extension/**
  - file_read: services/affil/**
  - file_write: extension/**
  - file_write: services/affil/**
  - bash: cd extension && npm test
  - bash: cd services/affil && go test ./...
disallowed_tools:
  - thêm permission cho phép sửa cookie/redirect domain sàn vào manifest (vi phạm DEC-AFFIL-18)
  - tạo đường gắn affiliate thứ hai bỏ qua TASK-AFFIL-002 (vi phạm DEC-AFFIL-20)
  - điều hướng affiliate không bắt nguồn từ user gesture (vi phạm DEC-AFFIL-17)
  - cho guardrail là cảnh báo mềm thay vì CI gate đỏ build (vi phạm DEC-AFFIL-19)

effort_hours: 5
sub_tasks:
  - "1.0h: no-cookie-stuffing.ts - quét codebase extension tìm dấu hiệu set cookie domain sàn / chrome.cookies.set / document.cookie trên host sàn"
  - "0.5h: user-gesture.ts - assert mọi mở affiliate link gắn với isTrusted user event; không mở từ alarm/timer/nền"
  - "1.0h: manifest-audit.ts - kiểm manifest KHÔNG có cookies permission cho host sàn, KHÔNG webRequestBlocking sửa redirect, host_permissions tối thiểu"
  - "0.5h: guardrail_assert.go - affil-svc khẳng định chỉ một route tạo link (TASK-AFFIL-002) + mọi link kèm disclosure"
  - "1.0h: 3 test extension (no-cookie-stuffing, manifest-audit, single-affiliate-path) - vi phạm làm đỏ build"
  - "0.5h: guardrail_assert_test.go - giả lập route thứ hai/disclosure rỗng -> assert fail"
  - "0.5h: CHROME-WEBSTORE-AFFILIATE-CHECKLIST.md - checklist policy 3/2025 (thực thi 10/06/2025) ánh xạ tới guardrail"

risk_if_skipped: "Vụ PayPal Honey là kịch bản tồn vong: MegaLag phơi bày Honey thay cookie affiliate nền, Honey mất ~3 triệu user trong 2 tuần, Google cập nhật chính sách Chrome Web Store 3/2025 (thực thi 10/06/2025) cấm chèn affiliate khi không có lợi ích trực tiếp + bắt buộc user-action + disclosure, rồi Rakuten/Impact/Awin lần lượt gỡ Honey tháng 01/2026. SănDeal là extension đọc cookie phiên sàn - dễ bị nghi y hệt Honey. Nếu không có guardrail tự động khẳng định KHÔNG cookie-stuffing/dropping/pop-under/auto-redirect/forced-install, một dòng code vô tình (hoặc một PR thiếu review) có thể đưa hành vi bị cấm vào, khiến extension bị Chrome gỡ và network đình chỉ - phá hủy cả kênh phân phối lẫn dòng doanh thu. Guardrail là CI gate biến lời hứa minh bạch (§5.4) thành ràng buộc kỹ thuật không thể vô tình vi phạm. Đây là tấm khiên bảo vệ chính moat niềm tin của sản phẩm."
---

## §1 - Mô tả (BCP-14 normative)

Hệ thống **MUST** có một bộ guardrail tự động (lint + assertion chạy như CI gate) khẳng định extension và affil-svc KHÔNG bao giờ thực hiện các hành vi bị cấm kiểu Honey, và mọi affiliate đi qua đúng một cửa user-initiated có disclosure (TASK-AFFIL-002). Hợp đồng:

1. Guardrail **MUST** khẳng định extension KHÔNG cookie-stuffing và KHÔNG cookie-dropping trên domain sàn: không gọi `chrome.cookies.set`/`document.cookie` để set cookie affiliate trên host sàn (DEC-AFFIL-16). Phát hiện -> test đỏ.
2. Guardrail **MUST** khẳng định KHÔNG pop-under, KHÔNG auto-redirect nền, KHÔNG forced-install: extension không mở tab/cửa sổ affiliate nền, không `chrome.tabs.update`/`window.location` điều hướng affiliate khi không có user gesture.
3. Mọi điều hướng affiliate **MUST** bắt nguồn từ một user gesture thật (sự kiện `isTrusted`, bấm) và đi qua TASK-AFFIL-002 với `user_initiated=true` (DEC-AFFIL-17). Guardrail `user-gesture.ts` khẳng định không có đường mở affiliate link từ `alarm`/`setTimeout`/`setInterval`/khởi động nền.
4. `manifest-audit.ts` **MUST** kiểm manifest extension: KHÔNG khai báo `cookies` permission cho host sàn, KHÔNG `webRequestBlocking` dùng để sửa redirect/chèn header affiliate, `host_permissions` chỉ ở mức tối thiểu cần để đọc DOM (DEC-AFFIL-18). Manifest vi phạm -> test đỏ.
5. Guardrail **MUST** khẳng định chỉ tồn tại đúng MỘT đường tạo affiliate link (TASK-AFFIL-002) (DEC-AFFIL-20): không có hàm/endpoint thứ hai tự ghép affiliate URL hay tự ghi `affiliate_click` ngoài luồng đó. `guardrail_assert.go` kiểm affil-svc.
6. Guardrail **MUST** khẳng định mọi affiliate link trả ra đều kèm disclosure không rỗng (đồng nhất TASK-AFFIL-002 #10): assertion fail nếu một đường nào trả `deep_link` mà không có `disclosure`.
7. Guardrail **MUST** chạy như CI gate bắt buộc (DEC-AFFIL-19): vi phạm bất kỳ làm đỏ build, KHÔNG phải cảnh báo mềm bỏ qua được. Đây là điều kiện merge.
8. **MUST** duy trì `CHROME-WEBSTORE-AFFILIATE-CHECKLIST.md` ánh xạ từng yêu cầu của chính sách Chrome Web Store (cập nhật 3/2025, thực thi 10/06/2025) tới guardrail/test tương ứng - mỗi mục policy có một kiểm tra chứng minh tuân thủ. Guardrail cũng **MUST** bao phủ auto-test mã giảm (TASK-CART-005, nếu có): user-initiated + nhịp người, không tự chốt đơn, không tự áp affiliate khi thử mã (liên kết §3.5 thuật toán 4).
9. Guardrail extension **MUST** quét cả mã nguồn (static) lẫn khẳng định runtime: static bắt pattern bị cấm trong code; runtime assert chặn nếu một mở-link xảy ra mà không gắn user gesture (defense in depth với TASK-EXT-003).
10. Guardrail **MUST** coi danh sách hành vi cấm là allowlist-of-paths cho affiliate: chỉ đường đã được phê duyệt (TASK-AFFIL-002) hợp lệ; mọi đường mới chạm affiliate phải thêm vào kiểm tra và qua review (tinh thần allowlist của TASK-EXT-003).
11. Khi guardrail phát hiện vi phạm, thông báo lỗi **MUST** chỉ rõ hành vi cấm nào, ở file/đường nào, và trích mục policy bị vi phạm - để sửa nhanh, không chỉ "fail".
12. **SHOULD** phát metric/đếm guardrail trong CI: số kiểm tra chạy, số vi phạm chặn - để có bằng chứng định lượng guardrail đang hoạt động qua thời gian.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao guardrail là CI gate, không phải tài liệu (DEC-AFFIL-19)?** Lời hứa "chúng tôi không làm như Honey" chỉ đáng tin nếu không thể vô tình vi phạm. Một tài liệu policy không chặn được một PR thêm `chrome.cookies.set` lúc 2 giờ sáng. Một test đỏ thì có. Biến mọi hành vi bị cấm thành kiểm tra tự động làm đỏ build đặt sự tuân thủ vào chính cơ chế merge, nơi nó không thể bị quên.

**Vì sao kiểm cả static lẫn runtime (§1 #9)?** Static lint bắt pattern bị cấm trong mã (ai đó gõ `document.cookie =` trên host sàn). Nhưng mã có thể né static qua gián tiếp. Runtime assert là lưới thứ hai: nếu một mở-link thực sự xảy ra mà không gắn user gesture, nó bị chặn tại chỗ. Hai lớp cùng nhau khó lọt hơn một.

**Vì sao manifest audit (DEC-AFFIL-18)?** Cookie-stuffing cần quyền chạm cookie domain sàn; auto-redirect nền cần quyền sửa request. Cách mạnh nhất để chứng minh extension không làm những việc đó là chứng minh nó không có quyền làm. Kiểm manifest không khai `cookies` permission cho host sàn và không `webRequestBlocking` sửa redirect đóng khả năng vi phạm ngay từ quyền hạn - principle of least privilege.

**Vì sao khẳng định chỉ một cửa affiliate (DEC-AFFIL-20)?** TASK-AFFIL-002 đã làm đúng (user-initiated + disclosure). Nhưng nếu ai đó thêm một hàm thứ hai tự ghép affiliate URL "cho tiện", cánh cửa Honey mở lại qua lối sau. Guardrail khẳng định đúng một đường tồn tại buộc mọi affiliate đi qua cửa đã kiểm soát - không có lối tắt nào.

**Vì sao ánh xạ checklist policy (§1 #8)?** Chính sách Chrome Web Store (3/2025, thực thi 10/06/2025) có nhiều mục cụ thể. Ánh xạ mỗi mục tới một guardrail/test biến "chúng tôi tuân policy" từ một khẳng định mơ hồ thành một bảng truy vết: mục này của policy được kiểm bởi test kia. Khi policy cập nhật, ta biết phải thêm/sửa kiểm tra nào.

**Vì sao thông báo lỗi chỉ rõ mục policy (§1 #11)?** Một dev gặp guardrail đỏ cần biết không chỉ "sai" mà "sai điều gì của policy nào và sửa ở đâu". Thông báo dẫn chiếu policy biến guardrail thành công cụ dạy về compliance, không chỉ rào chặn - giảm khả năng lặp lại vi phạm.

---

## §3 - Hợp đồng API / DDL

### Static guard - không cookie-stuffing (TypeScript)

```ts
// extension/src/guardrails/no-cookie-stuffing.ts
const SHOP_HOSTS = ["shopee.vn", "tiktok.com", "lazada.vn"];

// Pattern bị cấm: set cookie / chrome.cookies.set / điều hướng affiliate trên host sàn.
const BANNED = [
  /chrome\.cookies\.set/,
  /document\.cookie\s*=/,
  /chrome\.tabs\.update\([^)]*affiliate/i,
  /window\.open\([^)]*aff/i,        // pop-under affiliate
];

// scanSource trả danh sách vi phạm {file, line, rule, policyRef}. Rỗng = pass.
export function scanSource(files: SourceFile[]): Violation[] {
  const out: Violation[] = [];
  for (const f of files) {
    f.lines.forEach((ln, i) => {
      for (const re of BANNED) {
        if (re.test(ln) && touchesShopHost(f, ln, SHOP_HOSTS)) {
          out.push({ file: f.path, line: i + 1, rule: re.source,
            policyRef: "Chrome Web Store affiliate policy 2025-03 (enforced 2025-06-10): no cookie/redirect injection without user benefit + action" });
        }
      }
    });
  }
  return out;
}
```

### Runtime guard - user gesture (TypeScript)

```ts
// extension/src/guardrails/user-gesture.ts
// openAffiliate CHỈ chạy khi có user gesture thật (ev.isTrusted) và qua TASK-AFFIL-002.
export function openAffiliate(ev: Event | undefined, deepLink: string): void {
  if (!ev || !ev.isTrusted) {
    throw new Error("affiliate navigation must originate from a trusted user gesture"); // §1 #3
  }
  // mở trong tab user bấm; KHÔNG set cookie, KHÔNG nền (cookie do trang sàn set)
  window.open(deepLink, "_blank", "noopener");
}
```

### Manifest audit (TypeScript)

```ts
// extension/src/guardrails/manifest-audit.ts
export function auditManifest(m: Manifest): Violation[] {
  const out: Violation[] = [];
  if ((m.permissions ?? []).includes("cookies")) {
    out.push(viol("manifest", "cookies permission present", "least-privilege: no cookie access to shop hosts (§1 #4)"));
  }
  if ((m.permissions ?? []).includes("webRequestBlocking")) {
    out.push(viol("manifest", "webRequestBlocking present", "no blocking webRequest to rewrite redirects (§1 #4)"));
  }
  return out;
}
```

### Single-path assertion (Go)

```go
// services/affil/internal/affil/guardrail_assert.go

// AssertSingleAffiliatePath khẳng định đúng MỘT route tạo affiliate link (TASK-AFFIL-002)
// và mọi response link đều kèm disclosure không rỗng (§1 #5,#6).
func AssertSingleAffiliatePath(routes []Route) error {
    linkRoutes := 0
    for _, r := range routes {
        if r.CreatesAffiliateLink {
            linkRoutes++
            if !r.IncludesDisclosure {
                return fmt.Errorf("affiliate link route %s missing disclosure", r.Path)
            }
        }
    }
    if linkRoutes != 1 {
        return fmt.Errorf("expected exactly 1 affiliate-link path, found %d (no back-door)", linkRoutes)
    }
    return nil
}
```

---

## §4 - Acceptance criteria

1. `scanSource` trên codebase sạch -> 0 vi phạm (pass).
2. Bơm một file có `document.cookie = "aff=..."` gắn host `shopee.vn` -> `scanSource` trả 1 vi phạm với `policyRef` Chrome 2025; test đỏ.
3. Bơm `chrome.cookies.set` trên host sàn -> bị bắt.
4. Bơm `window.open(...aff...)` (pop-under affiliate) -> bị bắt.
5. `openAffiliate(undefined, link)` (không user gesture) -> ném lỗi; `openAffiliate(untrustedEvent, link)` -> ném lỗi.
6. `openAffiliate(trustedClickEvent, link)` -> mở link, không lỗi.
7. `auditManifest` với manifest có `cookies` permission -> 1 vi phạm; với `webRequestBlocking` -> 1 vi phạm.
8. `auditManifest` trên manifest tối thiểu hợp lệ (chỉ host_permissions đọc DOM) -> 0 vi phạm.
9. `AssertSingleAffiliatePath` với đúng một route link + disclosure -> `nil`.
10. `AssertSingleAffiliatePath` với hai route tạo link (back-door) -> lỗi; với route link thiếu disclosure -> lỗi.
11. Guardrail chạy trong CI là gate bắt buộc: mô phỏng vi phạm -> exit code khác 0 (đỏ build), không phải warning.
12. `CHROME-WEBSTORE-AFFILIATE-CHECKLIST.md` tồn tại và mỗi mục policy ánh xạ tới một test/guard cụ thể.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/guardrails/no-cookie-stuffing.test.ts
import { scanSource } from "../../src/guardrails/no-cookie-stuffing";

test("codebase sạch -> 0 vi phạm", () => {
  expect(scanSource(loadExtensionSource())).toHaveLength(0);
});

test("cookie-stuffing trên host sàn bị bắt + dẫn chiếu policy", () => {
  const f = mkFile("src/bad.ts", [`if (host==="shopee.vn") document.cookie = "aff=123";`]);
  const v = scanSource([f]);
  expect(v).toHaveLength(1);
  expect(v[0].policyRef).toMatch(/2025-06-10/);
});

test("pop-under affiliate bị bắt", () => {
  const f = mkFile("src/pop.ts", [`window.open("https://go.aff.x?sub=1","_blank")`]);
  expect(scanSource([f]).length).toBeGreaterThan(0);
});
```

```ts
// extension/test/guardrails/manifest-audit.test.ts
import { auditManifest } from "../../src/guardrails/manifest-audit";

test("cookies permission bị bắt", () => {
  expect(auditManifest({ permissions: ["cookies", "storage"] }).length).toBe(1);
});

test("manifest tối thiểu hợp lệ -> sạch", () => {
  expect(auditManifest({ permissions: ["storage", "alarms"],
    host_permissions: ["https://shopee.vn/*"] })).toHaveLength(0);
});
```

```ts
// extension/test/guardrails/single-affiliate-path.test.ts
import { openAffiliate } from "../../src/guardrails/user-gesture";

test("mở affiliate không user gesture -> ném lỗi", () => {
  expect(() => openAffiliate(undefined, "https://go.aff.x")).toThrow(/trusted user gesture/);
});

test("untrusted event -> ném lỗi", () => {
  expect(() => openAffiliate({ isTrusted: false } as any, "https://go.aff.x")).toThrow();
});
```

```go
// services/affil/internal/affil/guardrail_assert_test.go
func TestAssert_SinglePath_OK(t *testing.T) {
    routes := []Route{{Path: "/v1/affiliate/link", CreatesAffiliateLink: true, IncludesDisclosure: true}}
    require.NoError(t, AssertSingleAffiliatePath(routes))
}

func TestAssert_BackDoor_Fails(t *testing.T) {
    routes := []Route{
        {Path: "/v1/affiliate/link", CreatesAffiliateLink: true, IncludesDisclosure: true},
        {Path: "/v1/affiliate/auto", CreatesAffiliateLink: true, IncludesDisclosure: true}, // cửa thứ hai
    }
    require.Error(t, AssertSingleAffiliatePath(routes))
}

func TestAssert_NoDisclosure_Fails(t *testing.T) {
    routes := []Route{{Path: "/v1/affiliate/link", CreatesAffiliateLink: true, IncludesDisclosure: false}}
    require.Error(t, AssertSingleAffiliatePath(routes))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `no-cookie-stuffing.ts` (static scan pattern bị cấm trên host sàn) -> `user-gesture.ts` (runtime assert `isTrusted`) -> `manifest-audit.ts` (least-privilege manifest) -> `guardrail_assert.go` (affil-svc: đúng một route link + disclosure) -> 3 test extension + 1 test Go -> `CHROME-WEBSTORE-AFFILIATE-CHECKLIST.md` ánh xạ policy. Mọi guardrail nối vào CI là gate bắt buộc (`npm test` + `go test` trong pipeline merge); vi phạm trả exit khác 0. Static scan đọc nguồn extension; runtime guard là hàm bắt buộc đi qua khi mở affiliate (không có `window.open` affiliate nào khác ngoài `openAffiliate`).

---

## §7 - Phụ thuộc

- **TASK-AFFIL-002** - cửa affiliate hợp lệ duy nhất (user-initiated + disclosure); guardrail khẳng định không có cửa thứ hai và mọi link kèm disclosure.
- **TASK-EXT-003** - pipeline tối thiểu hóa (allowlist, local-first); guardrail dùng cùng tinh thần allowlist cho đường affiliate và bổ trợ defense-in-depth.
- **TASK-TRUST-001** - extension open-source + reproducible build: guardrail là bằng chứng kiểm chứng được rằng mã công khai không cookie-stuffing.
- **TASK-TRUST-002** - chính sách minh bạch/local-first; guardrail thực thi nó bằng CI gate.
- **TASK-AFFIL-005 (downstream)** - cashback chỉ an toàn khi guardrail đảm bảo affiliate sạch (không gian lận attribution từ phía ta).
- **NFR-AFFIL-001** - guardrail là cơ chế thực thi của ràng buộc compliance affiliate.
- Lib: test runner (`vitest`/`jest`) cho extension; `testing` cho Go.

---

## §8 - Payload ví dụ

### Vi phạm bị guardrail bắt (CI đỏ)

```text
$ npm test -- guardrails
FAIL extension/test/guardrails/no-cookie-stuffing.test.ts
  cookie-stuffing trên host sàn bị bắt
  Violation: src/checkout.ts:42  rule=/document\.cookie\s*=/
    policyRef: Chrome Web Store affiliate policy 2025-03 (enforced 2025-06-10):
               no cookie/redirect injection without user benefit + action
exit code: 1   # đỏ build, chặn merge
```

### Checklist ánh xạ (trích)

```text
docs/compliance/CHROME-WEBSTORE-AFFILIATE-CHECKLIST.md
| Policy (2025-03, enforce 2025-06-10) | Guardrail/test |
| No affiliate cookie without user benefit | no-cookie-stuffing.test.ts |
| User action required for affiliate     | user-gesture.test.ts (isTrusted) |
| Disclosure required                     | guardrail_assert_test.go (IncludesDisclosure) |
| No background redirect / pop-under      | scanSource window.open/tabs.update |
| Least-privilege permissions             | manifest-audit.test.ts |
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Tự động quét bản build đã đóng gói (zip) trước khi upload Chrome Web Store, ngoài quét nguồn - thêm bước CI khi có pipeline release của TASK-TRUST-001.
- Kiểm runtime trong môi trường thật (extension đang chạy) phát hiện điều hướng nền bất thường - thêm telemetry guard sau; hiện static + runtime-assert ở mức hàm.
- Mở rộng pattern cấm khi policy Chrome cập nhật tiếp - checklist là nơi theo dõi; thêm pattern khi policy đổi.
- Áp guardrail tương tự cho mobile app (TASK-MOBILE) khi có universal checkout assistant - nhân rộng sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Cookie-stuffing lọt vào code | scanSource test (§4 #2) | bị Chrome gỡ kiểu Honey | Static scan host sàn + CI đỏ (§1 #1) |
| Auto-redirect / pop-under nền | scanSource + user-gesture | vi phạm policy, mất niềm tin | Bắt pattern + runtime isTrusted (§1 #2,#3) |
| Manifest có quyền sửa cookie/redirect | manifest-audit test | mở khả năng vi phạm | Least-privilege audit (§1 #4) |
| Cửa affiliate thứ hai (back-door) | AssertSinglePath test | lối tắt né disclosure | Khẳng định đúng một route (§1 #5) |
| Link thiếu disclosure | AssertSinglePath test | hưởng lợi thiếu minh bạch | Bắt buộc disclosure (§1 #6) |
| Guardrail là warning mềm | CI exit-code test | vi phạm lọt qua merge | CI gate bắt buộc, đỏ build (§1 #7) |
| Điều hướng affiliate từ timer/nền | user-gesture assert | tự động hóa kiểu Honey | Chỉ từ isTrusted gesture (§1 #3) |
| Policy cập nhật, guardrail lỗi thời | checklist review | tuân nhầm policy cũ | CHROME-WEBSTORE-AFFILIATE-CHECKLIST cập nhật (§1 #8) |

---

## §11 - Ghi chú

- Guardrail biến lời hứa "không làm như Honey" thành CI gate không thể vô tình vi phạm - đỏ build chặn merge, không phải tài liệu khuyến nghị.
- Kiểm static (pattern trong mã) + runtime (isTrusted gesture) + manifest (least-privilege) là ba lớp khó lọt hơn một.
- Cách mạnh nhất để chứng minh không cookie-stuffing là không có quyền cookie domain sàn: manifest audit đóng khả năng từ quyền hạn.
- Khẳng định đúng một cửa affiliate (TASK-AFFIL-002) đóng lối sau cho hành vi Honey len vào qua một hàm "cho tiện".
- Checklist ánh xạ từng mục policy Chrome (3/2025, thực thi 10/06/2025) tới một test là bảng truy vết compliance, cập nhật khi policy đổi.
- Thông báo lỗi dẫn chiếu mục policy biến guardrail thành công cụ dạy compliance, giảm lặp lại vi phạm.
- Đây là tấm khiên bảo vệ moat niềm tin (§5.4) và kênh phân phối (Chrome Web Store) - mất chúng là mất sản phẩm.

---

*Hết TASK-AFFIL-004. Status: ready_to_implement (mục tiêu audit 10/10).*
