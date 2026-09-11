import React from "react";
import type { Maturity } from "@/lib/chart/types";

type MaturityNoticeProps = {
  maturity: Maturity;
  /** Honest thin-history hint when a few daily points exist but maturity is not MATURE. */
  pointCount?: number;
};

export function MaturityNotice({ maturity, pointCount }: MaturityNoticeProps) {
  if (maturity === "MATURE") return null;

  const pointsLabel =
    typeof pointCount === "number" && pointCount > 0
      ? ` Hiện có ${pointCount} điểm giá ngày trong khoảng đã chọn — chưa đủ để kết luận ổn định.`
      : "";

  if (maturity === "NEW") {
    return (
      <div
        role="status"
        className="mb-4 rounded-2xl border border-amber-200/80 bg-amber-50/80 px-4 py-3.5 text-sm leading-relaxed text-amber-950"
      >
        <p className="font-bold text-amber-900">Chưa đủ dữ liệu · lịch sử mỏng</p>
        <p className="mt-1 text-amber-800/90">
          Shopass mới bắt đầu thu thập lịch sử giá cho sản phẩm này. Không có biểu đồ giả và không
          suy diễn “sale thật / sale ảo” từ chuỗi trống hoặc quá ngắn.
          {pointsLabel} Hãy ghi nhận giá bạn đang thấy trên Shopee ở mục phía trên.
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
        Lịch sử còn mỏng — kết luận sẽ chính xác hơn khi có thêm điểm giá thật.
        {pointsLabel} Tiếp tục ghi nhận giá bạn thấy trên Shopee; Shopass không đệm dữ liệu giả.
      </p>
    </div>
  );
}
