import React from "react";
import type { Maturity } from "@/lib/chart/types";

export function MaturityNotice({ maturity }: { maturity: Maturity }) {
  if (maturity === "MATURE") return null;

  if (maturity === "NEW") {
    return (
      <div
        role="status"
        className="mb-4 rounded-2xl border border-amber-200/80 bg-amber-50/80 px-4 py-3.5 text-sm leading-relaxed text-amber-950"
      >
        <p className="font-bold text-amber-900">Chưa đủ dữ liệu</p>
        <p className="mt-1 text-amber-800/90">
          Shopass mới bắt đầu thu thập lịch sử giá cho sản phẩm này. Kết luận “sale thật / sale ảo”
          sẽ xuất hiện khi có đủ điểm giá — không suy diễn từ biểu đồ trống.
        </p>
      </div>
    );
  }

  // WARMING
  return (
    <div
      role="status"
      className="mb-4 rounded-2xl border border-sky-200/80 bg-sky-50/70 px-4 py-3.5 text-sm leading-relaxed text-sky-950"
    >
      <p className="font-bold text-sky-900">Đang tích lũy dữ liệu</p>
      <p className="mt-1 text-sky-800/90">
        Lịch sử còn mỏng. Hãy tiếp tục ghi nhận giá bạn thấy trên Shopee để biểu đồ và cảnh báo
        chính xác hơn.
      </p>
    </div>
  );
}
