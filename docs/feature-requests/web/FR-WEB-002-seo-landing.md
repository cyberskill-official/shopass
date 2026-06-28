---
id: FR-WEB-002
title: "Landing SEO - trang keyword (cách săn xu Shopee, lịch sale, mã freeship, sale thật hay sale ảo) render SSG/SSR + meta tags + structured data, kéo traffic organic giải bài toán GTM Phase 1"
module: WEB
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-WEB-001, FR-WEB-003, FR-DEAL-001, FR-DEAL-003]
depends_on: [FR-WEB-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.7 (GTM & funnel: site SEO keyword cách săn xu Shopee, lịch sale, mã freeship, sale thật hay sale ảo)"
  - "docs/... §8 (Phase 1 cold-start + SEO trước extension), §3.1 (web app Next.js)"
source_decisions:
  - "DEC-WEB-06: mỗi cụm keyword là một route tĩnh sinh sẵn (SSG) ở build time qua generateStaticParams; nội dung định hướng tìm kiếm, có thể revalidate (ISR) khi dữ liệu phụ (lịch sale) đổi"
  - "DEC-WEB-07: mỗi trang keyword MUST có metadata SEO đầy đủ qua Metadata API của Next (title, description, canonical, OpenGraph, Twitter card) + lang vi-VN"
  - "DEC-WEB-08: trang phát structured data JSON-LD (Article + FAQPage cho trang hỏi-đáp như sale thật hay sale ảo; ItemList cho lịch sale) để hưởng rich result"
  - "DEC-WEB-09: sinh sitemap.xml + robots.txt qua route file của Next (app/sitemap.ts, app/robots.ts) liệt kê mọi trang keyword - không soạn thủ công"
  - "DEC-WEB-10: trang landing là công khai (route group ngoài (app)) - KHÔNG yêu cầu đăng nhập, KHÔNG gọi API người dùng; chỉ nội dung tĩnh + link CTA tới đăng ký/cài extension"

language: "TypeScript 5.x; Next.js 14 (App Router, SSG/ISR, Metadata API, generateStaticParams); JSON-LD structured data"
service: shopass/web/
new_files:
  - web/app/(marketing)/[keyword]/page.tsx
  - web/app/(marketing)/[keyword]/content.ts
  - web/app/(marketing)/layout.tsx
  - web/app/sitemap.ts
  - web/app/robots.ts
  - web/lib/seo/jsonld.ts
  - web/lib/seo/keywords.ts
  - web/test/seo-metadata.test.ts
  - web/test/seo-jsonld.test.ts
  - web/test/sitemap.test.ts
modified_files:
  - web/next.config.mjs                      # bật ISR revalidate cho route marketing nếu cần
allowed_tools:
  - file_read: web/**
  - file_write: web/**
  - bash: cd web && npm test && npx tsc --noEmit
disallowed_tools:
  - render trang keyword phía client-only (CSR) làm crawler không thấy nội dung (vi phạm DEC-WEB-06, hỏng SEO)
  - bỏ canonical / title / description trên trang keyword (vi phạm DEC-WEB-07)
  - gọi API người dùng hoặc yêu cầu đăng nhập trên landing công khai (vi phạm DEC-WEB-10)
  - soạn sitemap/robots thủ công lệch danh sách keyword thật (vi phạm DEC-WEB-09)

effort_hours: 8
sub_tasks:
  - "1.0h: lib/seo/keywords.ts - danh sách cụm keyword nguồn (4 cụm lõi §5.7) + slug + intent"
  - "1.5h: app/(marketing)/[keyword]/content.ts + page.tsx - render SSG nội dung định hướng tìm kiếm"
  - "1.5h: generateStaticParams + generateMetadata (title/description/canonical/OG/Twitter) per keyword (DEC-WEB-07)"
  - "1.5h: lib/seo/jsonld.ts - sinh Article/FAQPage/ItemList JSON-LD; nhúng <script type=application/ld+json>"
  - "1.0h: app/sitemap.ts + app/robots.ts sinh tự động từ keywords.ts (DEC-WEB-09)"
  - "0.5h: app/(marketing)/layout.tsx - layout công khai + CTA đăng ký/cài extension, KHÔNG auth (DEC-WEB-10)"
  - "1.0h: test/seo-metadata.test.ts - title/description/canonical tồn tại + đúng per keyword"
  - "1.0h: test/seo-jsonld.test.ts + sitemap.test.ts - JSON-LD hợp lệ schema.org; sitemap liệt kê đủ keyword"

risk_if_skipped: "Theo §5.7 và §8, SEO là kênh GTM Phase 1 - chạy TRƯỚC khi ra extension để kéo traffic organic và giải đồng thời bài toán cold-start (vừa scraping tích lũy giá, vừa hút người dùng tự nhiên). CAC qua SEO/viral thấp (~20-50k VND, §4.1) là giả định nền của unit economics - không có SEO thì phải mua traffic, phá payback <2 tháng. Nếu render trang keyword client-only thì crawler Google không thấy nội dung và toàn bộ nỗ lực SEO vô nghĩa. Nếu thiếu meta tags/structured data thì mất rich result và thứ hạng - các cụm keyword cạnh tranh (cách săn xu Shopee, sale thật hay sale ảo) chính là nơi BeeCost đã xây thương hiệu (§5.6), SănDeal phải thắng trận SEO này để có cửa. Nếu landing yêu cầu đăng nhập thì funnel đứt ngay bước đầu - người tìm kiếm rời đi trước khi thấy giá trị."
---

## §1 - Mô tả (BCP-14 normative)

Web app **MUST** phục vụ một tập trang landing tối ưu SEO cho các cụm keyword GTM (§5.7), render phía server (SSG/ISR) để crawler đọc được, kèm metadata đầy đủ và structured data JSON-LD, hoàn toàn công khai không yêu cầu đăng nhập. Hợp đồng:

1. Web app **MUST** sinh một trang tĩnh cho mỗi cụm keyword lõi (§5.7): tối thiểu "cách săn xu Shopee", "lịch sale", "mã freeship", "sale thật hay sale ảo" - mỗi cụm là một route với slug riêng dưới group `(marketing)` (DEC-WEB-06).
2. Mỗi trang keyword **MUST** render SSG ở build time qua `generateStaticParams` (hoặc ISR với `revalidate` khi nội dung phụ thuộc dữ liệu đổi theo thời gian như lịch sale) - KHÔNG render client-only (DEC-WEB-06). Nội dung HTML phải có sẵn trong response đầu tiên cho crawler.
3. Mỗi trang **MUST** xuất metadata SEO đầy đủ qua Metadata API của Next (`generateMetadata`): `title`, `description`, `alternates.canonical`, `openGraph` (title/description/type/locale), `twitter` (card), và `lang="vi-VN"` (DEC-WEB-07). Mỗi keyword có title + description riêng, không trùng lặp.
4. Mỗi trang **MUST** nhúng structured data JSON-LD phù hợp loại nội dung (DEC-WEB-08): `FAQPage` cho trang hỏi-đáp ("sale thật hay sale ảo"), `ItemList` cho "lịch sale", `Article` cho trang hướng dẫn - dưới `<script type="application/ld+json">`, đúng schema.org.
5. Web app **MUST** sinh `sitemap.xml` qua `app/sitemap.ts` và `robots.txt` qua `app/robots.ts`, liệt kê tự động mọi trang keyword từ cùng nguồn `keywords.ts` (DEC-WEB-09) - KHÔNG soạn thủ công để tránh lệch.
6. Trang landing **MUST** công khai: thuộc group `(marketing)` ngoài `(app)`, KHÔNG qua middleware auth (FR-WEB-001 #7), KHÔNG gọi API người dùng, KHÔNG yêu cầu đăng nhập (DEC-WEB-10).
7. Mỗi trang **MUST** có ít nhất một CTA dẫn tới bước tiếp funnel (đăng ký free hoặc cài extension) - khớp chuỗi GTM §5.7 (SEO -> signup free -> Premium upsell).
8. `canonical` URL **MUST** trỏ về chính trang đó trên domain chuẩn (tránh trùng nội dung; mỗi keyword một canonical duy nhất).
9. Trang **MUST** đặt thẻ `lang` và locale `vi-VN`; nội dung tiếng Việt; cấu trúc heading hợp lệ (một `h1` mô tả keyword chính).
10. Nội dung keyword **MUST** trung thực với định vị SănDeal: không hứa hẹn quá mức, gắn với năng lực thật (phát hiện sale ảo FR-DEAL-001, biểu đồ giá FR-WEB-003); tránh nội dung spam/cloaking (rủi ro phạt SEO + lệch thông điệp niềm tin).
11. `app/sitemap.ts` **MUST** trả đúng định dạng `MetadataRoute.Sitemap` với `lastModified`; `app/robots.ts` cho phép crawl trang công khai và trỏ tới sitemap.
12. Toàn bộ **MUST** vượt `npx tsc --noEmit` sạch và `npm test` xanh.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao SSG/ISR thay vì CSR (DEC-WEB-06)?** SEO sống chết ở việc crawler đọc được nội dung trong HTML đầu tiên. CSR (render bằng JavaScript phía client) làm crawler thấy trang trống rồi mới chạy JS - rủi ro không index hoặc index chậm. SSG sinh HTML đầy đủ ở build time; ISR cho phép làm tươi trang phụ thuộc thời gian (lịch sale) mà vẫn giữ HTML tĩnh phục vụ nhanh. Đây là khác biệt giữa SEO hoạt động và SEO vô nghĩa.

**Vì sao metadata đầy đủ qua Metadata API (DEC-WEB-07)?** `title` + `description` quyết định cách trang hiện trên kết quả tìm kiếm và tỷ lệ click; `canonical` tránh phạt trùng nội dung; OpenGraph/Twitter card quyết định cách trang hiện khi chia sẻ (quan trọng cho virality §5.7). Mỗi keyword cần bộ metadata riêng - dùng `generateMetadata` per route đảm bảo không trùng.

**Vì sao JSON-LD structured data (DEC-WEB-08)?** Rich result (hộp FAQ, danh sách) tăng diện tích hiển thị và tỷ lệ click trên SERP. "Sale thật hay sale ảo" là câu hỏi tự nhiên - đánh dấu `FAQPage` cho cơ hội hộp hỏi-đáp; "lịch sale" là danh sách - `ItemList`. Đây là đòn bẩy SEO rẻ mà hiệu quả cao.

**Vì sao sinh sitemap/robots tự động (DEC-WEB-09)?** Sitemap soạn tay lệch danh sách keyword thật khi thêm/bớt trang - trang mới không được index, trang cũ trỏ sai. Sinh từ cùng nguồn `keywords.ts` đảm bảo sitemap luôn khớp tập trang thực tế.

**Vì sao landing công khai tuyệt đối (DEC-WEB-10)?** Người tìm "cách săn xu Shopee" chưa có tài khoản. Bắt đăng nhập là đứt funnel ngay bước đầu - họ rời đi. Landing phải mở, cho thấy giá trị, rồi mới mời đăng ký (CTA). Đồng thời trang công khai không gọi API người dùng nên không có rủi ro rò rỉ dữ liệu.

**Vì sao nội dung trung thực, không spam (§1 #10)?** Cloaking/spam SEO bị Google phạt và mâu thuẫn định vị niềm tin hậu-Honey của SănDeal. Nội dung phải gắn năng lực thật (sale ảo, biểu đồ) - vừa bền về SEO, vừa nhất quán thông điệp.

---

## §3 - Hợp đồng API / DDL

### Nguồn keyword (lib/seo/keywords.ts)

```ts
// web/lib/seo/keywords.ts
export type KeywordIntent = "guide" | "faq" | "list";

export interface KeywordPage {
  slug: string;            // dùng cho route + canonical + sitemap
  keyword: string;         // cụm keyword chính (h1)
  title: string;           // <title> SEO
  description: string;     // meta description
  intent: KeywordIntent;   // quyết loại JSON-LD
}

export const KEYWORD_PAGES: KeywordPage[] = [
  { slug: "cach-san-xu-shopee", keyword: "Cách săn xu Shopee",
    title: "Cách săn xu Shopee hiệu quả 2026 | SănDeal",
    description: "Hướng dẫn săn xu Shopee đúng cách, checklist nhắc nhở, không tự động click rủi ro tài khoản.",
    intent: "guide" },
  { slug: "lich-sale", keyword: "Lịch sale Shopee TikTok Lazada",
    title: "Lịch sale 3 sàn mới nhất - ngày đôi, payday | SănDeal",
    description: "Lịch các đợt sale lớn 1.1 đến 12.12 và payday trên Shopee, TikTok Shop, Lazada.",
    intent: "list" },
  { slug: "ma-freeship", keyword: "Mã freeship",
    title: "Mã freeship Shopee TikTok Lazada hôm nay | SănDeal",
    description: "Tổng hợp mã freeship và cách tối ưu freeship khi thanh toán đa sàn.",
    intent: "guide" },
  { slug: "sale-that-hay-sale-ao", keyword: "Sale thật hay sale ảo",
    title: "Sale thật hay sale ảo? Cách nhận biết | SănDeal",
    description: "Phân biệt sale thật và sale ảo bằng lịch sử giá 90 ngày, median90 và đáy giá.",
    intent: "faq" },
];
```

### JSON-LD (lib/seo/jsonld.ts)

```ts
// web/lib/seo/jsonld.ts
import type { KeywordPage } from "./keywords";

export function buildJsonLd(p: KeywordPage, url: string): object {
  if (p.intent === "faq") {
    return {
      "@context": "https://schema.org", "@type": "FAQPage",
      mainEntity: [{
        "@type": "Question", name: "Làm sao biết sale thật hay sale ảo?",
        acceptedAnswer: { "@type": "Answer",
          text: "So giá hiện tại với trung vị 90 ngày và đáy giá lịch sử; giá gốc bị thổi + giảm không thật là sale ảo." },
      }],
    };
  }
  if (p.intent === "list") {
    return { "@context": "https://schema.org", "@type": "ItemList", name: p.keyword, url };
  }
  return { "@context": "https://schema.org", "@type": "Article", headline: p.keyword, url, inLanguage: "vi-VN" };
}
```

### Metadata + page (app/(marketing)/[keyword]/page.tsx)

```ts
// web/app/(marketing)/[keyword]/page.tsx
import type { Metadata } from "next";
import { KEYWORD_PAGES } from "@/lib/seo/keywords";
import { buildJsonLd } from "@/lib/seo/jsonld";

const SITE = "https://sandeal.vn";

export function generateStaticParams() {
  return KEYWORD_PAGES.map((p) => ({ keyword: p.slug })); // SSG mọi keyword (DEC-WEB-06)
}

export function generateMetadata({ params }: { params: { keyword: string } }): Metadata {
  const p = KEYWORD_PAGES.find((k) => k.slug === params.keyword)!;
  const url = `${SITE}/${p.slug}`;
  return {
    title: p.title, description: p.description,
    alternates: { canonical: url },                         // DEC-WEB-07
    openGraph: { title: p.title, description: p.description, url, type: "article", locale: "vi_VN" },
    twitter: { card: "summary_large_image", title: p.title, description: p.description },
  };
}

export default function KeywordPage({ params }: { params: { keyword: string } }) {
  const p = KEYWORD_PAGES.find((k) => k.slug === params.keyword)!;
  const url = `${SITE}/${p.slug}`;
  return (
    <article lang="vi-VN">
      <h1>{p.keyword}</h1>
      {/* nội dung định hướng tìm kiếm, gắn năng lực thật (FR-DEAL-001 / FR-WEB-003) */}
      <a href="/login?signup=1">Dùng SănDeal miễn phí</a>{/* CTA funnel §5.7 */}
      <script type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(buildJsonLd(p, url)) }} />
    </article>
  );
}
```

### sitemap + robots

```ts
// web/app/sitemap.ts
import type { MetadataRoute } from "next";
import { KEYWORD_PAGES } from "@/lib/seo/keywords";

