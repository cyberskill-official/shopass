import { buildJsonLd } from "../lib/seo/jsonld";
import { KEYWORD_PAGES } from "../lib/seo/keywords";

describe("SEO JSON-LD", () => {
  it("trang FAQ sinh FAQPage; lịch sale sinh ItemList", () => {
    const faq = KEYWORD_PAGES.find((p) => p.intent === "faq")!;
    const list = KEYWORD_PAGES.find((p) => p.intent === "list")!;
    expect((buildJsonLd(faq, "https://sandeal.vn/x") as any)["@type"]).toBe("FAQPage");
    expect((buildJsonLd(list, "https://sandeal.vn/y") as any)["@type"]).toBe("ItemList");
  });

  it("JSON-LD có @context schema.org", () => {
    for (const p of KEYWORD_PAGES) {
      const ld = buildJsonLd(p, "https://sandeal.vn/z") as any;
      expect(ld["@context"]).toBe("https://schema.org");
    }
  });
});
