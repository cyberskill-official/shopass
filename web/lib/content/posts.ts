export type BlogPost = {
  slug: string;
  title: string;
  description: string;
  date: string; // YYYY-MM-DD
  tags: string[];
  body: string[];
};

export const BLOG_POSTS: BlogPost[] = [
  {
    slug: "sale-that-hay-sale-ao-10-giay",
    title: "Sale thật hay sale ảo — kiểm tra trong 10 giây",
    description:
      "Dán link sản phẩm, so giá hiện tại với median 90 ngày. Không cần đoán mò.",
    date: "2026-07-26",
    tags: ["sale-ao", "huong-dan"],
    body: [
      "Giảm 50% nghe hấp dẫn — cho đến khi bạn thấy “giá gốc” bị thổi trước đó. Shopass gọi đó là sale ảo.",
      "Cách nhanh: mở công cụ Kiểm tra sale ảo, dán link Shopee/Lazada/TikTok. Nếu sản phẩm đã được theo dõi, bạn thấy verdict so với trung vị 90 ngày ngay lập tức.",
      "Chưa có trong hệ thống? Để lại email — chúng tôi báo khi đủ lịch sử để kết luận trung thực (không bịa số khi dữ liệu mỏng).",
      "Muốn cảnh báo tự động? Tạo tài khoản và bật “Báo tôi khi chạm đáy” trong onboarding — mất dưới hai phút.",
    ],
  },
  {
    slug: "chao-shopass",
    title: "Chào Shopass — săn deal bằng dữ liệu, không bằng cảm tính",
    description:
      "Ra mắt closed beta: lịch sử giá đa sàn, phát hiện sale ảo, cảnh báo khi gần đáy. Affiliate chỉ khi bạn bấm.",
    date: "2026-07-26",
    tags: ["launch", "trust"],
    body: [
      "Shopass (CyberSkill) là công cụ theo dõi giá cho người mua Việt Nam trên Shopee, TikTok Shop và Lazada.",
      "Chúng tôi không làm coupon overlay kiểu Honey. Sau vụ Honey bị cắt affiliate vì ghi đè attribution, Shopass chọn kiến trúc đối lập: bạn thấy dữ liệu nào rời máy, và affiliate chỉ gắn khi bạn chủ động bấm. Chi tiết trên trang Minh bạch.",
      "Closed beta đang mở: theo dõi link Shopee, xem biểu đồ, bật cảnh báo. Premium (dự đoán đáy) đang thu danh sách chờ trên bảng giá.",
      "Nếu bạn săn deal hàng ngày — bắt đầu từ trang chủ hoặc dán một link vào Kiểm tra sale ảo.",
    ],
  },
];

export function getPost(slug: string): BlogPost | undefined {
  return BLOG_POSTS.find((p) => p.slug === slug);
}

export function postsByDateDesc(): BlogPost[] {
  return [...BLOG_POSTS].sort((a, b) => (a.date < b.date ? 1 : -1));
}
