import { buildJsonLd } from "../lib/seo/jsonld";
import { KEYWORD_PAGES, type KeywordPage } from "../lib/seo/keywords";

describe("SEO JSON-LD", () => {
  it("FAQ pages emit FAQPage; guides with FAQs emit FAQPage; list can emit ItemList", () => {
    const faq = KEYWORD_PAGES.find((p) => p.intent === "faq")!;
    const guide = KEYWORD_PAGES.find((p) => p.intent === "guide")!;
    const list = KEYWORD_PAGES.find((p) => p.intent === "list")!;
    const faqLd = buildJsonLd(faq, "https://shopass.cyberskill.world/x") as { "@type": string };
    expect(faqLd["@type"]).toBe("FAQPage");
    const guideLd = buildJsonLd(guide, "https://shopass.cyberskill.world/y") as { "@type": string };
    expect(guideLd["@type"]).toBe("FAQPage");
    const listLd = buildJsonLd(list, "https://shopass.cyberskill.world/z");
    if (Array.isArray(listLd)) {
      expect(listLd.some((b) => (b as { "@type": string })["@type"] === "ItemList")).toBe(true);
    } else {
      expect((listLd as { "@type": string })["@type"]).toBe("ItemList");
    }
  });

  it("JSON-LD có @context schema.org", () => {
    for (const p of KEYWORD_PAGES) {
      const ld = buildJsonLd(p, "https://shopass.cyberskill.world/z");
      const blocks = Array.isArray(ld) ? ld : [ld];
      for (const b of blocks) {
        expect((b as { "@context": string })["@context"]).toBe("https://schema.org");
      }
    }
  });

  it("list fixture without faqs still builds ItemList", () => {
    const bare: KeywordPage = {
      slug: "list-fixture",
      keyword: "Fixture list",
      title: "Fixture",
      description: "Fixture",
      intent: "list",
      cluster: "calendar",
      intro: ["x"],
      faqs: [],
      related: [],
    };
    expect((buildJsonLd(bare, "https://shopass.cyberskill.world/z") as { "@type": string })["@type"]).toBe(
      "ItemList",
    );
  });
});
