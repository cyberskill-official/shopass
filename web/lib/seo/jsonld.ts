import type { KeywordPage } from "./keywords";

export function buildJsonLd(p: KeywordPage, url: string): object {
  if (p.intent === "faq") {
    return {
      "@context": "https://schema.org",
      "@type": "FAQPage",
      mainEntity: [{
        "@type": "Question",
        name: "Làm sao biết sale thật hay sale ảo?",
        acceptedAnswer: {
          "@type": "Answer",
          text: "So giá hiện tại với trung vị 90 ngày và đáy giá lịch sử; giá gốc bị thổi + giảm không thật là sale ảo.",
        },
      }],
    };
  }
  if (p.intent === "list") {
    return {
      "@context": "https://schema.org",
      "@type": "ItemList",
      name: p.keyword,
      url,
    };
  }
  return {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: p.keyword,
    url,
    inLanguage: "vi-VN",
  };
}
