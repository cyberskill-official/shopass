"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { ReferralCard } from "@/components/referral/referral-card";
import {
  listTrackedProducts,
  trackShopeeProduct,
  type TrackedProduct,
} from "@/lib/track/api";

const dateFormatter = new Intl.DateTimeFormat("vi-VN", {
  dateStyle: "medium",
  timeStyle: "short",
});

function formatPlatform(platform: string): string {
  return platform.toLowerCase() === "shopee" ? "Shopee VN" : platform;
}

function userMessage(error: unknown, fallback: string): string {
  if (!(error instanceof Error)) return fallback;
  if (error.message === "invalid item_url") {
    return "Liên kết không hợp lệ. Vui lòng nhập link trực tiếp của sản phẩm trên Shopee.";
  }
  return error.message || fallback;
}

export default function DashboardPage() {
  const router = useRouter();
  const [itemURL, setItemURL] = useState("");
  const [products, setProducts] = useState<TrackedProduct[]>([]);
  const [isLoadingProducts, setIsLoadingProducts] = useState(true);
  const [listError, setListError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const loadTrackedProducts = useCallback(async () => {
    setIsLoadingProducts(true);
    setListError(null);
    try {
      setProducts(await listTrackedProducts());
    } catch (error) {
      setListError(userMessage(error, "Không thể tải danh sách sản phẩm. Vui lòng thử lại sau."));
    } finally {
      setIsLoadingProducts(false);
    }
  }, []);

  useEffect(() => {
    void loadTrackedProducts();
  }, [loadTrackedProducts]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const url = itemURL.trim();
    if (!url) return;

    setIsSubmitting(true);
    setSubmitError(null);
    try {
      const tracked = await trackShopeeProduct(url);
      setItemURL("");
      void loadTrackedProducts();
      router.push(`/products/${tracked.product_id}/chart`);
    } catch (error) {
      setSubmitError(userMessage(error, "Không thể theo dõi sản phẩm này. Hãy kiểm tra lại đường link."));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="space-y-8">
      <ReferralCard />

      {/* Tracker Form Section */}
      <section className="relative overflow-hidden rounded-[2rem] bg-slate-950 px-6 py-10 text-white shadow-2xl shadow-sky-100/40 sm:px-10 sm:py-12 lg:px-12">
        <div className="absolute -left-20 -top-20 h-64 w-64 rounded-full bg-sky-500/20 blur-3xl" aria-hidden="true" />
        <div className="absolute bottom-0 right-10 h-40 w-40 rounded-full bg-teal-500/20 blur-3xl" aria-hidden="true" />

        <div className="relative z-10 mx-auto max-w-4xl">
          <p className="mb-4 inline-flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-slate-400">
            <span className="h-1.5 w-1.5 rounded-full bg-sky-400" aria-hidden="true" />
            Closed beta
          </p>

          <h1 className="text-3xl font-black tracking-tight sm:text-4xl lg:text-5xl">
            Theo dõi giá với dữ liệu thật.<br className="hidden sm:block" />
            <span className="text-sky-400">Chỉ từ một liên kết.</span>
          </h1>

          <p className="mt-4 max-w-2xl text-sm leading-relaxed text-slate-400 sm:text-base">
            Dán đường link sản phẩm Shopee vào đây để tạo sản phẩm theo dõi. Trong closed beta, bạn có thể xác nhận giá đang thấy trên Shopee ngay tại biểu đồ để bắt đầu lịch sử giá.
          </p>

          <Link
            href="/capture-guide"
            className="mt-5 inline-flex cursor-pointer items-center gap-2 text-sm font-bold text-sky-300 transition hover:text-white focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-400/30"
          >
            Dùng nút lấy giá Shopee một chạm <span aria-hidden="true">→</span>
          </Link>

          <form className="mt-8 flex flex-col gap-3 sm:flex-row" onSubmit={handleSubmit}>
            <div className="relative flex-1">
              <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-slate-400" aria-hidden="true">
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" /></svg>
              </div>
              <input
                aria-label="Liên kết sản phẩm Shopee Việt Nam"
                type="url"
                required
                value={itemURL}
                onChange={(event) => setItemURL(event.target.value)}
                placeholder="https://shopee.vn/..."
                className="w-full rounded-2xl border border-white/10 bg-white/5 py-4 pl-12 pr-4 text-sm font-medium text-white outline-none ring-sky-500/40 transition focus:bg-white/10 focus:ring-4 placeholder:text-slate-500"
              />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="flex cursor-pointer items-center justify-center gap-2 rounded-2xl bg-sky-600 px-8 py-4 text-sm font-extrabold text-white shadow-xl shadow-sky-900/40 transition hover:bg-sky-500 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-400/40 disabled:cursor-not-allowed disabled:opacity-70"
            >
              {isSubmitting ? (
                <>
                  <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none" aria-hidden="true"><circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" className="opacity-25"></circle><path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" className="opacity-75"></path></svg>
                  Đang thêm...
                </>
              ) : "Theo dõi giá"}
            </button>
          </form>

          {submitError && (
            <div className="mt-4 flex items-center gap-2 rounded-xl bg-red-500/10 p-3 text-sm font-medium text-red-200">
              <svg className="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
              {submitError}
            </div>
          )}
        </div>
      </section>

      {/* List Section */}
      <section className="rounded-3xl border border-slate-200/60 bg-white p-6 shadow-sm sm:p-8 lg:p-10">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 className="text-2xl font-black text-slate-900">Sản phẩm đang theo dõi</h2>
            <p className="mt-1.5 text-sm text-slate-500">
              Bạn có <strong className="text-slate-900">{products.length}</strong> sản phẩm đã thêm vào danh sách theo dõi.
            </p>
          </div>

          <button
            type="button"
            onClick={() => void loadTrackedProducts()}
            disabled={isLoadingProducts}
            className="inline-flex cursor-pointer items-center gap-2 rounded-xl bg-slate-50 px-4 py-2.5 text-sm font-bold text-slate-600 transition hover:bg-slate-100 hover:text-slate-900 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/20 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <svg className={`h-4 w-4 ${isLoadingProducts ? "animate-spin" : ""}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
            Làm mới
          </button>
        </div>

        {listError ? (
          <div className="mt-8 flex flex-col items-center justify-center rounded-2xl border border-red-100 bg-red-50 p-10 text-center">
            <svg className="mb-3 h-10 w-10 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
            <h3 className="text-lg font-bold text-red-900">Lỗi kết nối</h3>
            <p className="mt-1 text-sm text-red-700">{listError}</p>
          </div>
        ) : isLoadingProducts ? (
          <div className="mt-8 grid gap-4 lg:grid-cols-2">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="animate-pulse-slow flex items-center justify-between rounded-2xl border border-slate-100 bg-slate-50 p-5">
                <div className="space-y-3">
                  <div className="h-4 w-48 rounded bg-slate-200" />
                  <div className="h-3 w-32 rounded bg-slate-200" />
                  <div className="h-3 w-40 rounded bg-slate-200" />
                </div>
                <div className="h-10 w-24 rounded-xl bg-slate-200" />
              </div>
            ))}
          </div>
        ) : products.length === 0 ? (
          <div className="mt-8 flex flex-col items-center justify-center rounded-3xl border border-dashed border-slate-200 bg-slate-50 py-16 px-6 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-white text-slate-300 shadow-sm">
              <svg className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" /></svg>
            </div>
              <h3 className="mt-4 text-lg font-bold text-slate-900">Chưa có sản phẩm nào được theo dõi</h3>
            <p className="mt-2 max-w-sm text-sm leading-relaxed text-slate-500">
              Dán liên kết Shopee vào biểu mẫu phía trên, hoặc dùng hướng dẫn nhanh để tới cảnh báo đầu tiên.
            </p>
            <Link
              href="/onboarding"
              className="mt-6 inline-flex cursor-pointer rounded-xl bg-slate-900 px-5 py-3 text-sm font-extrabold text-white transition hover:bg-sky-700 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-500/25"
            >
              Bắt đầu: link → biểu đồ → cảnh báo
            </Link>
          </div>
        ) : (
          <div className="mt-8 grid gap-4 lg:grid-cols-2">
            {products.map((product) => (
              <div
                key={product.product_id}
                className="group flex flex-col justify-between gap-4 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-sky-200 hover:shadow-md sm:flex-row sm:items-center"
              >
                <div className="min-w-0 flex-1">
                  <div className="mb-2 flex items-center gap-2">
                    <span className="rounded bg-orange-100 px-2 py-0.5 text-[10px] font-black uppercase text-orange-700">
                      {formatPlatform(product.platform)}
                    </span>
                    <span className="text-xs font-bold text-slate-400">Sản phẩm #{product.product_id}</span>
                  </div>
                  <h3 className="truncate font-bold text-slate-900 transition group-hover:text-sky-700" title={product.platform_item_id}>
                    Mã gốc: {product.platform_item_id}
                  </h3>
                  <p className="mt-1 text-xs font-medium text-slate-500">
                    Thêm vào lúc {dateFormatter.format(new Date(product.tracked_at))}
                  </p>
                </div>

                <Link
                  href={`/products/${product.product_id}/chart`}
                  aria-label="Xem biểu đồ"
                  className="inline-flex shrink-0 cursor-pointer items-center justify-center gap-1.5 rounded-xl bg-slate-900 px-4 py-2.5 text-sm font-bold text-white shadow-sm transition hover:-translate-y-0.5 hover:bg-sky-700 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-500/25"
                >
                  Phân tích
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
                </Link>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
