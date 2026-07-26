import type { KeywordPage } from "./keywords";

export function buildJsonLd(p: KeywordPage, url: string): object {
  if (p.intent === "faq" || (p.faqs && p.faqs.length > 0 && p.intent !== "list")) {
    return {
      "@context": "https://schema.org",
      "@type": "FAQPage",
      mainEntity: p.faqs.map((f) => ({
        "@type": "Question",
        name: f.q,
        acceptedAnswer: {
          "@type": "Answer",
          text: f.a,
        },
      })),
    };
  }
  if (p.intent === "list") {
    const entity: Record<string, unknown> = {
      "@context": "https://schema.org",
      "@type": "ItemList",
      name: p.keyword,
      url,
    };
    if (p.saleDate) {
      entity.itemListElement = [
        {
          "@type": "ListItem",
          position: 1,
          name: p.keyword,
          url,
          additionalProperty: {
            "@type": "PropertyValue",
            name: "saleDate",
            value: p.saleDate,
          },
        },
      ];
    }
    // FAQ still useful on calendar pages
    if (p.faqs.length > 0) {
      return [entity, {
        "@context": "https://schema.org",
        "@type": "FAQPage",
        mainEntity: p.faqs.map((f) => ({
          "@type": "Question",
          name: f.q,
          acceptedAnswer: { "@type": "Answer", text: f.a },
        })),
      }];
    }
    return entity;
  }
  return {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: p.keyword,
    description: p.description,
    url,
    inLanguage: "vi-VN",
  };
}
