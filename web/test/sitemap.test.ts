import sitemap from "../app/sitemap";
import { KEYWORD_PAGES } from "../lib/seo/keywords";

describe("Sitemap", () => {
  it("sitemap liệt kê đủ mọi keyword", () => {
    const urls = sitemap().map((e) => e.url);
    for (const p of KEYWORD_PAGES) {
      expect(urls).toContain(`https://sandeal.vn/${p.slug}`);
    }
  });
});
