"use client";

import { FormEvent, useState } from "react";
import { trackEvent } from "@/lib/analytics";
import { submitWaitlist } from "@/lib/billing/waitlist";

type Tier = "premium_basic" | "premium_plus" | "premium_pro";

export function WaitlistModal({
  tier,
  open,
  onClose,
}: {
  tier: Tier;
  open: boolean;
  onClose: () => void;
}) {
  const [email, setEmail] = useState("");
  const [zalo, setZalo] = useState("");
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!open) return null;

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await submitWaitlist({
        email,
        zalo: zalo || undefined,
        source: "pricing",
        tier_interest: tier,
      });
      trackEvent("waitlist-submit", { tier, source: "pricing" });
      setDone(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Lỗi");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/50 px-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="waitlist-title"
    >
      <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
        <div className="flex items-start justify-between gap-4">
          <h2 id="waitlist-title" className="text-lg font-black text-slate-900">
            Đăng ký chờ Premium
          </h2>
          <button type="button" className="text-sm font-bold text-slate-500 hover:text-slate-800" onClick={onClose}>
            Đóng
          </button>
        </div>
        {done ? (
          <p className="mt-4 text-sm leading-relaxed text-slate-600">
            Đã ghi nhận. Chúng tôi sẽ email khi thanh toán Premium mở chính thức.
          </p>
        ) : (
          <form className="mt-4 space-y-4" onSubmit={onSubmit}>
            <label className="block text-sm font-semibold text-slate-700">
              Email
              <input
                required
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm"
                autoComplete="email"
              />
            </label>
            <label className="block text-sm font-semibold text-slate-700">
              Zalo (tuỳ chọn)
              <input
                type="tel"
                value={zalo}
                onChange={(e) => setZalo(e.target.value)}
                className="mt-1 w-full rounded-xl border border-slate-200 px-3 py-2 text-sm"
                placeholder="09…"
              />
            </label>
            {error && <p className="text-sm text-red-600">{error}</p>}
            <button
              type="submit"
              disabled={busy}
              className="w-full rounded-xl bg-sky-700 px-4 py-3 text-sm font-extrabold text-white hover:bg-sky-800 disabled:opacity-60"
            >
              {busy ? "Đang gửi…" : "Giữ chỗ"}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
