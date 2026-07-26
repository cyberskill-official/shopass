import { buildJsonLd } from "../lib/seo/jsonld";
import { KEYWORD_PAGES, type KeywordPage } from "../lib/seo/keywords";

describe("SEO JSON-LD", () => {
  it("FAQ pages emit FAQPage; guides emit Article; list intent emits ItemList", () => {
    const faq = KEYWORD_PAGES.find((p) => p.intent === "faq")!;
    const guide = KEYWORD_PAGES.find((p) => p.intent === "guide")!;
    const list: KeywordPage = {
      slug: "list-fixture",
      keyword: "Fixture list",
      title: "Fixture",
      description: "Fixture",
      intent: "list",
    };
    expect((buildJsonLd(faq, "https://shopass.cyberskill.world/x") as { "@type": string })["@type"]).toBe(
      "FAQPage",
    );
    expect((buildJsonLd(guide, "https://shopass.cyberskill.world/y") as { "@type": string })["@type"]).toBe(
      "Article",
    );
    expect((buildJsonLd(list, "https://shopass.cyberskill.world/z") as { "@type": string })["@type"]).toBe(
      "ItemList",
    );
  });

  it("JSON-LD có @context schema.org", () => {
    for (const p of KEYWORD_PAGES) {
      const ld = buildJsonLd(p, "https://shopass.cyberskill.world/z") as { "@context": string };
      expect(ld["@context"]).toBe("https://schema.org");
    }
  });
});
