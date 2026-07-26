"use client";

import Link from "next/link";
import { FormEvent, useMemo, useState } from "react";
import { trackEvent } from "@/lib/analytics";
import { submitWaitlist } from "@/lib/billing/waitlist";
import { daysUntil, nextDoubleDate } from "@/lib/tools/sale-calendar";

export function SaleCalendarTool() {
  const next = useMemo(() => nextDoubleDate(), []);
  const remaining = daysUntil(next.date);
  const [email, setEmail] = useState("");
  const [zalo, setZalo] = useState("");
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onSubscribe = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    trackEvent("tool-submit", { tool: "sale-calendar" });
    try {
      await submitWaitlist({
        email,
        zalo: zalo || undefined,
        source: "sale-calendar",
      });
      trackEvent("lead-captured", { tool: "sale-calendar", source: "sale-calendar" });
      setDone(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Lỗi");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto max-w-2xl px-6 py-12">
      <p className="text-sm font-bold text-sky-800">
        <Link href="/">Shopass</Link>
        {" · "}
        <Link href="/kiem-tra-sale-ao">Kiểm tra sale ảo</Link>
      </p>
      <h1 className="mt-4 text-3xl font-black tracking-tight text-slate-900">Lịch sale ngày đôi</h1>
      <p className="mt-3 text-slate-600">
        Đếm ngược tới đợt sale lớn tiếp theo (1.1–12.12). Đăng ký nhắc email/Zalo — không spam.
      </p>

      <div className="mt-10 border border-slate-200 bg-gradient-to-br from-sky-50 to-white p-8">
        <p className="text-sm font-bold uppercase tracking-wide text-slate-500">Sale tiếp theo</p>
        <p className="mt-2 text-4xl font-black text-sky-900">{next.label}</p>
        <p className="mt-2 text-lg font-semibold text-slate-700">
          Còn <span className="font-black text-slate-950">{remaining}</span> ngày
        </p>
        <p className="mt-1 text-sm text-slate-500">
          {next.date.toLocaleDateString("vi-VN", { timeZone: "UTC", weekday: "long", day: "numeric", month: "long", year: "numeric" })}
        </p>
      </div>

      <section className="mt-10">
        <h2 className="text-lg font-black">Nhắc trước ngày sale</h2>
        {done ? (
          <p className="mt-4 text-sm font-semibold text-emerald-700">Đã ghi nhận — sẽ nhắc trước ngày đôi.</p>
        ) : (
          <form className="mt-4 space-y-3" onSubmit={onSubscribe}>
            <label className="block text-sm font-semibold text-slate-700">
              Email
              <input
                required
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm"
              />
            </label>
            <label className="block text-sm font-semibold text-slate-700">
              Zalo (tuỳ chọn)
              <input
                type="tel"
                value={zalo}
                onChange={(e) => setZalo(e.target.value)}
                className="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm"
              />
            </label>
            {error && <p className="text-sm text-red-600">{error}</p>}
            <button
              type="submit"
              disabled={busy}
              className="rounded-xl bg-sky-700 px-5 py-3 text-sm font-extrabold text-white hover:bg-sky-800 disabled:opacity-60"
            >
              {busy ? "Đang gửi…" : "Nhắc tôi"}
            </button>
          </form>
        )}
      </section>
    </div>
  );
}
