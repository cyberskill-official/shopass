import { generateMetadata, generateStaticParams } from "../app/(marketing)/[keyword]/page";
import { KEYWORD_PAGES } from "../lib/seo/keywords";

describe("SEO Metadata", () => {
  it("SSG sinh param cho mọi keyword lõi", () => {
    const params = generateStaticParams();
    expect(params.length).toBe(KEYWORD_PAGES.length);
    expect(params.map((p) => p.keyword)).toContain("sale-that-hay-sale-ao");
  });

  it("mỗi keyword có title/description/canonical riêng", async () => {
    const seen = new Set<string>();
    for (const p of KEYWORD_PAGES) {
      const meta = await generateMetadata({ params: Promise.resolve({ keyword: p.slug }) });
      expect(meta.title).toBeTruthy();
      expect(meta.description).toBeTruthy();
      expect((meta.alternates as any).canonical).toContain(p.slug);
      expect(seen.has(meta.title as string)).toBe(false); // không trùng title
      seen.add(meta.title as string);
    }
  });
});
