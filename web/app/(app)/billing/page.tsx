"use client";

import { FormEvent, useState } from "react";
import {
  startCheckout,
  type Gateway,
  type PlanTier,
} from "@/lib/billing/api";

const PLANS: { tier: PlanTier; label: string; price: string }[] = [
  { tier: "premium_basic", label: "Premium Basic", price: "29.000đ / tháng" },
  { tier: "premium_plus", label: "Premium Plus", price: "49.000đ / tháng" },
  { tier: "premium_pro", label: "Premium Pro", price: "79.000đ / tháng" },
];

const GATEWAYS: { id: Gateway; label: string }[] = [
  { id: "momo", label: "MoMo" },
  { id: "zalopay", label: "ZaloPay" },
  { id: "vnpay", label: "VNPay" },
  { id: "vietqr", label: "VietQR" },
];

export default function BillingPage() {
  const [planTier, setPlanTier] = useState<PlanTier>("premium_basic");
  const [gateway, setGateway] = useState<Gateway>("momo");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [qr, setQr] = useState<string | null>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    setQr(null);
    try {
      const result = await startCheckout(planTier, gateway);
      if (result.pay_url) {
        window.location.href = result.pay_url;
        return;
      }
      if (result.qr_payload) {
        setQr(result.qr_payload);
        return;
      }
      setError("Cổng thanh toán không trả về liên kết thanh toán.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Thanh toán thất bại");
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className="mx-auto max-w-xl space-y-8">
      <header>
        <h1 className="text-3xl font-black tracking-tight text-slate-900">Nâng cấp Shopass</h1>
        <p className="mt-2 text-base text-slate-500">
          Chọn gói và cổng thanh toán. Sau khi thanh toán sandbox thành công, quyền Premium được kích hoạt qua IPN.
        </p>
      </header>

      <form onSubmit={onSubmit} className="space-y-6 rounded-2xl border border-slate-200 bg-white p-6">
        <fieldset className="space-y-2">
          <legend className="text-sm font-semibold text-slate-700">Gói</legend>
          {PLANS.map((p) => (
            <label key={p.tier} className="flex cursor-pointer items-center gap-3 rounded-lg border border-slate-100 px-3 py-2">
              <input
                type="radio"
                name="plan"
                checked={planTier === p.tier}
                onChange={() => setPlanTier(p.tier)}
              />
              <span className="flex-1 font-medium text-slate-800">{p.label}</span>
              <span className="text-sm text-slate-500">{p.price}</span>
            </label>
          ))}
        </fieldset>

        <fieldset className="space-y-2">
          <legend className="text-sm font-semibold text-slate-700">Cổng thanh toán</legend>
          <div className="flex flex-wrap gap-2">
            {GATEWAYS.map((g) => (
              <button
                key={g.id}
                type="button"
                onClick={() => setGateway(g.id)}
                className={`rounded-xl px-4 py-2 text-sm font-bold ${
                  gateway === g.id
                    ? "bg-slate-900 text-white"
                    : "bg-slate-100 text-slate-700"
                }`}
              >
                {g.label}
              </button>
            ))}
          </div>
        </fieldset>

        <button
          type="submit"
          disabled={busy}
          className="w-full rounded-xl bg-slate-900 px-4 py-3 text-sm font-extrabold text-white disabled:opacity-50"
        >
          {busy ? "Đang tạo phiên…" : "Thanh toán"}
        </button>

        {error ? (
          <p className="text-sm text-rose-600">{error}</p>
        ) : null}
        {qr ? (
          <p className="break-all rounded-lg bg-slate-50 p-3 font-mono text-xs text-slate-700">
            {qr}
          </p>
        ) : null}
      </form>
    </section>
  );
}