export interface CollectedField {
  field: string;
  purpose: string;
  legalBasis: string;
}

export const COLLECTED_FIELDS: CollectedField[] = [
  { field: "platform", purpose: "Biết bạn đang xem sàn nào để tra đúng dữ liệu giá", legalBasis: "Đồng thuận - mục đích theo dõi giá" },
  { field: "productId", purpose: "Tra cứu sản phẩm để hiện lịch sử giá và sale ảo", legalBasis: "Đồng thuận - mục đích theo dõi giá" },
  { field: "price", purpose: "Theo dõi biến động giá để cảnh báo và vẽ biểu đồ", legalBasis: "Đồng thuận - mục đích theo dõi giá" },
  { field: "qty", purpose: "Tính tối ưu voucher/giỏ hàng cho bạn", legalBasis: "Đồng thuận - mục đích tối ưu giỏ" },
];

export const NEVER_COLLECTED = [
  "cookie", "mật khẩu", "token phiên sàn", "header xác thực",
  "email", "số điện thoại", "tên", "địa chỉ", "định danh người dùng sàn",
] as const;
