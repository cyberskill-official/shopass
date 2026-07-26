export type CompareCell = "yes" | "no" | "partial" | "n/a";

export type CompareRow = {
  feature: string;
  shopass: CompareCell;
  beecost: CompareCell;
  note?: string;
};

/** Factual matrix as of 2026-07-26 publish. HITL must verify before indexing. */
export const SHOPASS_VS_BEECOST: CompareRow[] = [
  {
    feature: "Theo dõi giá Shopee VN",
    shopass: "yes",
    beecost: "yes",
  },
  {
    feature: "TikTok Shop",
    shopass: "partial",
    beecost: "no",
    note: "Shopass có parser/track path; live scrape mở dần theo R24.",
  },
  {
    feature: "Lazada",
    shopass: "partial",
    beecost: "yes",
    note: "Shopass parser sẵn; độ phủ scrape theo lộ trình.",
  },
  {
    feature: "Lịch sử giá / biểu đồ",
    shopass: "yes",
    beecost: "yes",
  },
  {
    feature: "Phát hiện sale ảo (median 90d)",
    shopass: "yes",
    beecost: "partial",
    note: "Shopass có verdict SALE_AO/SALE_XIN; BeeCost thiên về lịch sử.",
  },
  {
    feature: "Dự đoán đáy (p_bottom)",
    shopass: "yes",
    beecost: "no",
    note: "Shopass Premium; BeeCost không công bố forecast tương đương.",
  },
  {
    feature: "Tối ưu giỏ / voucher stack",
    shopass: "partial",
    beecost: "no",
    note: "Cart module trong repo; bề mặt người dùng đang mở dần.",
  },
  {
    feature: "Extension mã nguồn mở (MIT)",
    shopass: "partial",
    beecost: "no",
    note: "Extension MIT trong monorepo; public mirror chờ R36.",
  },
  {
    feature: "Affiliate chỉ khi user bấm",
    shopass: "yes",
    beecost: "n/a",
    note: "Shopass ghi rõ anti cookie-stuffing trên /minh-bach.",
  },
  {
    feature: "Chính sách bảo mật / điều khoản VN",
    shopass: "yes",
    beecost: "yes",
  },
];

export const CELL_LABEL: Record<CompareCell, string> = {
  yes: "Có",
  no: "Không",
  partial: "Một phần",
  "n/a": "Không rõ / N/A",
};
