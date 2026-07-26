import { siteURL } from "@/lib/site-url";

const FAQ = [
  {
    q: "Sale thật hay ảo là gì?",
    a: "Sale ảo là khi giá gốc bị thổi rồi 'giảm' xuống mức vẫn cao hơn trung vị hoặc đáy gần đây. Shopass so giá hiện tại với lịch sử 90 ngày để phân biệt.",
  },
  {
    q: "Shopass hỗ trợ sàn nào?",
    a: "Shopee, TikTok Shop và Lazada (Việt Nam trước). Theo dõi bằng link sản phẩm hoặc extension trên trang đang xem.",
  },
  {
    q: "Extension đọc gì trên máy tôi?",
    a: "Chỉ giá/ID công khai và (khi bật đồng bộ) voucher hiển thị. Không cookie phiên, mật khẩu hay token sàn — xem trang Minh bạch.",
  },
  {
    q: "Có cần trả phí không?",
    a: "Bắt đầu miễn phí. Premium sẽ mở thêm cảnh báo nâng cao; xem bảng giá khi có.",
  },
  {
    q: "Affiliate có bị gắn ngầm không?",
    a: "Không. Tham số affiliate chỉ gắn khi bạn chủ động bấm liên kết — không cookie-stuffing.",
  },
] as const;

export function landingJsonLd(): object[] {
  return [
    {
      "@context": "https://schema.org",
      "@type": "Organization",
      name: "Shopass",
      url: siteURL,
      parentOrganization: {
        "@type": "Organization",
        name: "CyberSkill Software Solutions Consultancy and Development JSC",
      },
    },
    {
      "@context": "https://schema.org",
      "@type": "WebSite",
      name: "Shopass",
      url: siteURL,
      inLanguage: "vi-VN",
      description: "Biết khi nào giá chạm đáy trên Shopee, TikTok Shop, Lazada.",
    },
    {
      "@context": "https://schema.org",
      "@type": "FAQPage",
      mainEntity: FAQ.map((item) => ({
        "@type": "Question",
        name: item.q,
        acceptedAnswer: { "@type": "Answer", text: item.a },
      })),
    },
  ];
}

export { FAQ as LANDING_FAQ };
