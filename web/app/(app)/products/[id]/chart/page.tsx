"use client";

import React, { type FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { fetchChart, submitBrowserPrice } from "@/lib/chart/fetch-chart";
import { PriceChart } from "@/components/price-chart/price-chart";
import { VerdictBadge } from "@/components/price-chart/verdict-badge";
import { MaturityNotice } from "@/components/price-chart/maturity-notice";
import { RangeSelector } from "@/components/price-chart/range-selector";
import { BottomAlertCta } from "@/components/onboarding/bottom-alert-cta";
import type { ChartResponse, Range } from "@/lib/chart/types";

export default function ProductChartPage() {
  const params = useParams();
  const router = useRouter();
  const productId = Number(params.id);

  const [range, setRange] = useState<Range>("90d");
  const [data, setData] = useState<ChartResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [refreshKey, setRefreshKey] = useState(0);
  const [browserPrice, setBrowserPrice] = useState("");
  const [captureStatus, setCaptureStatus] = useState<string | null>(null);
  const [capturing, setCapturing] = useState(false);

  useEffect(() => {
    let ignore = false;

    async function loadData() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetchChart(productId, range);
        if (!ignore) {
          setData(res);
        }
      } catch (error) {
        if (!ignore) {
          setError(
            error instanceof Error && error.message
              ? error.message
              : "Đã xảy ra lỗi khi tải biểu đồ"
          );
        }
      } finally {
        if (!ignore) setLoading(false);
      }
    }

    if (productId && !isNaN(productId)) {
      loadData();
    }

    return () => { ignore = true; };
  }, [productId, range, refreshKey]);

  async function handleBrowserPrice(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setCaptureStatus(null);

    // Accept the common "6.490.000đ" copy/paste format, while storing only
    // an integer amount of Vietnamese đồng.
    const normalized = browserPrice.replace(/[^\d]/g, "");
    const price = Number(normalized);

    if (!Number.isSafeInteger(price) || price <= 0) {
      setCaptureStatus("Nhập giá hợp lệ, ví dụ 6.490.000.");
      return;
    }

    setCapturing(true);
    try {
      const result = await submitBrowserPrice(productId, price);
      setBrowserPrice("");
      setCaptureStatus(
        result.written
          ? "Đã ghi nhận giá bạn xác nhận. Biểu đồ đang được làm mới."
          : "Mức giá này đã được ghi nhận rồi. Biểu đồ đang được làm mới."
      );
      setRefreshKey((value) => value + 1);
    } catch (submitError) {
      setCaptureStatus(
        submitError instanceof Error && submitError.message
          ? submitError.message
          : "Không thể lưu giá lúc này. Vui lòng thử lại."
      );
    } finally {
      setCapturing(false);
    }
  }

  if (error) {
    return (
      <div className="mx-auto max-w-5xl">
        <button onClick={() => router.back()} className="mb-6 inline-flex items-center gap-1.5 text-sm font-bold text-slate-500 hover:text-slate-900">
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M10 19l-7-7m0 0l7-7m-7 7h18" /></svg>
          Quay lại danh sách
        </button>
        <div className="flex flex-col items-center justify-center rounded-3xl border border-red-100 bg-red-50 py-16 px-6 text-center shadow-sm">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-white text-red-500 shadow-sm">
            <svg className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
          </div>
          <h2 className="mt-5 text-xl font-black text-red-900">Không thể tải dữ liệu</h2>
          <p className="mt-2 text-sm text-red-700">{error}</p>
          <button onClick={() => window.location.reload()} className="mt-6 rounded-xl bg-white px-5 py-2.5 text-sm font-bold text-slate-700 shadow-sm transition hover:bg-slate-50 border border-slate-200">
            Thử lại
          </button>
        </div>
      </div>
    );
  }

  // Calculate percentage change
  let priceChange = null;
  let priceChangePercent = 0;
  if (data && data.daily.length > 0 && data.annotations.trailing_min > 0) {
    const currentPrice = data.daily[data.daily.length - 1].close_p;
    const minPrice = data.annotations.trailing_min;
    if (currentPrice > minPrice) {
      priceChangePercent = ((currentPrice - minPrice) / minPrice) * 100;
      priceChange = { type: "up", percent: priceChangePercent };
    } else if (currentPrice === minPrice) {
      priceChange = { type: "bottom", percent: 0 };
    }
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6">
      <div>
        <button onClick={() => router.back()} className="mb-4 inline-flex items-center gap-1.5 text-sm font-bold text-slate-500 transition hover:text-slate-900">
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M10 19l-7-7m0 0l7-7m-7 7h18" /></svg>
          Quay lại danh sách
        </button>

        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <p className="inline-flex items-center gap-1.5 rounded bg-blue-50 px-2 py-1 text-[10px] font-black uppercase tracking-widest text-blue-700">
              Shopee VN
            </p>
            <h1 className="mt-2 text-2xl font-black tracking-tight text-slate-900 sm:text-3xl lg:text-4xl">
              Sản phẩm #{productId}
            </h1>
            <p className="mt-1.5 text-sm text-slate-500">Mã gốc từ nền tảng: {productId}</p>
          </div>

          <div className="shrink-0">
            {data && <VerdictBadge verdict={data.annotations.verdict} maturity={data.maturity} />}
          </div>
        </div>
      </div>

      <div className="rounded-[2rem] border border-slate-200/60 bg-white p-5 shadow-lg shadow-slate-200/40 sm:p-7 md:p-8">
        <div className="flex flex-col gap-4 border-b border-slate-100 pb-6 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-lg font-black text-slate-900">Lịch sử giá & Phân tích</h2>
            <p className="mt-1 text-xs text-slate-500">Cập nhật lúc {new Date().toLocaleTimeString('vi-VN')}</p>
          </div>
          <RangeSelector currentRange={range} onChange={setRange} />
        </div>

        {data && data.daily.length > 0 ? (
          <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4 sm:gap-4 lg:gap-5">
            <div className="rounded-2xl border border-slate-100 bg-slate-50/50 p-4 sm:p-5">
              <p className="text-xs font-bold text-slate-500">Giá hiện tại</p>
              <p className="mt-2 text-xl font-black text-slate-950 sm:text-2xl">
                {new Intl.NumberFormat("vi-VN").format(data.daily[data.daily.length - 1].close_p)}₫
              </p>
            </div>

            <div className="rounded-2xl border border-blue-100 bg-blue-50/30 p-4 sm:p-5">
              <p className="text-xs font-bold text-blue-700">Trung vị {range === '90d' ? '90' : range === '30d' ? '30' : '180'} ngày</p>
              <p className="mt-2 text-xl font-black text-blue-900 sm:text-2xl">
                {data.annotations.median90 > 0 ? new Intl.NumberFormat("vi-VN").format(data.annotations.median90) + "₫" : "—"}
              </p>
            </div>

            <div className="rounded-2xl border border-emerald-100 bg-emerald-50/30 p-4 sm:p-5">
              <p className="text-xs font-bold text-emerald-700">Đáy thấp nhất</p>
              <p className="mt-2 text-xl font-black text-emerald-900 sm:text-2xl">
                {data.annotations.trailing_min > 0 ? new Intl.NumberFormat("vi-VN").format(data.annotations.trailing_min) + "₫" : "—"}
              </p>
            </div>

            <div className="rounded-2xl border border-slate-100 bg-slate-50/50 p-4 sm:p-5">
              <p className="text-xs font-bold text-slate-500">Biến động so với đáy</p>
              {priceChange ? (
                <div className="mt-2 flex items-center gap-1.5">
                  {priceChange.type === "up" ? (
                    <>
                      <span className="flex h-6 w-6 items-center justify-center rounded-full bg-rose-100 text-rose-600">
                        <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 10l7-7m0 0l7 7m-7-7v18" /></svg>
                      </span>
                      <span className="text-lg font-black text-rose-600">+{priceChange.percent.toFixed(1)}%</span>
                    </>
                  ) : (
                    <>
                      <span className="flex h-6 w-6 items-center justify-center rounded-full bg-emerald-100 text-emerald-600">
                        <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M19 14l-7 7m0 0l-7-7m7 7V3" /></svg>
                      </span>
                      <span className="text-lg font-black text-emerald-600">Đang ở đáy</span>
                    </>
                  )}
                </div>
              ) : (
                <p className="mt-2 text-lg font-black text-slate-400">—</p>
              )}
            </div>
          </div>
        ) : (
          loading && (
            <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
              {[1, 2, 3, 4].map(i => (
                <div key={i} className="animate-pulse-slow h-24 rounded-2xl bg-slate-100" />
              ))}
            </div>
          )
        )}

        <div className="mt-8">
          {data && <MaturityNotice maturity={data.maturity} />}
        </div>

        {productId > 0 && (
          <div className="mt-6">
            <BottomAlertCta productId={productId} />
          </div>
        )}

        <section className="mt-6 overflow-hidden rounded-2xl border border-blue-100 bg-gradient-to-br from-blue-50 via-white to-indigo-50/60">
          <div className="flex flex-col gap-5 p-5 sm:p-6 lg:flex-row lg:items-end lg:justify-between">
            <div className="max-w-xl">
              <p className="text-[10px] font-black uppercase tracking-[0.18em] text-blue-700">
                Browser-assisted beta
              </p>
              <h3 className="mt-2 text-base font-black text-slate-900 sm:text-lg">
                Ghi nhận giá bạn đang thấy trên Shopee
              </h3>
              <p className="mt-2 text-sm leading-6 text-slate-600">
                Mở sản phẩm ở tab Shopee, kiểm tra giá hiển thị rồi dán vào đây. Shopass chỉ lưu mức giá do bạn chủ động xác nhận — không truy cập tài khoản Shopee hoặc tự quét nền từ trình duyệt của bạn.
              </p>
              <Link href="/capture-guide" className="mt-3 inline-flex text-sm font-bold text-blue-700 transition hover:text-blue-900">
                Cài nút lấy giá một chạm →
              </Link>
            </div>

            <form onSubmit={handleBrowserPrice} className="w-full shrink-0 lg:max-w-sm" aria-label="Ghi nhận giá từ trình duyệt">
              <label htmlFor="browser-price" className="text-xs font-bold text-slate-700">
                Giá đang hiển thị (VNĐ)
              </label>
              <div className="mt-2 flex gap-2">
                <div className="relative min-w-0 flex-1">
                  <input
                    id="browser-price"
                    value={browserPrice}
                    onChange={(event) => setBrowserPrice(event.target.value)}
                    inputMode="numeric"
                    autoComplete="off"
                    placeholder="Ví dụ 6.490.000"
                    disabled={capturing}
                    className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 pr-8 text-sm font-bold text-slate-900 outline-none transition placeholder:font-medium placeholder:text-slate-400 focus:border-blue-400 focus:ring-4 focus:ring-blue-100 disabled:cursor-not-allowed disabled:bg-slate-50"
                  />
                  <span className="pointer-events-none absolute inset-y-0 right-3 flex items-center text-xs font-bold text-slate-400">₫</span>
                </div>
                <button
                  type="submit"
                  disabled={capturing || !browserPrice.trim()}
                  className="inline-flex h-11 shrink-0 items-center justify-center rounded-xl bg-slate-950 px-4 text-sm font-bold text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-300"
                >
                  {capturing ? "Đang lưu..." : "Lưu giá"}
                </button>
              </div>
              {captureStatus && (
                <p
                  className={`mt-2 text-xs font-medium ${
                    captureStatus.startsWith("Đã") || captureStatus.startsWith("Mức")
                      ? "text-emerald-700"
                      : "text-rose-600"
                  }`}
                  role="status"
                >
                  {captureStatus}
                </p>
              )}
            </form>
          </div>
        </section>

        <div className="relative mt-6 h-[400px] w-full rounded-2xl bg-white sm:h-[450px]">
          {loading && !data && (
            <div className="absolute inset-0 z-10 flex flex-col items-center justify-center rounded-2xl bg-white/90 backdrop-blur-sm">
              <div className="h-10 w-10 animate-spin rounded-full border-4 border-slate-100 border-t-blue-600" />
              <p className="mt-4 font-bold text-slate-600">Đang tải biểu đồ...</p>
            </div>
          )}

          {data && data.daily.length === 0 && !loading && (
            <div className="absolute inset-0 flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-200 bg-slate-50 text-center">
              <svg className="h-10 w-10 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" /></svg>
              <p className="mt-3 font-bold text-slate-700">Chưa có dữ liệu</p>
              <p className="mt-1 max-w-sm text-sm text-slate-500">Dùng mục “Ghi nhận giá bạn đang thấy trên Shopee” ở trên để thêm điểm giá đầu tiên do bạn xác nhận.</p>
            </div>
          )}

          {data && data.daily.length > 0 && <PriceChart data={data} />}
        </div>
      </div>
    </div>
  );
}
