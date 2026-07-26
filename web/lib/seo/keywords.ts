export type KeywordIntent = "guide" | "faq" | "list";

export interface KeywordPage {
  slug: string;
  keyword: string;
  title: string;
  description: string;
  intent: KeywordIntent;
}

export const KEYWORD_PAGES: KeywordPage[] = [
  { slug: "cach-san-xu-shopee", keyword: "Cách săn xu Shopee",
    title: "Cách săn xu Shopee hiệu quả 2026 | Shopass",
    description: "Hướng dẫn săn xu Shopee đúng cách, checklist nhắc nhở, không tự động click rủi ro tài khoản.",
    intent: "guide" },
  { slug: "ma-freeship", keyword: "Mã freeship",
    title: "Mã freeship Shopee TikTok Lazada hôm nay | Shopass",
    description: "Tổng hợp mã freeship và cách tối ưu freeship khi thanh toán đa sàn.",
    intent: "guide" },
  { slug: "sale-that-hay-sale-ao", keyword: "Sale thật hay sale ảo",
    title: "Sale thật hay sale ảo? Cách nhận biết | Shopass",
    description: "Phân biệt sale thật và sale ảo bằng lịch sử giá 90 ngày, median90 và đáy giá.",
    intent: "faq" },
];