export default function sitemap(): MetadataRoute.Sitemap {
  return KEYWORD_PAGES.map((p) => ({
    url: `https://sandeal.vn/${p.slug}`, lastModified: new Date(),
  })); // DEC-WEB-09: sinh từ cùng nguồn keyword
}

// web/app/robots.ts
import type { MetadataRoute } from "next";
export default function robots(): MetadataRoute.Robots {
  return { rules: { userAgent: "*", allow: "/" }, sitemap: "https://sandeal.vn/sitemap.xml" };
}
```

---

## §4 - Acceptance criteria

1. `generateStaticParams` trả đúng một entry cho mỗi cụm keyword lõi (tối thiểu 4 cụm §5.7); route SSG, HTML có nội dung trong response đầu.
2. Mỗi trang có `title` + `description` riêng (không trùng) qua `generateMetadata`; có `alternates.canonical` trỏ chính trang đó.
3. Mỗi trang có `openGraph` (type article, locale vi_VN) + `twitter` card.
4. Trang `faq` ("sale thật hay sale ảo") nhúng JSON-LD `FAQPage`; "lịch sale" nhúng `ItemList`; trang guide nhúng `Article`; JSON-LD parse hợp lệ.
5. `app/sitemap.ts` trả `MetadataRoute.Sitemap` chứa URL của TẤT CẢ keyword trong `keywords.ts`; `app/robots.ts` trỏ tới sitemap.
6. Trang landing thuộc group `(marketing)`, KHÔNG khớp matcher middleware auth của FR-WEB-001; grep: không gọi `apiFetch` người dùng trên trang keyword.
7. Mỗi trang có ít nhất một CTA link tới đăng ký/cài extension (funnel §5.7).
8. Mỗi trang có đúng một `h1` chứa keyword chính; `lang="vi-VN"`.
9. Thêm/bớt một keyword trong `keywords.ts` tự động phản ánh trong route, sitemap, và metadata - không phải sửa thủ công nơi khác.
10. `npx tsc --noEmit` sạch; `npm test` xanh.

---

## §5 - Kiểm thử (verification)

```ts
// web/test/seo-metadata.test.ts
import { generateMetadata, generateStaticParams } from "../app/(marketing)/[keyword]/page";
import { KEYWORD_PAGES } from "../lib/seo/keywords";

