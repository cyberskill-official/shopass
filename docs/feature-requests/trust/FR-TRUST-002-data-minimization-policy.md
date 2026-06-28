---
id: FR-TRUST-002
title: "Chính sách minh bạch tối thiểu hóa dữ liệu + xử lý local-first - tài liệu hóa đúng dữ liệu thu thập (chỉ productId/price/qty), neo vào allowlist pipeline làm nguồn sự thật máy kiểm"
module: TRUST
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-EXT-003, FR-EXT-006, FR-TRUST-001, FR-TRUST-003, FR-COMPLY-001, FR-COMPLY-005, NFR-TRUST-001]
depends_on: [FR-EXT-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.4 (trust & security: (3) chính sách dữ liệu minh bạch không gửi cookie/mật khẩu; (4) xử lý dữ liệu tối thiểu hóa, local-first)"
  - "docs/... §3.2 (gửi backend dạng tối thiểu hóa: chỉ productId/giá/số lượng), §3.8 (token không rời client)"
source_decisions:
  - "DEC-TRUST-06: chính sách tối thiểu hóa là VĂN BẢN có thể kiểm chứng (machine-checkable), KHÔNG phải tuyên bố marketing; nó tham chiếu allowlist FR-EXT-003 làm nguồn sự thật"
  - "DEC-TRUST-07: local-first - mọi chuẩn hóa/khử trùng dữ liệu xảy ra trên client; backend chỉ nhận payload đã làm sạch (nhất quán DEC-EXT-16); chính sách phải nêu rõ ranh giới này"
  - "DEC-TRUST-08: chính sách liệt kê ĐÚNG tập dữ liệu thu thập {platform, productId, price, qty} (+voucher hiển thị) và mục đích từng trường; cập nhật trường phải cập nhật chính sách (test khóa)"
  - "DEC-TRUST-09: chính sách phải gắn cơ sở pháp lý PDPL (FR-COMPLY-001) cho từng mục đích xử lý; tối thiểu hóa dữ liệu là nguyên tắc PDPL, không chỉ là PR"
  - "DEC-TRUST-10: data-flow diagram (client -> minimize -> queue -> backend) là một phần chính sách; chỉ ra đúng điểm dữ liệu rời máy và điểm nào KHÔNG có dữ liệu nhạy cảm"

language: "Markdown (DATA-POLICY.md, data-flow) + TypeScript test (policy-allowlist parity); tham chiếu Go/SQL của backend"
service: shopass/docs/trust/ + shopass/extension/
new_files:
  - docs/trust/DATA-MINIMIZATION-POLICY.md
  - docs/trust/data-flow.md
  - extension/test/policy-allowlist-parity.test.ts
  - extension/src/policy/collected-fields.ts
modified_files:
  - extension/src/pipeline/allowlist.ts   # export hằng ALLOWED_* để policy test tham chiếu (nguồn sự thật chung)
allowed_tools:
  - file_read: docs/trust/**, extension/**
  - file_write: docs/trust/**, extension/**
  - bash: cd extension && npm test
disallowed_tools:
  - viết chính sách như tuyên bố marketing không nối với code thực thi (vi phạm DEC-TRUST-06)
  - khai trong chính sách tập dữ liệu rộng hơn pipeline thực gửi, hoặc hẹp hơn (vi phạm DEC-TRUST-08)
  - mô tả backend 'lọc giúp' thay vì local-first (vi phạm DEC-TRUST-07)
  - bỏ cơ sở pháp lý PDPL cho mục đích xử lý (vi phạm DEC-TRUST-09)

effort_hours: 5
sub_tasks:
  - "1.5h: DATA-MINIMIZATION-POLICY.md - tập dữ liệu thu thập + mục đích từng trường + cam kết KHÔNG cookie/mật khẩu/token + cơ sở pháp lý PDPL"
  - "0.5h: data-flow.md - sơ đồ luồng client -> minimize -> queue -> backend, chỉ điểm dữ liệu rời máy"
  - "0.5h: collected-fields.ts - hằng nguồn sự thật mô tả trường thu thập + mục đích (dùng cho policy + UI consent FR-EXT-006)"
  - "0.5h: export ALLOWED_* trong allowlist.ts để test tham chiếu"
  - "1.5h: policy-allowlist-parity.test.ts - chính sách <-> allowlist parity; lệch -> fail"
  - "0.5h: liên kết chính sách vào DISCLOSURE.md (FR-TRUST-001) + UI consent (FR-EXT-006) - một nguồn nội dung"

risk_if_skipped: "Tài liệu nguồn (§5.4) yêu cầu rõ 'chính sách dữ liệu minh bạch (không gửi cookie/mật khẩu)' và 'xử lý dữ liệu tối thiểu hóa, local-first' như hai trong năm trụ niềm tin. Nếu chính sách chỉ là văn bản marketing không nối với hành vi thật, nó sẽ trôi khỏi thực tế ngay khi pipeline đổi: người dùng đọc một đằng, extension làm một nẻo - đúng kiểu mất niềm tin mà SănDeal muốn tránh hậu-Honey. Tối thiểu hóa dữ liệu cũng là nguyên tắc PDPL (Luật 91/2025): thu thập vượt mục đích là vi phạm. Neo chính sách vào allowlist FR-EXT-003 qua test biến 'minh bạch' thành ràng buộc máy kiểm: chính sách không thể nói nhiều/ít hơn thực tế. Bỏ FR này, lời hứa tối thiểu hóa thành khẩu hiệu trống - phá nền tảng niềm tin và mời rủi ro PDPL."
---

## §1 - Mô tả (BCP-14 normative)

SănDeal **MUST** công bố một chính sách tối thiểu hóa dữ liệu kiểm chứng được: liệt kê đúng tập dữ liệu thu thập, mục đích từng trường, cam kết KHÔNG gửi cookie/mật khẩu/token, và neo vào allowlist pipeline (FR-EXT-003) làm nguồn sự thật máy kiểm. Hợp đồng:

1. `DATA-MINIMIZATION-POLICY.md` **MUST** liệt kê CHÍNH XÁC tập dữ liệu extension thu thập và gửi đi - đúng tập tối thiểu `{platform, productId, price, qty}` (+ voucher hiển thị) của FR-EXT-003 - không rộng hơn, không hẹp hơn (DEC-TRUST-08).
2. Chính sách **MUST** nêu mục đích xử lý của TỪNG trường (productId để tra cứu sản phẩm, price để theo dõi giá, qty/voucher để tối ưu giỏ) - không có trường nào "thu thập để dành".
3. Chính sách **MUST** khẳng định tường minh KHÔNG thu thập/gửi: cookie phiên sàn, mật khẩu, token phiên/header xác thực, email, số điện thoại, tên, địa chỉ, hay định danh người dùng sàn thật.
4. Chính sách **MUST** mô tả nguyên tắc local-first (DEC-TRUST-07): mọi chuẩn hóa/khử trùng diễn ra trên client; backend chỉ nhận payload đã làm sạch; KHÔNG có nhánh "gửi thô để backend lọc".
5. `data-flow.md` **MUST** trình bày sơ đồ luồng dữ liệu (client -> `minimize` -> queue -> backend) chỉ rõ ĐIỂM dữ liệu rời máy và khẳng định tại điểm đó payload không chứa dữ liệu nhạy cảm (nối FR-EXT-003).
6. Chính sách **MUST** gắn cơ sở pháp lý PDPL (FR-COMPLY-001) cho mỗi mục đích xử lý; nêu rõ tối thiểu hóa dữ liệu là nguyên tắc PDPL (Luật 91/2025), không chỉ là cam kết tự nguyện (DEC-TRUST-09).
7. **MUST** có `collected-fields.ts`: một hằng nguồn sự thật mô tả từng trường thu thập + mục đích, dùng chung cho chính sách, UI consent (FR-EXT-006), và disclosure (FR-TRUST-001) - để ba nơi không thể lệch nhau.
8. **MUST** có test `policy-allowlist-parity.test.ts` khẳng định: tập trường mô tả trong `collected-fields.ts` khớp đúng allowlist `OutboundPayload` của FR-EXT-003. Thêm/bớt trường ở allowlist mà không cập nhật chính sách -> test fail.
9. Chính sách **MUST** nêu chính sách lưu trữ/xóa (retention) cho dữ liệu thu thập, nhất quán với quyền chủ thể dữ liệu (FR-COMPLY-003, DSAR).
10. Chính sách **MUST** dùng ngôn ngữ người dùng phổ thông hiểu được (không chỉ thuật ngữ kỹ thuật) - vì đối tượng là người tiêu dùng VN lo ngại lộ dữ liệu (§5.4).
11. **MUST NOT** chứa tuyên bố không thể kiểm chứng hoặc rộng hơn hành vi thật (ví dụ "chúng tôi không bao giờ chia sẻ dữ liệu" nếu B2B aggregate FR-B2B-001 có chia sẻ dạng ẩn danh - phải mô tả đúng, kèm điều kiện k-anonymity).
12. Chính sách **MUST** được tham chiếu từ DISCLOSURE.md (FR-TRUST-001) và UI consent (FR-EXT-006) - một nguồn nội dung duy nhất, tránh ba bản lệch nhau.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao chính sách phải máy kiểm được (DEC-TRUST-06)?** Chính sách quyền riêng tư thông thường là văn bản pháp lý tách rời code - nó đúng ngày viết rồi trôi dần khỏi thực tế khi sản phẩm đổi. SănDeal đảo cách làm: chính sách neo vào allowlist thực thi (FR-EXT-003) qua test, nên nếu pipeline đổi tập dữ liệu mà chính sách không cập nhật, CI đỏ. "Minh bạch" trở thành ràng buộc kỹ thuật, không phải thiện chí.

**Vì sao một nguồn sự thật chung `collected-fields.ts` (§1 #7)?** Cùng một thông tin xuất hiện ba nơi: chính sách, UI consent lúc cài (FR-EXT-006), disclosure store (FR-TRUST-001). Nếu mỗi nơi viết tay, chúng lệch nhau theo thời gian. Một hằng chung làm cả ba dẫn xuất từ một chỗ - sửa một lần, đồng bộ mọi nơi.

**Vì sao local-first là một phần chính sách (DEC-TRUST-07)?** "Backend lọc giúp" nghĩa là dữ liệu thô đã rời máy - quá muộn. Triết lý SănDeal (§5.4) là xử lý ngay trên client, chỉ kết quả sạch mới truyền đi. Chính sách phải nói rõ ranh giới này để người dùng hiểu dữ liệu nhạy cảm không bao giờ chạm server, và để data-flow diagram chỉ đúng điểm rời máy.

**Vì sao gắn cơ sở pháp lý PDPL (DEC-TRUST-09)?** Tối thiểu hóa dữ liệu không chỉ là điểm bán niềm tin - nó là nghĩa vụ pháp lý theo PDPL (Luật 91/2025): xử lý phải đúng mục đích, đơn mục đích, không thu thập thừa. Gắn mỗi mục đích với cơ sở pháp lý làm chính sách vừa minh bạch vừa phòng thủ pháp lý, nối với khung consent FR-COMPLY-001.

**Vì sao cấm tuyên bố rộng hơn hành vi (§1 #11)?** Hứa "không bao giờ chia sẻ dữ liệu" rồi B2B (FR-B2B-001) bán insight ẩn danh là mâu thuẫn phá niềm tin. Chính sách phải mô tả đúng - kể cả phần chia sẻ ẩn danh hóa với điều kiện k-anonymity - thay vì hứa tuyệt đối rồi vi phạm. Trung thực có điều kiện bền hơn tuyệt đối giả.

**Vì sao ngôn ngữ phổ thông (§1 #10)?** Đối tượng là người tiêu dùng VN, không phải luật sư hay kỹ sư. Một chính sách chỉ toàn thuật ngữ thì minh bạch trên giấy nhưng mờ với người đọc. Viết để người mua hàng hiểu được chính là tinh thần "minh bạch" thật của §5.4.

---

## §3 - Hợp đồng API / DDL

### collected-fields.ts (nguồn sự thật mô tả trường)

```ts
// extension/src/policy/collected-fields.ts
export interface CollectedField {
  field: string;        // tên trường khớp allowlist OutboundPayload (FR-EXT-003)
  purpose: string;      // mục đích xử lý (ngôn ngữ người dùng)
  legalBasis: string;   // cơ sở pháp lý PDPL (FR-COMPLY-001)
}

export const COLLECTED_FIELDS: CollectedField[] = [
  { field: "platform", purpose: "Biết bạn đang xem sàn nào để tra đúng dữ liệu giá", legalBasis: "Đồng thuận - mục đích theo dõi giá" },
  { field: "productId", purpose: "Tra cứu sản phẩm để hiện lịch sử giá và sale ảo", legalBasis: "Đồng thuận - mục đích theo dõi giá" },
  { field: "price",     purpose: "Theo dõi biến động giá để cảnh báo và vẽ biểu đồ", legalBasis: "Đồng thuận - mục đích theo dõi giá" },
  { field: "qty",       purpose: "Tính tối ưu voucher/giỏ hàng cho bạn",              legalBasis: "Đồng thuận - mục đích tối ưu giỏ" },
];

// KHÔNG bao giờ thu thập: cookie, mật khẩu, token phiên sàn, email, SĐT, tên, địa chỉ.
export const NEVER_COLLECTED = [
  "cookie", "mật khẩu", "token phiên sàn", "header xác thực",
  "email", "số điện thoại", "tên", "địa chỉ", "định danh người dùng sàn",
] as const;
```

### policy-allowlist-parity.test.ts (chính sách phải khớp pipeline)

```ts
// extension/test/policy-allowlist-parity.test.ts
import { COLLECTED_FIELDS } from "../src/policy/collected-fields";
import { ALLOWED_ITEM_FIELDS } from "../src/pipeline/allowlist";

test("mọi trường item ở allowlist đều được mô tả trong chính sách", () => {
  const described = new Set(COLLECTED_FIELDS.map(f => f.field));
  for (const f of ALLOWED_ITEM_FIELDS) expect(described.has(f)).toBe(true);
});

test("chính sách KHÔNG mô tả trường nào pipeline không gửi (không phóng đại/giấu)", () => {
  const allowed = new Set<string>(["platform", ...ALLOWED_ITEM_FIELDS]);
  for (const f of COLLECTED_FIELDS) expect(allowed.has(f.field)).toBe(true);
});

test("mọi trường thu thập đều có mục đích + cơ sở pháp lý (không 'thu để dành')", () => {
  for (const f of COLLECTED_FIELDS) {
    expect(f.purpose.length).toBeGreaterThan(0);
    expect(f.legalBasis.length).toBeGreaterThan(0);
  }
});
```

### data-flow.md (cấu trúc bắt buộc)

```markdown
# SănDeal - Luồng dữ liệu (data flow)

Trang sàn (DOM)
   │  content script đọc giá/giỏ (FR-EXT-002)
   ▼
[minimize] allowlist + redact + validate  (FR-EXT-003)  <-- CHẠY TRÊN CLIENT
   │  CHỈ {platform, productId, price, qty} (+voucher) đi tiếp; cookie/PII bị loại
   ▼  <===== ĐÂY là điểm DỮ LIỆU RỜI MÁY (đã sạch) =====
hàng đợi đồng bộ (FR-EXT-005) -> đính JWT SănDeal (KHÔNG token sàn)
   ▼
Backend price-svc (lưu price_snapshot)

KHÔNG có mũi tên nào mang cookie/mật khẩu/token sàn rời máy.
```

---

## §4 - Acceptance criteria

1. `DATA-MINIMIZATION-POLICY.md` liệt kê đúng `{platform, productId, price, qty}` (+ voucher) và mục đích từng trường.
2. Chính sách khẳng định tường minh KHÔNG thu thập cookie/mật khẩu/token/email/SĐT/tên/địa chỉ.
3. Chính sách mô tả local-first: chuẩn hóa trên client, backend chỉ nhận payload sạch, không "gửi thô để backend lọc".
4. `data-flow.md` có sơ đồ chỉ rõ điểm dữ liệu rời máy và khẳng định tại đó không có dữ liệu nhạy cảm.
5. Chính sách gắn cơ sở pháp lý PDPL cho mỗi mục đích (nối FR-COMPLY-001).
6. `collected-fields.ts` mô tả từng trường + purpose + legalBasis; được tham chiếu bởi chính sách, UI consent, disclosure.
7. `policy-allowlist-parity.test.ts`: mọi trường allowlist (FR-EXT-003) được mô tả; KHÔNG mô tả trường pipeline không gửi.
8. Thêm trường mới vào allowlist mà không cập nhật `collected-fields.ts` -> test đỏ.
9. Chính sách nêu retention/xóa nhất quán FR-COMPLY-003 (DSAR).
10. Chính sách dùng ngôn ngữ phổ thông (không chỉ thuật ngữ); rà soát đọc hiểu.
11. KHÔNG có tuyên bố tuyệt đối mâu thuẫn với B2B aggregate ẩn danh (FR-B2B-001) - mô tả đúng kèm điều kiện.
12. Chính sách được tham chiếu từ DISCLOSURE.md (FR-TRUST-001) và UI consent (FR-EXT-006); `npm test` xanh.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/policy-allowlist-parity.test.ts (xem §3)
// Khẳng định parity hai chiều: chính sách không nói nhiều/ít hơn pipeline thật.
```

```bash
# Kiểm chính sách có đủ các cam kết "KHÔNG thu thập"
for k in cookie "mật khẩu" "token phiên" email "số điện thoại"; do
  grep -qi "$k" docs/trust/DATA-MINIMIZATION-POLICY.md || { echo "thiếu cam kết: $k"; exit 1; }
done

# Kiểm data-flow có chỉ điểm dữ liệu rời máy
grep -qi "rời máy" docs/trust/data-flow.md

# Kiểm parity policy <-> allowlist
cd extension && npm test -- policy-allowlist-parity
```

```ts
// extension/test/policy-allowlist-parity.test.ts (phần PDPL)
import { COLLECTED_FIELDS, NEVER_COLLECTED } from "../src/policy/collected-fields";

test("danh sách NEVER_COLLECTED bao gồm các loại nhạy cảm cốt lõi", () => {
  for (const must of ["cookie", "mật khẩu", "token phiên sàn"])
    expect(NEVER_COLLECTED).toContain(must as any);
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `collected-fields.ts` (định nghĩa nguồn sự thật trường + mục đích + cơ sở pháp lý) + export `ALLOWED_*` từ `allowlist.ts` -> `policy-allowlist-parity.test.ts` (khóa chính sách vào pipeline) -> `DATA-MINIMIZATION-POLICY.md` (dẫn xuất nội dung từ `collected-fields.ts`, thêm local-first + retention + cơ sở pháp lý PDPL) -> `data-flow.md` (sơ đồ điểm rời máy) -> liên kết vào DISCLOSURE.md (FR-TRUST-001) + UI consent (FR-EXT-006). Quy ước: mọi thay đổi tập dữ liệu phải sửa `collected-fields.ts` + allowlist cùng lúc; test parity là cổng giữ hai bên đồng bộ.

---

## §7 - Phụ thuộc

- **FR-EXT-003** - allowlist `OutboundPayload` là nguồn sự thật cho tập dữ liệu; chính sách + test parity neo vào đây.
- **FR-EXT-006** - UI consent lúc cài dùng `collected-fields.ts` cho nội dung minh bạch khi xin đồng thuận.
- **FR-TRUST-001 (đồng cấp)** - DISCLOSURE.md trên store dẫn xuất cùng nguồn nội dung; reproducible build cho phép kiểm chứng chính sách khớp mã.
- **FR-TRUST-003 (downstream)** - audit độc lập dùng chính sách + data-flow + parity test làm tài liệu tham chiếu khi kiểm KHÔNG gửi cookie/PII.
- **FR-COMPLY-001** - khung consent PDPL cung cấp cơ sở pháp lý cho từng mục đích xử lý nêu trong chính sách.
- **FR-COMPLY-003** - DSAR (quyền truy cập/xóa) khớp mục retention của chính sách.
- **FR-COMPLY-005** - cưỡng chế no-cleartext/token; chính sách khẳng định đúng bất biến mà COMPLY-005 kiểm máy.

---

## §8 - Payload ví dụ

### Trích DATA-MINIMIZATION-POLICY.md (đoạn dữ liệu thu thập)

```markdown
## Chúng tôi thu thập gì và vì sao
| Dữ liệu | Vì sao cần | Cơ sở pháp lý (PDPL) |
|---|---|---|
| Sàn đang xem | Tra đúng dữ liệu giá sàn đó | Đồng thuận - theo dõi giá |
| ID sản phẩm | Hiện lịch sử giá + cảnh báo sale ảo | Đồng thuận - theo dõi giá |
| Giá hiển thị | Vẽ biểu đồ + cảnh báo giảm giá | Đồng thuận - theo dõi giá |
| Số lượng trong giỏ | Tối ưu voucher/giỏ cho bạn | Đồng thuận - tối ưu giỏ |

## Chúng tôi KHÔNG BAO GIỜ thu thập
Cookie đăng nhập, mật khẩu, token phiên, email, số điện thoại, tên, địa chỉ.
Dữ liệu được làm sạch ngay trên máy bạn (local-first); chỉ thông tin sản phẩm
ở trên mới rời máy. Bạn có thể tự kiểm bằng mã nguồn mở (xem DISCLOSURE).
```

### Sự kiện CI khi pipeline đổi mà chính sách quên cập nhật

```
FAIL policy-allowlist-parity.test.ts
  > allowlist thêm trường "rating" nhưng collected-fields.ts chưa mô tả
  -> cập nhật collected-fields.ts (purpose + legalBasis) rồi mới merge được
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Bản dịch chính sách sang tiếng Anh + ngôn ngữ SEA (ID/TH) - khi mở rộng per-country (FR-COMPLY-007); cùng cơ chế parity.
- Mức chi tiết retention cụ thể (số ngày giữ price_snapshot quy cho user) - căn theo FR-PRICE-002 (raw 18 tháng) + FR-COMPLY-003; tinh ở giai đoạn DSAR.
- Mô tả B2B aggregate ẩn danh chi tiết - bổ sung khi FR-B2B-001 chốt ngưỡng k-anonymity; FR này chỉ cấm tuyên bố tuyệt đối mâu thuẫn.
- Chính sách cho dữ liệu mobile (FR-MOBILE-001) - mở rộng cùng khung khi mobile ship (P3).

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Chính sách trôi khỏi hành vi thật | policy-allowlist-parity test đỏ | minh bạch giả, mất niềm tin | Neo chính sách vào allowlist qua test (DEC-TRUST-06) |
| Khai tập dữ liệu rộng hơn thực | parity test (chiều ngược) | tự buộc tội thu thừa | Test cấm mô tả trường pipeline không gửi (§1 #1) |
| Khai hẹp hơn thực (giấu trường) | parity test (chiều xuôi) | lừa dối người dùng | Test bắt mọi trường allowlist phải được khai (§4 #7) |
| Ba bản (policy/consent/disclosure) lệch nhau | review + nguồn chung | thông điệp mâu thuẫn | Một `collected-fields.ts` cho cả ba (§1 #7) |
| Mô tả "backend lọc giúp" | review chính sách | sai bản chất local-first | Chính sách + data-flow chỉ đúng điểm rời máy (DEC-TRUST-07) |
| Thiếu cơ sở pháp lý PDPL | review + field.legalBasis | yếu phòng thủ pháp lý | Mỗi mục đích gắn cơ sở pháp lý (DEC-TRUST-09) |
| Tuyên bố tuyệt đối mâu thuẫn B2B | review | hứa rồi vi phạm | Mô tả đúng kèm điều kiện k-anonymity (§1 #11) |
| Ngôn ngữ quá kỹ thuật | rà đọc hiểu | minh bạch trên giấy, mờ với user | Viết ngôn ngữ phổ thông (§1 #10) |

---

## §11 - Ghi chú

- Chính sách tối thiểu hóa của SănDeal khác chính sách quyền riêng tư thông thường ở một điểm: nó được nối với code thực thi qua test, nên không thể nói nhiều/ít hơn hành vi thật.
- `collected-fields.ts` là một nguồn sự thật cho ba bề mặt (chính sách, consent, disclosure) - sửa một lần, đồng bộ mọi nơi, không lệch.
- Local-first không phải khẩu hiệu: data-flow diagram chỉ đúng điểm dữ liệu rời máy, và tại điểm đó payload đã qua minimize nên không có cookie/PII.
- Tối thiểu hóa dữ liệu vừa là điểm bán niềm tin vừa là nghĩa vụ PDPL - gắn cơ sở pháp lý cho mỗi mục đích làm chính sách phòng thủ được cả về niềm tin lẫn pháp lý.
- Trung thực có điều kiện (mô tả đúng cả phần B2B ẩn danh) bền hơn hứa tuyệt đối rồi vi phạm.
- FR-TRUST-003 (audit độc lập) dùng chính sách này làm tài liệu đối chiếu khi chứng minh extension KHÔNG gửi cookie/mật khẩu.

---

*Hết FR-TRUST-002. Status: ready_to_implement (mục tiêu audit 10/10).*
