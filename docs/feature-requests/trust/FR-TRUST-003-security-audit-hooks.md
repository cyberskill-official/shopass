---
id: FR-TRUST-003
title: "Hook security audit độc lập - chứng minh KHÔNG gửi cookie/mật khẩu (egress test, SBOM, verify reproducible build); báo cáo audit công khai có thể tái chạy bởi bên thứ ba"
module: TRUST
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-EXT-003, FR-COMPLY-005, FR-TRUST-001, FR-TRUST-002, NFR-TRUST-001]
depends_on: [FR-EXT-003, FR-COMPLY-005]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.4 (trust & security: (2) security audit độc lập; (3) chứng minh KHÔNG gửi cookie/mật khẩu)"
  - "docs/... §3.8 (no-cleartext, token không rời client), §5.2 (Chrome gỡ extension - cần bằng chứng kiểm chứng được)"
source_decisions:
  - "DEC-TRUST-11: audit là tập hook TÁI CHẠY ĐƯỢC (egress test + SBOM + verify reproducible build), KHÔNG phải một báo cáo PDF một lần; bên thứ ba chạy lại ra cùng kết luận"
  - "DEC-TRUST-12: egress test chặn-tất-cả mạng rồi cho extension chạy luồng đọc giỏ; assert KHÔNG có request nào mang cookie/token/PII rời máy (kiểm tại biên mạng, không tin self-report)"
  - "DEC-TRUST-13: SBOM (CycloneDX) liệt kê mọi dependency + hash; quét lỗ hổng đã biết; pin dependency để SBOM ổn định và tái lập"
  - "DEC-TRUST-14: audit hook gọi verify-reproducible (FR-TRUST-001) + payload_guard (FR-COMPLY-005) như thành phần con; FR-TRUST-003 là lớp tổng hợp 'bằng chứng đầu cuối'"
  - "DEC-TRUST-15: kết quả audit xuất thành báo cáo có cấu trúc (JSON + markdown) + hướng dẫn để người ngoài tự chạy; thất bại bất kỳ hook -> audit FAIL"

language: "TypeScript/Node (egress test rig, Playwright network intercept) + shell (orchestrate) + CycloneDX SBOM"
service: shopass/extension/ + shopass/audit/
new_files:
  - audit/run-security-audit.sh
  - audit/egress/egress-guard.test.ts
  - audit/egress/network-trap.ts
  - audit/sbom/generate-sbom.sh
  - audit/sbom/scan-vulnerabilities.sh
  - audit/report/audit-report.template.md
  - audit/report/build-report.ts
  - audit/THIRD-PARTY-AUDIT-GUIDE.md
modified_files:
  - extension/.github/workflows/reproducible-publish-gate.yml   # gọi run-security-audit.sh trong gate
