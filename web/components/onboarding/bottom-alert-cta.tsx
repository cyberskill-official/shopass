"use client";

import Link from "next/link";
import { useState } from "react";
import { trackEvent } from "@/lib/analytics";
import { createAhaAlert } from "@/lib/onboarding/aha-alert";

export function BottomAlertCta({ productId }: { productId: number }) {
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [premiumNote, setPremiumNote] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onClick = async () => {
    setBusy(true);
    setError(null);
    try {
      const result = await createAhaAlert(productId);
      trackEvent("first-alert", {
        product_id: productId,
        rule_type: result.rule_type,
        premium_deferred: result.premium_deferred,
        surface: "chart",
      });
      setPremiumNote(result.premium_deferred);
      setDone(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Lỗi");
    } finally {
      setBusy(false);
    }
  };

  if (done) {
    return (
      <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-5 py-4 text-sm text-emerald-900">
        <p className="font-bold">Đã bật cảnh báo.</p>
        {premiumNote ? (
          <p className="mt-1">
            Free: dùng cảnh báo sale thật.{" "}
            <Link href="/bang-gia" className="font-bold underline">
              Premium cho dự đoán đáy
            </Link>
          </p>
        ) : (
          <p className="mt-1">Bạn sẽ nhận push khi gần đáy.</p>
        )}
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50 px-5 py-4">
      <p className="text-sm font-bold text-slate-900">Bước tiếp theo</p>
      <p className="mt-1 text-sm text-slate-600">Một chạm để được báo khi giá gần đáy (hoặc sale thật trên Free).</p>
      <button
        type="button"
        onClick={() => void onClick()}
        disabled={busy}
        className="mt-3 rounded-xl bg-slate-950 px-4 py-2.5 text-sm font-extrabold text-white disabled:opacity-60"
      >
        {busy ? "Đang bật…" : "Báo tôi khi chạm đáy"}
      </button>
      {error && <p className="mt-2 text-sm text-red-600">{error}</p>}
    </div>
  );
}
