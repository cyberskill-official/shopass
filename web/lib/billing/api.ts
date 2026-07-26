import { apiFetch } from "@/lib/api";

export type PlanTier = "premium_basic" | "premium_plus" | "premium_pro";
export type Gateway = "momo" | "zalopay" | "vnpay" | "vietqr";

export interface CheckoutResult {
  order_ref: string;
  gateway: string;
  amount: number;
  pay_url?: string;
  qr_payload?: string;
}

export async function startCheckout(
  planTier: PlanTier,
  gateway: Gateway,
): Promise<CheckoutResult> {
  const res = await apiFetch("/v1/billing/checkout", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ plan_tier: planTier, gateway }),
  });
  if (!res.ok) {
    throw new Error("Không thể tạo phiên thanh toán");
  }
  return (await res.json()) as CheckoutResult;
}
