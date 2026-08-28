import Link from "next/link";

type ChartEmptyStateProps = {
  onFocusCapture?: () => void;
};

/**
 * Honest empty chart when `price_daily` has no points (even if raw snapshots exist).
 * Does not invent fake series — guides the user to confirm a browser price.
 */
export function ChartEmptyState({ onFocusCapture }: ChartEmptyStateProps) {
  return (
    <div className="absolute inset-0 flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-300/80 bg-gradient-to-b from-slate-50 to-white px-6 py-10 text-center">
      <div
        className="flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-slate-400 shadow-sm ring-1 ring-slate-200/80"
        aria-hidden="true"
      >
        <svg className="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.75}
            d="M7 12l3-3 3 3 4-4M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
          />
        </svg>
      </div>
      <p className="mt-4 text-base font-black tracking-tight text-slate-900">
        Chưa đủ dữ liệu để vẽ biểu đồ
      </p>
      <p className="mt-2 max-w-md text-sm leading-relaxed text-slate-500">
        Sản phẩm đã được theo dõi, nhưng chưa có điểm giá ngày nào trong khoảng bạn chọn. Shopass
        không hiển thị biểu đồ giả — hãy ghi nhận giá bạn đang thấy trên Shopee ở mục phía trên.
      </p>
      <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
        <button
          type="button"
          onClick={onFocusCapture}
          className="inline-flex cursor-pointer items-center justify-center rounded-xl bg-slate-950 px-5 py-2.5 text-sm font-extrabold text-white shadow-sm transition hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/25"
        >
          Ghi nhận giá đang thấy
        </button>
        <Link
          href="/capture-guide"
          className="inline-flex cursor-pointer items-center justify-center rounded-xl border border-slate-200 bg-white px-5 py-2.5 text-sm font-bold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/20"
        >
          Hướng dẫn lấy giá một chạm
        </Link>
      </div>
    </div>
  );
}