test("SSG sinh param cho mọi keyword lõi", () => {
  const params = generateStaticParams();
  expect(params.length).toBe(KEYWORD_PAGES.length);
  expect(params.map((p) => p.keyword)).toContain("sale-that-hay-sale-ao");
});

test("mỗi keyword có title/description/canonical riêng", () => {
  const seen = new Set<string>();
  for (const p of KEYWORD_PAGES) {
    const meta = generateMetadata({ params: { keyword: p.slug } });
    expect(meta.title).toBeTruthy();
    expect(meta.description).toBeTruthy();
    expect((meta.alternates as any).canonical).toContain(p.slug);
    expect(seen.has(meta.title as string)).toBe(false); // không trùng title
    seen.add(meta.title as string);
  }
});
```

```ts
// web/test/seo-jsonld.test.ts
import { buildJsonLd } from "../lib/seo/jsonld";
import { KEYWORD_PAGES } from "../lib/seo/keywords";

test("trang FAQ sinh FAQPage; lịch sale sinh ItemList", () => {
  const faq = KEYWORD_PAGES.find((p) => p.intent === "faq")!;
  const list = KEYWORD_PAGES.find((p) => p.intent === "list")!;
  expect((buildJsonLd(faq, "https://sandeal.vn/x") as any)["@type"]).toBe("FAQPage");
  expect((buildJsonLd(list, "https://sandeal.vn/y") as any)["@type"]).toBe("ItemList");
});

