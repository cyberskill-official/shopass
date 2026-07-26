"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { VerdictBadge } from "@/components/price-chart/verdict-badge";
import { trackEvent } from "@/lib/analytics";
import { submitWaitlist } from "@/lib/billing/waitlist";
import { checkFakeSale, type FakeSaleCheckResult } from "@/lib/tools/fake-sale";

function formatVND(n: number): string {
  return new Intl.NumberFormat("vi-VN").format(n) + "đ";
}

export function FakeSaleChecker() {
  const [url, setURL] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<FakeSaleCheckResult | null>(null);
  const [email, setEmail] = useState("");
  const [leadDone, setLeadDone] = useState(false);
  const [leadBusy, setLeadBusy] = useState(false);

  const onCheck = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    setResult(null);
    setLeadDone(false);
    trackEvent("tool-submit", { tool: "fake-sale" });
    try {
      const got = await checkFakeSale(url.trim());
      setResult(got);
      if (got.tracked) {
        trackEvent("verdict-shown", { verdict: got.verdict, platform: got.platform });
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Lỗi");
    } finally {
      setBusy(false);
    }
  };

  const onLead = async (e: FormEvent) => {
    e.preventDefault();
    setLeadBusy(true);
    setError(null);
    try {
      await submitWaitlist({ email, source: "tool" });
      trackEvent("lead-captured", { tool: "fake-sale", source: "tool" });
      setLeadDone(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Lỗi");
    } finally {
      setLeadBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-2xl px-6 py-12">
      <p className="text-sm font-bold text-sky-800">
        <Link href="/">Shopass</Link>
        {" · "}
        <Link href="/lich-sale">Lịch sale</Link>
      </p>
      <h1 className="mt-4 text-3xl font-black tracking-tight text-slate-900">
        Kiểm tra sale ảo
      </h1>
      <p className="mt-3 text-slate-600">
        Dán link sản phẩm Shopee / Lazada / TikTok. Nếu Shopass đã theo dõi, bạn thấy verdict so với
        trung vị 90 ngày — không cần đăng nhập.
      </p>

      <form className="mt-8 space-y-4" onSubmit={onCheck}>
        <label className="block text-sm font-semibold text-slate-700">
          Link sản phẩm
          <input
            required
            type="url"
            value={url}
            onChange={(e) => setURL(e.target.value)}
            placeholder="https://shopee.vn/…-i.shop.item"
            className="mt-1 w-full rounded-xl border border-slate-200 px-3 py-3 text-sm"
          />
        </label>
        <button
          type="submit"
          disabled={busy}
          className="rounded-xl bg-sky-700 px-5 py-3 text-sm font-extrabold text-white hover:bg-sky-800 disabled:opacity-60"
        >
          {busy ? "Đang kiểm tra…" : "Kiểm tra"}
        </button>
      </form>

      {error && <p className="mt-4 text-sm text-red-600">{error}</p>}

      {result?.tracked && (
        <section className="mt-10 border-t border-slate-200 pt-8">
          <div className="flex flex-wrap items-center gap-3">
            <VerdictBadge verdict={result.verdict} maturity={result.maturity} />
            {result.maturity === "NEW" && (
              <span className="text-sm font-semibold text-slate-500">Chưa đủ lịch sử (&lt;14 ngày)</span>
            )}
          </div>
          <dl className="mt-6 grid gap-4 sm:grid-cols-3">
            <div>
              <dt className="text-xs font-bold uppercase tracking-wide text-slate-500">Giá hiện tại</dt>
              <dd className="mt-1 text-lg font-black text-slate-900">{formatVND(result.current_price)}</dd>
            </div>
            <div>
              <dt className="text-xs font-bold uppercase tracking-wide text-slate-500">Median 90 ngày</dt>
              <dd className="mt-1 text-lg font-black text-slate-900">{formatVND(result.median90)}</dd>
            </div>
            <div>
              <dt className="text-xs font-bold uppercase tracking-wide text-slate-500">Đáy gần đây</dt>
              <dd className="mt-1 text-lg font-black text-slate-900">{formatVND(result.trailing_min)}</dd>
            </div>
          </dl>
          <p className="mt-6 text-sm text-slate-600">
            Muốn cảnh báo khi giá chạm đáy?{" "}
            <Link href="/login?signup=1" className="font-bold text-sky-800 underline">
              Tạo tài khoản miễn phí
            </Link>
          </p>
        </section>
      )}

      {result && !result.tracked && (
        <section className="mt-10 border-t border-slate-200 pt-8">
          <h2 className="text-lg font-black text-slate-900">Chưa có trong Shopass</h2>
          <p className="mt-2 text-sm text-slate-600">
            Để lại email — chúng tôi báo khi verdict sẵn sàng (sau khi lịch sử giá đủ dài).
          </p>
          {leadDone ? (
            <p className="mt-4 text-sm font-semibold text-emerald-700">Đã ghi nhận email.</p>
          ) : (
            <form className="mt-4 flex flex-col gap-3 sm:flex-row" onSubmit={onLead}>
              <input
                required
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@email.com"
                className="flex-1 rounded-xl border border-slate-200 px-3 py-3 text-sm"
              />
              <button
                type="submit"
                disabled={leadBusy}
                className="rounded-xl bg-slate-950 px-5 py-3 text-sm font-extrabold text-white disabled:opacity-60"
              >
                {leadBusy ? "Đang gửi…" : "Báo tôi"}
              </button>
            </form>
          )}
        </section>
      )}
    </div>
  );
}
