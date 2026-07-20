"use client";

import Link from "next/link";
import { Suspense, type FormEvent, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { submitBrowserPrice } from "@/lib/chart/fetch-chart";
import { canonicalShopeeProductURL, normalizeCapturedPrice } from "@/lib/browser-capture";
import { trackShopeeProduct } from "@/lib/track/api";

function messageFor(error: unknown): string {
  if (error instanceof Error && error.message) {
    if (error.message === "invalid item_url") {
      return "Shopass chưa nhận diện được liên kết sản phẩm Shopee này.";
    }
    return error.message;
  }
  return "Không thể lưu giá lúc này. Vui lòng thử lại.";
}

function BrowserCaptureContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [itemURL, setItemURL] = useState("");
  const [priceText, setPriceText] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const url = canonicalShopeeProductURL(searchParams.get("url"));
    const price = normalizeCapturedPrice(searchParams.get("price"));
    setItemURL(url ?? "");
    setPriceText(price ? new Intl.NumberFormat("vi-VN").format(price) : "");
  }, [searchParams]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);

    const normalizedURL = canonicalShopeeProductURL(itemURL);
    const price = normalizeCapturedPrice(priceText);
    if (!normalizedURL) {
      setError("Dán liên kết trực tiếp của một sản phẩm Shopee Việt Nam.");
      return;
    }
    if (!price) {
      setError("Kiểm tra và nhập giá hợp lệ bằng VNĐ trước khi lưu.");
      return;
    }

    setSubmitting(true);
    try {
      const tracked = await trackShopeeProduct(normalizedURL);
      await submitBrowserPrice(tracked.product_id, price);
      router.replace(`/products/${tracked.product_id}/chart?captured=1`);
    } catch (captureError) {
      setError(messageFor(captureError));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="mx-auto max-w-2xl py-4 sm:py-8">
      <Link href="/dashboard" className="inline-flex items-center gap-1.5 text-sm font-bold text-slate-500 transition hover:text-slate-900">
        ← Về bảng điều khiển
      </Link>

      <section className="mt-5 overflow-hidden rounded-[2rem] border border-blue-100 bg-gradient-to-br from-blue-50 via-white to-violet-50 p-6 shadow-xl shadow-blue-100/30 sm:p-9">
        <p className="text-[10px] font-black uppercase tracking-[0.2em] text-blue-700">Ghi giá từ trình duyệt</p>
        <h1 className="mt-3 text-2xl font-black tracking-tight text-slate-950 sm:text-3xl">Xác nhận giá trước khi lưu</h1>
        <p className="mt-3 max-w-xl text-sm leading-6 text-slate-600">
          Shopass chỉ nhận liên kết sản phẩm và mức giá đang hiển thị. Hệ thống không đọc tài khoản, cookie hay lịch sử duyệt Shopee của bạn.
        </p>

        <form className="mt-7 space-y-5" onSubmit={handleSubmit}>
          <label className="block">
            <span className="text-xs font-black uppercase tracking-wide text-slate-700">Liên kết sản phẩm Shopee</span>
            <input
              type="url"
              value={itemURL}
              onChange={(event) => setItemURL(event.target.value)}
              placeholder="https://shopee.vn/...-i.123.456"
              autoComplete="off"
              disabled={submitting}
              className="mt-2 h-12 w-full rounded-xl border border-slate-200 bg-white px-4 text-sm font-medium text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-blue-400 focus:ring-4 focus:ring-blue-100 disabled:bg-slate-50"
            />
          </label>

          <label className="block">
            <span className="text-xs font-black uppercase tracking-wide text-slate-700">Giá bạn đang thấy (VNĐ)</span>
            <input
              inputMode="numeric"
              value={priceText}
              onChange={(event) => setPriceText(event.target.value)}
              placeholder="Ví dụ 6.490.000"
              autoComplete="off"
              disabled={submitting}
              className="mt-2 h-12 w-full rounded-xl border border-slate-200 bg-white px-4 text-base font-black text-slate-900 outline-none transition placeholder:text-sm placeholder:font-medium placeholder:text-slate-400 focus:border-blue-400 focus:ring-4 focus:ring-blue-100 disabled:bg-slate-50"
            />
          </label>

          <div className="rounded-xl border border-amber-100 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-900">
            Hãy chọn đúng phân loại trên Shopee trước khi xác nhận. Nếu Shopee hiển thị một khoảng giá, Shopass sẽ lưu giá bạn kiểm tra ở bước này.
          </div>

          {error && <p role="alert" className="rounded-xl bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700">{error}</p>}

          <button
            type="submit"
            disabled={submitting}
            className="inline-flex h-12 w-full items-center justify-center rounded-xl bg-slate-950 px-5 text-sm font-black text-white shadow-lg shadow-slate-300 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-300"
          >
            {submitting ? "Đang tạo lịch sử giá…" : "Xác nhận và lưu vào biểu đồ"}
          </button>
        </form>
      </section>
    </main>
  );
}

export default function BrowserCapturePage() {
  return (
    <Suspense
      fallback={(
        <main className="mx-auto max-w-2xl py-8 text-sm font-semibold text-slate-500">
          Đang mở bước xác nhận giá…
        </main>
      )}
    >
      <BrowserCaptureContent />
    </Suspense>
  );
}