test("JSON-LD có @context schema.org", () => {
  for (const p of KEYWORD_PAGES) {
    const ld = buildJsonLd(p, "https://sandeal.vn/z") as any;
    expect(ld["@context"]).toBe("https://schema.org");
  }
});
```

```ts
// web/test/sitemap.test.ts
import sitemap from "../app/sitemap";
import { KEYWORD_PAGES } from "../lib/seo/keywords";

test("sitemap liệt kê đủ mọi keyword", () => {
  const urls = sitemap().map((e) => e.url);
  for (const p of KEYWORD_PAGES) {
    expect(urls).toContain(`https://sandeal.vn/${p.slug}`);
  }
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `lib/seo/keywords.ts` (nguồn sự thật cụm keyword) -> `lib/seo/jsonld.ts` (sinh structured data theo intent) -> `app/(marketing)/[keyword]/page.tsx` (`generateStaticParams` + `generateMetadata` + render SSG + JSON-LD + CTA) -> `app/(marketing)/layout.tsx` (layout công khai, không auth) -> `app/sitemap.ts` + `app/robots.ts` (sinh tự động từ keywords) -> tests. Trang marketing nằm ngoài matcher middleware của FR-WEB-001 nên không bị guard. ISR (`export const revalidate = ...`) bật cho trang phụ thuộc thời gian như "lịch sale" nếu nội dung cần làm tươi định kỳ; các trang còn lại SSG thuần.

---

## §7 - Phụ thuộc

- **FR-WEB-001** - cung cấp scaffold Next.js + route group; landing dùng group `(marketing)` công khai, nằm ngoài guard `(app)` (depends_on cứng).
- **FR-DEAL-001 (nội dung tham chiếu)** - năng lực phát hiện sale ảo là nội dung thật cho trang "sale thật hay sale ảo"; landing không gọi API nhưng mô tả đúng năng lực.
- **FR-WEB-003 (nội dung tham chiếu)** - biểu đồ giá là minh chứng cho thông điệp; CTA dẫn người dùng tới trải nghiệm này sau đăng ký.
- Lib: `next` 14 (Metadata API, MetadataRoute, generateStaticParams); test qua Jest/Vitest.

---

## §8 - Payload ví dụ

### HTML render của trang "sale thật hay sale ảo" (rút gọn)

```html
<article lang="vi-VN">
  <h1>Sale thật hay sale ảo</h1>
  <p>...nội dung định hướng tìm kiếm, gắn năng lực phát hiện sale ảo bằng lịch sử giá 90 ngày...</p>
  <a href="/login?signup=1">Dùng SănDeal miễn phí</a>
  <script type="application/ld+json">
    {"@context":"https://schema.org","@type":"FAQPage","mainEntity":[{"@type":"Question",
     "name":"Làm sao biết sale thật hay sale ảo?","acceptedAnswer":{"@type":"Answer",
     "text":"So giá hiện tại với trung vị 90 ngày và đáy giá lịch sử..."}}]}
  </script>
</article>
```

### sitemap.xml (do app/sitemap.ts sinh)

```xml
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://sandeal.vn/cach-san-xu-shopee</loc><lastmod>2026-06-28</lastmod></url>
  <url><loc>https://sandeal.vn/lich-sale</loc><lastmod>2026-06-28</lastmod></url>
  <url><loc>https://sandeal.vn/ma-freeship</loc><lastmod>2026-06-28</lastmod></url>
  <url><loc>https://sandeal.vn/sale-that-hay-sale-ao</loc><lastmod>2026-06-28</lastmod></url>
</urlset>
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Mở rộng tập keyword theo nghiên cứu search volume thực tế - thêm dần vào `keywords.ts`, sitemap tự cập nhật.
- Nội dung động "lịch sale sắp tới" lấy từ dữ liệu mốc ngày đôi (FR-DEAL-003 double_dates) - cân nhắc ISR khi có nguồn.
- Trang blog/nội dung dài cho long-tail keyword - giai đoạn SEO sâu hơn.
- Đa ngôn ngữ trang landing khi mở SEA (hreflang) - bám i18n nền của FR-WEB-001.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Render client-only | crawler thấy trang trống | không index, hỏng SEO | SSG/ISR, HTML có sẵn (DEC-WEB-06) |
| Thiếu title/description | seo-metadata test | snippet xấu, click thấp | generateMetadata bắt buộc (DEC-WEB-07) |
| Trùng title nhiều trang | test seen-set | loãng SEO | Mỗi keyword title riêng |
| Thiếu canonical | review meta | phạt trùng nội dung | alternates.canonical per trang |
| JSON-LD sai schema | seo-jsonld test | mất rich result | buildJsonLd đúng @type (DEC-WEB-08) |
| Sitemap lệch danh sách keyword | sitemap test | trang mới không index | Sinh từ keywords.ts (DEC-WEB-09) |
| Landing bắt đăng nhập | grep apiFetch/guard | đứt funnel bước đầu | group (marketing) công khai (DEC-WEB-10) |
| Nội dung spam/cloaking | review nội dung | phạt SEO + lệch niềm tin | Gắn năng lực thật (§1 #10) |
| Thiếu CTA | review trang | không chuyển bước funnel | Mỗi trang ít nhất một CTA (§1 #7) |

---

## §11 - Ghi chú

- SEO là kênh GTM Phase 1, chạy trước extension (§8) để kéo organic và song hành giải cold-start.
- CAC thấp qua SEO/viral là giả định nền của unit economics (§4.1) - không có SEO thì phải mua traffic, phá payback.
- SSG/ISR là khác biệt giữa SEO hoạt động và vô nghĩa: crawler đọc HTML đầu tiên, không chờ JS.
- Cụm keyword lõi (cách săn xu Shopee, sale thật hay sale ảo) là sân BeeCost đã xây thương hiệu (§5.6); thắng SEO ở đây là điều kiện có cửa.
- Nội dung trung thực gắn năng lực thật giữ cả thứ hạng bền lẫn thông điệp niềm tin hậu-Honey.
- Sitemap/robots sinh tự động từ một nguồn keyword giữ chúng luôn khớp tập trang thực tế khi mở rộng.

---

*Hết FR-WEB-002. Status: ready_to_implement (mục tiêu audit 10/10).*
