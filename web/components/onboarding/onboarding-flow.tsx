"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { MaturityNotice } from "@/components/price-chart/maturity-notice";
import { PriceChart } from "@/components/price-chart/price-chart";
import { VerdictBadge } from "@/components/price-chart/verdict-badge";
import { trackEvent } from "@/lib/analytics";
import { fetchChart } from "@/lib/chart/fetch-chart";
import type { ChartResponse } from "@/lib/chart/types";
import { createAhaAlert } from "@/lib/onboarding/aha-alert";
import { trackShopeeProduct } from "@/lib/track/api";

type Step = "paste" | "chart" | "done";

export function OnboardingFlow() {
  const [step, setStep] = useState<Step>("paste");
  const [url, setURL] = useState("");
  const [productId, setProductId] = useState<number | null>(null);
  const [chart, setChart] = useState<ChartResponse | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [premiumNote, setPremiumNote] = useState(false);

  const onTrack = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const tracked = await trackShopeeProduct(url.trim());
      trackEvent("first-track", { product_id: tracked.product_id, already: tracked.already_tracked });
      setProductId(tracked.product_id);
      try {
        setChart(await fetchChart(tracked.product_id, "90d"));
      } catch {
        setChart(null);
      }
      setStep("chart");
    } catch (err) {
      setError(
        err instanceof Error && err.message === "invalid item_url"
          ? "Liên kết không hợp lệ. Dán link sản phẩm Shopee VN (…-i.shop.item)."
          : err instanceof Error
            ? err.message
            : "Không theo dõi được sản phẩm",
      );
    } finally {
      setBusy(false);
    }
  };

  const onAlert = async () => {
    if (productId == null) return;
    setBusy(true);
    setError(null);
    try {
      const result = await createAhaAlert(productId);
      trackEvent("first-alert", {
        product_id: productId,
        rule_type: result.rule_type,
        premium_deferred: result.premium_deferred,
      });
      setPremiumNote(result.premium_deferred);
      setStep("done");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Không tạo được cảnh báo");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-3xl space-y-8">
      <header>
        <p className="text-xs font-bold uppercase tracking-widest text-sky-800">Bắt đầu trong 2 phút</p>
        <h1 className="mt-2 text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">
          Dán link → xem lịch sử → bật cảnh báo
        </h1>
        <p className="mt-3 max-w-xl text-slate-600">
          Đây là khoảnh khắc “aha” của Shopass: thấy giá thật trên biểu đồ, rồi một nút để được báo khi
          gần đáy hoặc khi sale thật.
        </p>
      </header>

      {step === "paste" && (
        <form onSubmit={onTrack} className="space-y-4 rounded-3xl border border-slate-200 bg-white p-6 sm:p-8">
          <label className="block text-sm font-semibold text-slate-700">
            Link sản phẩm Shopee / TikTok / Lazada
            <input
              required
              type="url"
              value={url}
              onChange={(e) => setURL(e.target.value)}
              placeholder="https://shopee.vn/…-i.shop.item"
              className="mt-2 w-full rounded-xl border border-slate-200 px-3 py-3 text-sm"
              aria-label="Link sản phẩm để theo dõi"
            />
          </label>
          <p className="text-xs text-slate-500">
            Closed beta hiện resolve tốt nhất với Shopee VN. Các sàn khác sẽ mở dần.
          </p>
          <button
            type="submit"
            disabled={busy}
            className="rounded-xl bg-sky-700 px-5 py-3 text-sm font-extrabold text-white hover:bg-sky-800 disabled:opacity-60"
          >
            {busy ? "Đang thêm…" : "Theo dõi & xem biểu đồ"}
          </button>
        </form>
      )}

      {step === "chart" && productId != null && (
        <section className="space-y-6 rounded-3xl border border-slate-200 bg-white p-6 sm:p-8">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h2 className="text-xl font-black text-slate-900">Sản phẩm #{productId}</h2>
              <p className="mt-1 text-sm text-slate-500">Đã thêm vào danh sách theo dõi của bạn.</p>
            </div>
            {chart && <VerdictBadge verdict={chart.annotations.verdict} maturity={chart.maturity} />}
          </div>

          {chart ? (
            <>
              <MaturityNotice maturity={chart.maturity} />
              {chart.daily.length > 0 ? (
                <div className="h-[280px]">
                  <PriceChart data={chart} />
                </div>
              ) : (
                <p className="rounded-xl border border-dashed border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
                  Chưa có đủ điểm giá — bạn vẫn có thể bật cảnh báo; lịch sử sẽ dày thêm khi scrape/browser
                  capture chạy (R27).
                </p>
              )}
            </>
          ) : (
            <p className="text-sm text-slate-600">
              Biểu đồ tạm chưa tải được — bạn vẫn có thể bật cảnh báo ngay.
            </p>
          )}

          <button
            type="button"
            onClick={() => void onAlert()}
            disabled={busy}
            className="w-full rounded-xl bg-slate-950 px-5 py-4 text-sm font-extrabold text-white hover:bg-slate-800 disabled:opacity-60 sm:w-auto"
          >
            {busy ? "Đang bật…" : "Báo tôi khi chạm đáy"}
          </button>
          <p className="text-xs text-slate-500">
            Free: nếu dự đoán đáy cần Premium, Shopass tự bật cảnh báo “sale thật” để bạn vẫn có tín hiệu.
          </p>
        </section>
      )}

      {step === "done" && productId != null && (
        <section className="rounded-3xl border border-emerald-200 bg-emerald-50/60 p-6 sm:p-8">
          <h2 className="text-xl font-black text-emerald-950">Xong — cảnh báo đã bật</h2>
          <p className="mt-2 text-sm text-emerald-900">
            {premiumNote
              ? "Tài khoản Free chưa mở dự đoán đáy — đã bật cảnh báo sale thật. Nâng Premium khi muốn p_bottom."
              : "Cảnh báo chạm đáy đã được tạo. Shopass sẽ báo khi tín hiệu đủ mạnh."}
          </p>
          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              href={`/products/${productId}/chart`}
              className="rounded-xl bg-slate-950 px-4 py-3 text-sm font-extrabold text-white"
            >
              Xem biểu đồ đầy đủ
            </Link>
            <Link
              href="/dashboard"
              className="rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm font-bold text-slate-800"
            >
              Về bảng điều khiển
            </Link>
            {premiumNote && (
              <Link href="/bang-gia" className="rounded-xl px-4 py-3 text-sm font-bold text-sky-800 underline">
                Xem bảng giá Premium
              </Link>
            )}
          </div>
        </section>
      )}

      {error && <p className="text-sm text-red-600">{error}</p>}
    </div>
  );
}
