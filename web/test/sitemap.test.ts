import sitemap from "../app/sitemap";
import { KEYWORD_PAGES } from "../lib/seo/keywords";
import { siteURL } from "../lib/site-url";

describe("Sitemap", () => {
  it("sitemap liệt kê đủ mọi keyword", () => {
    const urls = sitemap().map((e) => e.url);
    for (const p of KEYWORD_PAGES) {
      expect(urls).toContain(`${siteURL}/${p.slug}`);
    }
  });
});
