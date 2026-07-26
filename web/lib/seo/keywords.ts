export type KeywordIntent = "guide" | "faq" | "list";
export type KeywordCluster = "calendar" | "coupon" | "tactics" | "verdict" | "compare";

export type KeywordFAQ = { q: string; a: string };

export interface KeywordPage {
  slug: string;
  keyword: string;
  title: string;
  description: string;
  intent: KeywordIntent;
  cluster: KeywordCluster;
  /** Unique intro paragraphs — never shared boilerplate across pages. */
  intro: string[];
  faqs: KeywordFAQ[];
  /** Related keyword slugs and/or absolute app paths starting with /. */
  related: string[];
  /** Optional structured sale date for calendar pages (UTC). */
  saleDate?: string;
}

export const KEYWORD_PAGES: KeywordPage[] = [
  {
    slug: "cach-san-xu-shopee",
    keyword: "Cách săn xu Shopee",
    title: "Cách săn xu Shopee hiệu quả 2026 | Shopass",
    description: "Hướng dẫn săn xu Shopee đúng cách, checklist nhắc nhở, không tự động click rủi ro tài khoản.",
    intent: "guide",
    cluster: "tactics",
    intro: [
      "Xu Shopee giảm trực tiếp vào đơn — nhưng click ảo hoặc extension bấm hộ dễ làm tài khoản bị hạn chế.",
      "Cách an toàn: mở nhiệm vụ xu trong app, hoàn thành thao tác bạn chủ động, rồi dùng xu khi thanh toán. Shopass không tự click nhiệm vụ giúp bạn.",
    ],
    faqs: [
      {
        q: "Có nên dùng tool tự săn xu không?",
        a: "Không khuyến khích. Rủi ro khóa tài khoản cao hơn lợi ích vài nghìn xu.",
      },
      {
        q: "Xu có hết hạn không?",
        a: "Có — kiểm tra hạn dùng trong ví xu trước khi để dành quá lâu.",
      },
    ],
    related: ["ma-freeship", "ma-giam-gia-shopee", "san-deal-flash-sale", "/bang-gia"],
  },
  {
    slug: "ma-freeship",
    keyword: "Mã freeship",
    title: "Mã freeship Shopee TikTok Lazada hôm nay | Shopass",
    description: "Tổng hợp mã freeship và cách tối ưu freeship khi thanh toán đa sàn.",
    intent: "guide",
    cluster: "coupon",
    intro: [
      "Freeship thường gắn điều kiện giá trị đơn, ngành hàng hoặc khung giờ — không phải mọi mã đều áp mọi sản phẩm.",
      "Trước khi săn mã, hãy chắc giá sản phẩm đã hợp lý (không phải sale ảo). Freeship trên giá đã thổi vẫn là mua đắt.",
    ],
    faqs: [
      {
        q: "Mã freeship lấy ở đâu?",
        a: "Trong app sàn (ví voucher), livestream, hoặc trang thương hiệu. Shopass không phát mã giả — chúng tôi giúp kiểm giá trước.",
      },
    ],
    related: ["ma-giam-gia-shopee", "ma-freeship-tiktok", "sale-that-hay-sale-ao", "/kiem-tra-sale-ao"],
  },
  {
    slug: "sale-that-hay-sale-ao",
    keyword: "Sale thật hay sale ảo",
    title: "Sale thật hay sale ảo? Cách nhận biết | Shopass",
    description: "Phân biệt sale thật và sale ảo bằng lịch sử giá 90 ngày, median90 và đáy giá.",
    intent: "faq",
    cluster: "verdict",
    intro: [
      "Sale ảo = giá gốc bị đẩy lên rồi “giảm sâu” về mức gần trung vị lịch sử. Nhãn % giảm lớn nhưng bạn không tiết kiệm thật.",
      "Cách chắc: so giá hiện tại với trung vị 90 ngày và đáy gần đây — đúng việc công cụ Kiểm tra sale ảo của Shopass làm.",
    ],
    faqs: [
      {
        q: "Làm sao biết sale thật hay sale ảo?",
        a: "So giá hiện tại với trung vị 90 ngày và đáy giá lịch sử; giá gốc bị thổi + giảm không thật là sale ảo.",
      },
      {
        q: "Chưa có đủ lịch sử thì sao?",
        a: "Shopass không kết luận khi dữ liệu mỏng (<14 ngày). Đăng ký nhận verdict khi đủ điểm giá.",
      },
    ],
    related: ["gia-goc-la-gi", "/kiem-tra-sale-ao", "cach-theo-doi-gia-shopee", "/onboarding"],
  },
  // --- Batch 1 new pages (R42) ---
  {
    slug: "lich-sale-8-8",
    keyword: "Lịch sale 8.8",
    title: "Lịch sale 8.8 Shopee TikTok Lazada | Shopass",
    description: "Ngày đôi 8.8 sắp tới — đếm ngược, checklist giá, nhắc email/Zalo.",
    intent: "list",
    cluster: "calendar",
    saleDate: "2026-08-08",
    intro: [
      "8.8 là mốc sale giữa năm: voucher chồng lớp, flash sale khung giờ, và nhiều “giảm sốc” trên giá đã chỉnh.",
      "Trước ngày đôi: theo dõi 3–5 sản phẩm bạn thật sự muốn mua; ngày sale chỉ quyết định nếu giá chạm đáy hoặc thấp hơn median 90 ngày.",
    ],
    faqs: [
      {
        q: "8.8 có phải lúc rẻ nhất không?",
        a: "Không luôn. Nhiều SKU rẻ hơn ở payday hoặc sau 8.8 vài ngày. Cứ so với lịch sử giá.",
      },
    ],
    related: ["lich-sale-9-9", "lich-sale-11-11", "/lich-sale", "sale-that-hay-sale-ao"],
  },
  {
    slug: "lich-sale-9-9",
    keyword: "Lịch sale 9.9",
    title: "Lịch sale 9.9 — chuẩn bị săn deal | Shopass",
    description: "Mốc 9.9: checklist theo dõi giá, tránh sale ảo, nhắc trước ngày đôi.",
    intent: "list",
    cluster: "calendar",
    saleDate: "2026-09-09",
    intro: [
      "9.9 thường mở campaign lớn hơn 8.8 trên một số ngành (điện tử, làm đẹp). Cạnh tranh voucher cao = dễ bị đánh lạc hướng bởi % giảm.",
      "Hãy gắn cảnh báo trước 3–7 ngày: nếu giá đã ở đáy trước 9.9, không cần chờ “siêu sale” để mua.",
    ],
    faqs: [
      {
        q: "Nên mua trước hay trong 9.9?",
        a: "Mua khi giá tốt so với lịch sử — có thể trước hoặc trong ngày đôi. Đừng mua chỉ vì banner.",
      },
    ],
    related: ["lich-sale-8-8", "lich-sale-12-12", "/lich-sale", "/onboarding"],
  },
  {
    slug: "lich-sale-11-11",
    keyword: "Lịch sale 11.11",
    title: "Lịch sale 11.11 Singles Day | Shopass",
    description: "11.11 — đợt sale lớn cuối năm: chiến lược theo dõi giá và tránh thổi giá gốc.",
    intent: "list",
    cluster: "calendar",
    saleDate: "2026-11-11",
    intro: [
      "11.11 là ngày đôi lớn nhất nhiều sàn. Người bán thường tăng list price trước đó 2–4 tuần.",
      "Bắt đầu theo dõi từ giữa tháng 10. Shopass giữ median 90 ngày để bạn không bị “giảm 70%” đánh lừa.",
    ],
    faqs: [
      {
        q: "11.11 có rẻ hơn 12.12 không?",
        a: "Tùy ngành. Điện tử hay chạm đáy quanh 11.11; thời trang có thể chờ 12.12. Theo dõi SKU cụ thể.",
      },
    ],
    related: ["lich-sale-12-12", "lich-sale-9-9", "/lich-sale", "/kiem-tra-sale-ao"],
  },
  {
    slug: "lich-sale-12-12",
    keyword: "Lịch sale 12.12",
    title: "Lịch sale 12.12 cuối năm | Shopass",
    description: "12.12 và sóng clearance: khi nào nên chốt đơn, khi nào chờ thêm.",
    intent: "list",
    cluster: "calendar",
    saleDate: "2026-12-12",
    intro: [
      "12.12 đóng năm với voucher lớn + hàng tồn. Một số mặt hàng thật sự đáy năm; số khác chỉ “sale hình ảnh”.",
      "Dùng biểu đồ cả năm (khi đủ dữ liệu) thay vì tin banner “thấp nhất năm”.",
    ],
    faqs: [
      {
        q: "Sau 12.12 giá có giảm tiếp không?",
        a: "Hàng mùa vụ có thể giảm thêm; hàng hot thường tăng lại. Cảnh báo đáy hữu ích hơn đoán mò.",
      },
    ],
    related: ["lich-sale-11-11", "payday-sale-shopee", "/lich-sale", "/bang-gia"],
  },
  {
    slug: "payday-sale-shopee",
    keyword: "Payday sale Shopee",
    title: "Payday sale Shopee là gì? Lịch & mẹo | Shopass",
    description: "Payday sale quanh ngày lương: khung giờ, voucher, và cách không mua đắt.",
    intent: "list",
    cluster: "calendar",
    intro: [
      "Payday sale Shopee thường rơi vào cuối tháng / đầu tháng (theo chiến dịch từng kỳ), tập trung voucher và xu.",
      "Đây là lúc tốt để dùng hết voucher — nhưng vẫn kiểm tra giá so với median trước khi thêm vào giỏ.",
    ],
    faqs: [
      {
        q: "Payday có cố định ngày không?",
        a: "Không tuyệt đối — theo lịch campaign Shopee từng tháng. Theo dõi thông báo app và trang lịch sale Shopass.",
      },
    ],
    related: ["lich-sale-8-8", "ma-giam-gia-shopee", "/lich-sale", "cach-san-xu-shopee"],
  },
  {
    slug: "ma-giam-gia-shopee",
    keyword: "Mã giảm giá Shopee",
    title: "Mã giảm giá Shopee — dùng sao cho không mua đắt | Shopass",
    description: "Phân loại voucher Shopee, điều kiện ẩn, và kiểm giá trước khi áp mã.",
    intent: "guide",
    cluster: "coupon",
    intro: [
      "Voucher shop, voucher sàn, và mã vận chuyển chồng lên nhau — nhưng điều kiện tối thiểu đơn hay ngành hàng hay làm mã “đẹp” thành vô dụng.",
      "Quy tắc Shopass: xác nhận giá nền đã hợp lý, rồi mới tối ưu mã. Mã 50k trên giá đã thổi 100k vẫn là lỗ.",
    ],
    faqs: [
      {
        q: "Nên lưu mã trước hay xem giá trước?",
        a: "Xem lịch sử giá trước. Mã chỉ là lớp cuối cùng.",
      },
    ],
    related: ["ma-freeship", "cach-san-xu-shopee", "sale-that-hay-sale-ao", "/kiem-tra-sale-ao"],
  },
  {
    slug: "ma-freeship-tiktok",
    keyword: "Mã freeship TikTok",
    title: "Mã freeship TikTok Shop hôm nay | Shopass",
    description: "Freeship TikTok Shop: chỗ lấy mã, điều kiện, và theo dõi giá trước livestream.",
    intent: "guide",
    cluster: "coupon",
    intro: [
      "TikTok Shop gắn mã freeship với livestream và xu hướng ngắn — mã hết nhanh và điều kiện đổi từng giờ.",
      "Trước khi chốt trong live: mở link sản phẩm trên Shopass (khi đã track) để xem có đang cao hơn median không.",
    ],
    faqs: [
      {
        q: "Live có luôn rẻ hơn không?",
        a: "Không. Nhiều live tăng “giá gạch” rồi giảm về mức bình thường. Cần lịch sử giá.",
      },
    ],
    related: ["ma-freeship", "san-deal-flash-sale", "/kiem-tra-sale-ao", "/onboarding"],
  },
  {
    slug: "san-deal-flash-sale",
    keyword: "Săn deal flash sale",
    title: "Cách săn deal flash sale không bị hớ | Shopass",
    description: "Flash sale khung giờ: chuẩn bị giỏ, kiểm giá, cảnh báo — không F5 vô ích.",
    intent: "guide",
    cluster: "tactics",
    intro: [
      "Flash sale thắng nhờ chuẩn bị: sản phẩm đã trong wishlist, voucher sẵn, và bạn biết mức giá “đáng mua” trước đó.",
      "Shopass giúp bạn đặt mức đó bằng lịch sử + cảnh báo — thay vì đoán trong 3 giây đếm ngược.",
    ],
    faqs: [
      {
        q: "Có nên dùng bot flash sale không?",
        a: "Không. Vi phạm điều khoản sàn và rủi ro khóa tài khoản.",
      },
    ],
    related: ["cach-theo-doi-gia-shopee", "payday-sale-shopee", "/onboarding", "/bang-gia"],
  },
  {
    slug: "gia-goc-la-gi",
    keyword: "Giá gốc là gì",
    title: "Giá gốc là gì? Vì sao % giảm hay ảo | Shopass",
    description: "Giải thích giá gốc / giá gạch trên sàn và cách không bị đánh lừa bởi phần trăm giảm.",
    intent: "faq",
    cluster: "verdict",
    intro: [
      "“Giá gốc” trên thẻ sản phẩm thường do người bán tự đặt — không phải giá thị trường đã kiểm chứng.",
      "Phần trăm giảm chỉ có nghĩa khi bạn so với giá giao dịch gần đây (median / đáy), không phải với số gạch tự điền.",
    ],
    faqs: [
      {
        q: "Giá gốc có bắt buộc đúng không?",
        a: "Không. Đó là nhãn marketing. Hãy tin lịch sử giá giao dịch hơn nhãn %.",
      },
      {
        q: "Làm sao kiểm tra nhanh?",
        a: "Dùng Kiểm tra sale ảo của Shopass hoặc xem biểu đồ 90 ngày sau khi theo dõi sản phẩm.",
      },
    ],
    related: ["sale-that-hay-sale-ao", "/kiem-tra-sale-ao", "cach-theo-doi-gia-shopee"],
  },
  {
    slug: "cach-theo-doi-gia-shopee",
    keyword: "Cách theo dõi giá Shopee",
    title: "Cách theo dõi giá Shopee miễn phí | Shopass",
    description: "Dán link Shopee, xem lịch sử, bật cảnh báo đáy — không cần F5 mỗi ngày.",
    intent: "guide",
    cluster: "tactics",
    intro: [
      "Theo dõi giá đúng nghĩa là có chuỗi điểm giá theo thời gian + quy tắc cảnh báo — không phải nhớ “hôm qua thấy rẻ”.",
      "Trên Shopass: dán link → xem biểu đồ → một nút “Báo tôi khi chạm đáy”. Closed beta đang tối ưu Shopee VN trước.",
    ],
    faqs: [
      {
        q: "Có cần cài extension không?",
        a: "Không bắt buộc để theo dõi. Extension hữu ích khi bạn muốn gửi giá đang thấy trên trang Shopee vào biểu đồ.",
      },
    ],
    related: ["/onboarding", "/kiem-tra-sale-ao", "sale-that-hay-sale-ao", "san-deal-flash-sale"],
  },
];

export function getKeywordPage(slug: string): KeywordPage | undefined {
  return KEYWORD_PAGES.find((p) => p.slug === slug);
}

const PATH_LABELS: Record<string, string> = {
  "/lich-sale": "Công cụ lịch sale",
  "/kiem-tra-sale-ao": "Kiểm tra sale ảo",
  "/onboarding": "Bắt đầu theo dõi",
  "/bang-gia": "Bảng giá",
  "/blog": "Blog",
};

export function relatedPages(p: KeywordPage): { href: string; label: string }[] {
  return p.related.map((ref) => {
    if (ref.startsWith("/")) {
      return { href: ref, label: PATH_LABELS[ref] ?? ref };
    }
    const hit = getKeywordPage(ref);
    return { href: `/${ref}`, label: hit?.keyword ?? ref };
  });
}
