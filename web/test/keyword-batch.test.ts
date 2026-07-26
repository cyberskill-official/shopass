import { KEYWORD_PAGES } from "../lib/seo/keywords";

describe("R42 keyword batch-1", () => {
  it("ships at least 10 new pages beyond the original 3 cores", () => {
    expect(KEYWORD_PAGES.length).toBeGreaterThanOrEqual(13);
    const batch1 = [
      "lich-sale-8-8",
      "lich-sale-9-9",
      "lich-sale-11-11",
      "lich-sale-12-12",
      "payday-sale-shopee",
      "ma-giam-gia-shopee",
      "ma-freeship-tiktok",
      "san-deal-flash-sale",
      "gia-goc-la-gi",
      "cach-theo-doi-gia-shopee",
    ];
    for (const slug of batch1) {
      const p = KEYWORD_PAGES.find((k) => k.slug === slug);
      expect(p).toBeTruthy();
      expect(p!.intro.length).toBeGreaterThanOrEqual(2);
      expect(p!.faqs.length).toBeGreaterThanOrEqual(1);
      expect(p!.related.length).toBeGreaterThanOrEqual(2);
    }
  });

  it("every page has unique intro first line", () => {
    const first = KEYWORD_PAGES.map((p) => p.intro[0]);
    expect(new Set(first).size).toBe(first.length);
  });
});