allowed_tools:
  - file_read: audit/**, extension/**, services/comply/**
  - file_write: audit/**
  - bash: bash audit/run-security-audit.sh
disallowed_tools:
  - viết báo cáo audit thủ công không tái chạy được (vi phạm DEC-TRUST-11)
  - tin self-report của extension thay vì kiểm tại biên mạng (vi phạm DEC-TRUST-12 - cookie có thể lọt mà code 'nói' là không)
  - bỏ qua SBOM/vuln scan rồi tuyên bố 'đã audit' (vi phạm DEC-TRUST-13)
  - cho audit PASS khi một hook con (egress/sbom/reproducible/payload_guard) FAIL (vi phạm DEC-TRUST-15)

effort_hours: 6
sub_tasks:
  - "1.5h: network-trap.ts + egress-guard.test.ts - chặn mạng, chạy luồng đọc giỏ qua Playwright, bắt mọi outbound, assert không cookie/token/PII"
  - "1.0h: generate-sbom.sh (CycloneDX) + scan-vulnerabilities.sh - liệt kê dependency + hash + quét CVE đã biết"
  - "1.0h: run-security-audit.sh - orchestrate egress + sbom + verify-reproducible (FR-TRUST-001) + payload_guard (FR-COMPLY-005); FAIL nếu hook nào đỏ"
  - "1.0h: build-report.ts + audit-report.template.md - tổng hợp kết quả thành JSON + markdown có cấu trúc"
  - "1.0h: THIRD-PARTY-AUDIT-GUIDE.md - hướng dẫn người ngoài tự clone + chạy lại toàn bộ audit"
  - "0.5h: nối run-security-audit.sh vào CI publish gate (chặn ship nếu audit FAIL)"

risk_if_skipped: "Tài liệu nguồn (§5.4) liệt kê 'security audit độc lập' và 'chứng minh KHÔNG gửi cookie/mật khẩu' là trụ niềm tin. FR-TRUST-001 mở source + tái lập (ai cũng đọc/build được), FR-TRUST-002 công bố chính sách, FR-COMPLY-005 cưỡng chế bằng kiểm tĩnh - nhưng tĩnh chưa đủ: một request runtime vẫn có thể mang cookie ra ngoài qua đường mà grep không thấy. Egress test ở BIÊN MẠNG là bằng chứng động: chặn tất cả, cho extension chạy thật, và quan sát đúng những gì rời máy. Nếu cookie lọt, test bắt được dù code 'nói' là không. SBOM + vuln scan đóng rủi ro dependency độc hại (kênh tấn công supply-chain). Quan trọng nhất: audit phải TÁI CHẠY được bởi bên thứ ba - một báo cáo PDF một lần thì không ai kiểm lại được, đúng kiểu 'tin tôi đi' mà SănDeal muốn tránh hậu-Honey. Bỏ FR này, các trụ kia thiếu mảnh bằng chứng đầu cuối quan trọng nhất: quan sát thực tế dữ liệu rời máy."
---

## §1 - Mô tả (BCP-14 normative)

SănDeal **MUST** cung cấp một bộ hook security audit tái chạy được, chứng minh bằng quan sát thực tế (không tin self-report) rằng extension KHÔNG gửi cookie/mật khẩu/token ra ngoài, kèm SBOM + verify reproducible build, và xuất báo cáo bên thứ ba tự chạy lại được. Hợp đồng:

1. Audit **MUST** là tập hook TÁI CHẠY ĐƯỢC (egress test + SBOM/vuln scan + verify reproducible build + payload guard), KHÔNG phải báo cáo một lần (DEC-TRUST-11). Bên thứ ba chạy lại phải ra cùng kết luận.
2. **MUST** có egress test (`egress-guard.test.ts` + `network-trap.ts`): chặn toàn bộ mạng, cho extension chạy luồng đọc giỏ thật (qua Playwright), bắt MỌI request outbound, và assert KHÔNG request nào chứa cookie phiên sàn, token, mật khẩu, hay PII (DEC-TRUST-12). Kiểm tại BIÊN MẠNG, không dựa lời khai của code.
3. Egress test **MUST** kiểm cả URL, header, và body của mỗi outbound; bất kỳ chuỗi giống cookie/token/PII trong bất kỳ phần nào -> test FAIL.
4. Egress test **MUST** chỉ cho phép outbound tới endpoint backend SănDeal đã biết với payload đúng allowlist (`OutboundPayload` FR-EXT-003); request tới host lạ -> FAIL (chống rò qua kênh bên).
5. **MUST** sinh SBOM định dạng CycloneDX (`generate-sbom.sh`) liệt kê mọi dependency + phiên bản + hash; và quét lỗ hổng đã biết (`scan-vulnerabilities.sh`) (DEC-TRUST-13). Dependency phải pin để SBOM ổn định.
6. SBOM scan **MUST** FAIL audit nếu phát hiện CVE mức cao/nghiêm trọng chưa có miễn trừ ghi rõ lý do.
7. `run-security-audit.sh` **MUST** orchestrate và gọi như thành phần con (DEC-TRUST-14): (a) egress test, (b) SBOM + vuln scan, (c) `verify-reproducible.sh` (FR-TRUST-001), (d) `payload_guard` (FR-COMPLY-005). Audit PASS chỉ khi MỌI hook con PASS (DEC-TRUST-15).
8. Audit **MUST** xuất báo cáo có cấu trúc: JSON (máy đọc) + markdown (người đọc) liệt kê từng hook, kết quả, và bằng chứng (hash, danh sách outbound đã quan sát, số CVE).
9. **MUST** có `THIRD-PARTY-AUDIT-GUIDE.md`: hướng dẫn người ngoài clone repo + chạy lại toàn bộ audit trên máy họ và đối chiếu kết quả - điều kiện để "audit độc lập" có nghĩa.
10. CI publish gate (FR-TRUST-001) **MUST** gọi `run-security-audit.sh`; audit FAIL **MUST** chặn ship lên store.
11. Egress test **MUST** bao phủ luồng nhạy cảm nhất: đọc giỏ hàng có phiên đăng nhập sàn (nơi cookie hiện diện) - đảm bảo ngay cả khi cookie có sẵn trong context, nó KHÔNG rời máy.
12. **MUST NOT** có nhánh nào trong audit bỏ qua hook con khi "chạy nhanh"/"môi trường CI khác"; mọi lần chạy phải đủ bốn hook hoặc audit báo INCOMPLETE (không được ngầm PASS).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao kiểm tại biên mạng, không tin self-report (DEC-TRUST-12)?** Code có thể "nói" nó không gửi cookie, nhưng một bug, một thư viện bên thứ ba, hay một đường fetch quên lọc vẫn có thể mang cookie ra ngoài. Kiểm tĩnh (grep, FR-COMPLY-005) bắt được mẫu đã biết, nhưng không thấy hành vi runtime. Egress test chặn tất cả rồi quan sát đúng những gì thật sự rời máy - đây là bằng chứng động, mạnh hơn mọi lời khai trong mã.

**Vì sao audit phải tái chạy được (DEC-TRUST-11)?** Một báo cáo PDF "đã được hãng X audit" là "tin tôi đi" - người dùng không kiểm lại được, và nó cũ ngay khi mã đổi. Audit của SănDeal là tập hook ai cũng chạy lại được: clone repo, chạy script, xem cùng kết quả. Đây là khác biệt cốt lõi hậu-Honey: bằng chứng kiểm chứng được thay cho uy tín mượn.

**Vì sao SBOM + vuln scan (DEC-TRUST-13)?** "Không gửi cookie" không đủ nếu một dependency độc hại làm việc đó sau lưng. Supply-chain là kênh tấn công thật. SBOM liệt kê chính xác mọi thứ extension phụ thuộc (+ hash), vuln scan bắt lỗ hổng đã biết. Pin dependency làm SBOM ổn định để tái lập và để biết chính xác cái gì chạy.

**Vì sao chỉ cho outbound tới endpoint đã biết với payload đúng allowlist (§1 #4)?** Rò dữ liệu không chỉ qua "gửi cookie tới backend" - còn có thể qua gửi dữ liệu tới host lạ (kênh bên). Whitelist endpoint + assert payload khớp allowlist FR-EXT-003 đóng cả hai: đúng đích VÀ đúng nội dung. Bất kỳ outbound ngoài luật -> đỏ.

**Vì sao audit là lớp tổng hợp gọi các hook con (DEC-TRUST-14)?** FR-TRUST-001 lo reproducible, FR-COMPLY-005 lo kiểm tĩnh + payload guard. FR-TRUST-003 không lặp lại chúng mà gọi chúng như mảnh ghép và thêm mảnh còn thiếu (egress động + SBOM), rồi tổng hợp thành "bằng chứng đầu cuối". Audit PASS chỉ khi cả chuỗi PASS - một mắt xích đỏ là cả audit đỏ.

**Vì sao bao phủ luồng có phiên đăng nhập (§1 #11)?** Điểm nhạy cảm nhất là khi cookie phiên sàn ĐANG hiện diện trong context (user đã đăng nhập, extension đọc giỏ). Nếu test chạy không có phiên, nó không chứng minh được gì về tình huống thật. Phải đặt cookie vào context rồi chứng minh nó vẫn không rời máy - đó mới là bằng chứng đúng kịch bản lo ngại.

---

## §3 - Hợp đồng API / DDL

### network-trap.ts (bắt mọi outbound)

```ts
// audit/egress/network-trap.ts
import type { Page, Request } from "playwright";

export interface Outbound { url: string; headers: Record<string,string>; body: string; }

const BACKEND_HOST = "api.sandeal.vn";              // endpoint hợp lệ DUY NHẤT
const CRED_RE = /(SPC_|session|sessionid|token|bearer|password|mật khẩu|@[\w.]+\.\w+)/i;

export function installTrap(page: Page, captured: Outbound[]) {
  page.on("request", (req: Request) => {
    captured.push({ url: req.url(), headers: req.headers(), body: req.postData() ?? "" });
  });
}

export function assertNoCredentialEgress(captured: Outbound[]) {
  for (const o of captured) {
    const host = new URL(o.url).host;
    if (host !== BACKEND_HOST)
      throw new Error(`outbound tới host lạ: ${host} (chỉ cho ${BACKEND_HOST})`);
    const blob = `${o.url}\n${JSON.stringify(o.headers)}\n${o.body}`;
    if (CRED_RE.test(blob))
      throw new Error(`PHÁT HIỆN credential/PII rời máy trong request tới ${host}`);
  }
}
```

### egress-guard.test.ts (chạy luồng thật, có phiên đăng nhập)

```ts
// audit/egress/egress-guard.test.ts
import { chromium } from "playwright";
import { installTrap, assertNoCredentialEgress, Outbound } from "./network-trap";

test("đọc giỏ có cookie phiên sàn -> KHÔNG cookie/token rời máy", async () => {
  const browser = await chromium.launch();
  const ctx = await browser.newContext();
  // đặt cookie phiên sàn vào context (mô phỏng user đã đăng nhập - kịch bản nhạy cảm nhất)
  await ctx.addCookies([{ name: "SPC_SESSION", value: "eyJ...secret", domain: ".shopee.vn", path: "/" }]);
  const page = await ctx.newPage();
  const captured: Outbound[] = [];
  installTrap(page, captured);

  await loadExtensionAndReadCart(page);     // chạy luồng đọc giỏ thật (FR-EXT-002/003)

  assertNoCredentialEgress(captured);       // assert tại biên mạng
  await browser.close();
});
```

### run-security-audit.sh (orchestrate, FAIL nếu hook con đỏ)

```bash
#!/usr/bin/env bash
# audit/run-security-audit.sh
set -euo pipefail
FAIL=0

echo "== [1/4] Egress test (động) =="
( cd extension && npx playwright test ../audit/egress ) || FAIL=1

echo "== [2/4] SBOM + vuln scan =="
bash audit/sbom/generate-sbom.sh && bash audit/sbom/scan-vulnerabilities.sh || FAIL=1

echo "== [3/4] Verify reproducible build (FR-TRUST-001) =="
bash extension/scripts/verify-reproducible.sh "$(git rev-parse HEAD)" "${SHIPPED:-extension/dist.zip}" || FAIL=1

echo "== [4/4] Payload guard tĩnh (FR-COMPLY-005) =="
( cd services/comply && go test ./internal/audit/... ) || FAIL=1

node audit/report/build-report.ts          # tổng hợp JSON + markdown
if [ "$FAIL" -ne 0 ]; then echo "AUDIT FAIL"; exit 1; fi
echo "AUDIT PASS - bằng chứng đầu cuối: KHÔNG cookie/mật khẩu rời máy"
```

---

## §4 - Acceptance criteria

1. `run-security-audit.sh` chạy đủ bốn hook con; thiếu hook nào -> báo INCOMPLETE, không ngầm PASS.
2. Egress test đặt cookie phiên vào context, chạy luồng đọc giỏ, bắt mọi outbound; luồng sạch -> PASS.
3. Cố tình chèn một fetch lén gửi cookie -> egress test FAIL (bắt được tại biên mạng).
4. Outbound tới host khác `api.sandeal.vn` -> egress test FAIL (chống kênh bên).
5. Outbound mang payload ngoài allowlist (FR-EXT-003) -> FAIL.
6. `generate-sbom.sh` xuất SBOM CycloneDX liệt kê mọi dependency + hash.
7. `scan-vulnerabilities.sh` chèn một dependency có CVE cao đã biết -> audit FAIL (trừ khi có miễn trừ ghi lý do).
8. `run-security-audit.sh` gọi `verify-reproducible.sh` (FR-TRUST-001) + `payload_guard` (FR-COMPLY-005); một trong hai đỏ -> audit FAIL.
9. Audit xuất báo cáo JSON + markdown liệt kê từng hook + kết quả + bằng chứng (hash, danh sách outbound, số CVE).
10. `THIRD-PARTY-AUDIT-GUIDE.md` đủ để người ngoài clone + chạy lại + đối chiếu; chạy thử theo guide ra cùng kết luận.
11. CI publish gate gọi audit; audit FAIL chặn ship lên store.
12. Audit PASS chỉ khi cả bốn hook PASS (không hook nào bị bỏ qua trong "chế độ nhanh").

---

## §5 - Kiểm thử (verification)

```ts
// audit/egress/egress-guard.test.ts (xem §3) - bằng chứng động chính.

// Test "negative control": chèn rò có chủ đích để chứng minh test BẮT được
test("negative control: fetch lén gửi cookie -> egress test phải FAIL", async () => {
  const captured = [{ url: "https://api.sandeal.vn/x", headers: { cookie: "SPC_SESSION=eyJ..." }, body: "" }];
  expect(() => assertNoCredentialEgress(captured)).toThrow(/credential\/PII/);
});

test("negative control: outbound host lạ -> FAIL", () => {
  const captured = [{ url: "https://evil.example/x", headers: {}, body: "{}" }];
  expect(() => assertNoCredentialEgress(captured)).toThrow(/host lạ/);
});
```

```bash
# Chạy toàn bộ audit (giống bên thứ ba sẽ chạy)
bash audit/run-security-audit.sh

# Kiểm báo cáo có cấu trúc tồn tại + liệt kê đủ 4 hook
jq '.hooks | length == 4' audit/report/audit-report.json

# Theo THIRD-PARTY-AUDIT-GUIDE.md, một người ngoài chạy:
git clone <repo> && cd <repo> && bash audit/run-security-audit.sh   # PHẢI ra cùng PASS
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `network-trap.ts` (bắt + assert outbound) -> `egress-guard.test.ts` (chạy luồng đọc giỏ có phiên qua Playwright) -> `generate-sbom.sh` + `scan-vulnerabilities.sh` (CycloneDX + CVE) -> `run-security-audit.sh` (orchestrate bốn hook, gọi `verify-reproducible.sh` của FR-TRUST-001 + `payload_guard` của FR-COMPLY-005) -> `build-report.ts` + template (JSON + markdown) -> `THIRD-PARTY-AUDIT-GUIDE.md` (hướng dẫn tái chạy) -> nối vào CI publish gate. Audit là lớp tổng hợp: nó không lặp lại reproducible/payload_guard mà gọi chúng và thêm egress động + SBOM. Mọi hook đỏ là audit đỏ; mọi lần chạy đủ bốn hook hoặc báo INCOMPLETE.

---

## §7 - Phụ thuộc

- **FR-EXT-003** - allowlist `OutboundPayload` là chuẩn để egress test assert payload đúng (đúng nội dung rời máy).
- **FR-COMPLY-005** - `payload_guard` (kiểm tĩnh + payload nhận từ extension không chứa cookie/token) là một hook con của audit; FR-TRUST-003 thêm tầng động.
- **FR-TRUST-001** - `verify-reproducible.sh` là một hook con; reproducible build + source công khai làm bằng chứng "mã chạy = mã công khai".
- **FR-TRUST-002** - chính sách + data-flow là tài liệu đối chiếu cho audit (audit chứng minh hành vi khớp chính sách công bố).
- Hạ tầng: Playwright (network intercept), CycloneDX (SBOM), trình quét CVE; CI (GitHub Actions) cho publish gate.
- NFR-TRUST-001 - audit này là phép verification chính cho ngưỡng "token không rời client".

---

## §8 - Payload ví dụ

### Báo cáo audit (markdown, trích)

```markdown
# SănDeal Security Audit - v1.4.0 (commit a1b2c3d)
| Hook | Kết quả | Bằng chứng |
|---|---|---|
| Egress (động) | PASS | 3 outbound, tất cả tới api.sandeal.vn, 0 cookie/token/PII |
| SBOM + vuln | PASS | 41 dependency (hash kèm), 0 CVE cao |
| Reproducible build | PASS | rebuilt SHA-256 == shipped 9f2c...e71a |
| Payload guard (tĩnh) | PASS | 0 finding cleartext/token |
=> AUDIT PASS. Tự kiểm: `bash audit/run-security-audit.sh` (xem THIRD-PARTY-AUDIT-GUIDE.md)
```

### Báo cáo audit (JSON, trích để máy đọc)

```json
{
  "version": "1.4.0", "commit": "a1b2c3d", "verdict": "PASS",
  "hooks": [
    { "name": "egress", "pass": true, "outbound_count": 3, "credential_leaks": 0 },
    { "name": "sbom",   "pass": true, "deps": 41, "high_cve": 0 },
    { "name": "reproducible", "pass": true, "shipped_sha256": "9f2c...e71a" },
    { "name": "payload_guard", "pass": true, "findings": 0 }
  ]
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Thuê hãng audit bên ngoài ký xác nhận chạy lại bộ hook - bổ sung uy tín, nhưng cơ chế tái chạy đã là bằng chứng nền; xét khi có ngân sách.
- Egress test cho TikTok Shop/Lazada (P2) - cùng bộ test, thêm khi content script hai sàn đó ship (FR-EXT-007/008).
- Fuzzing payload đầu vào content script để tìm kênh rò ẩn - lớp kiểm sâu hơn, giai đoạn sau.
- Giám sát egress liên tục ở production (không chỉ trong CI) - cần hạ tầng quan sát phía client; xét ở giai đoạn trưởng thành.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Cookie rời máy qua đường grep không thấy | egress test (động) | rò credential runtime | Kiểm tại biên mạng, không tin self-report (DEC-TRUST-12) |
| Rò qua host lạ (kênh bên) | egress assert host | dữ liệu ra ngoài backend | Whitelist endpoint + assert payload (§1 #4) |
| Dependency độc hại gửi dữ liệu | SBOM + vuln scan | supply-chain compromise | SBOM + pin + quét CVE (DEC-TRUST-13) |
| Báo cáo audit một lần, không tái chạy | review (thiếu guide) | "tin tôi đi", cũ khi mã đổi | Hook tái chạy + THIRD-PARTY guide (DEC-TRUST-11) |
| Audit PASS dù một hook con đỏ | run-security-audit exit code | bằng chứng giả | PASS chỉ khi cả bốn hook PASS (DEC-TRUST-15) |
| Test chạy không có phiên đăng nhập | review fixture | không chứng minh kịch bản thật | Đặt cookie vào context trước khi chạy (§1 #11) |
| Bỏ hook trong "chế độ nhanh" | báo INCOMPLETE | ngầm PASS nguy hiểm | Mọi lần chạy đủ bốn hook hoặc INCOMPLETE (§1 #12) |
| Mã chạy khác mã công khai | verify-reproducible (hook con) | audit trên mã sai | Gọi verify-reproducible của FR-TRUST-001 (§1 #7) |

---

## §11 - Ghi chú

- Egress test là mảnh bằng chứng mạnh nhất: nó quan sát đúng những gì rời máy trong luồng thật có phiên đăng nhập, thay cho mọi lời khai trong mã.
- Audit của SănDeal khác "báo cáo đã được hãng X kiểm" ở chỗ tái chạy được: ai cũng clone + chạy + thấy cùng kết quả, không phải tin uy tín mượn.
- FR-TRUST-003 là lớp tổng hợp: nó gọi reproducible (FR-TRUST-001) + payload_guard (FR-COMPLY-005) như mảnh ghép và thêm egress động + SBOM để thành bằng chứng đầu cuối.
- Whitelist endpoint + assert payload đóng cả hai kênh rò: gửi sai đích VÀ gửi sai nội dung.
- SBOM + pin dependency biến "extension phụ thuộc gì" thành danh sách chính xác, kiểm chứng được - đóng kênh tấn công supply-chain.
- Đây là mắt xích cuối của ba trụ niềm tin (mở source + chính sách + audit) mà §5.4 đặt ra; cùng nhau, chúng hóa giải nỗi lo malware/scam của ~45% người tiêu dùng VN (Ken Research) bằng bằng chứng chứ không bằng lời hứa.

---

*Hết FR-TRUST-003. Status: ready_to_implement (mục tiêu audit 10/10).*
