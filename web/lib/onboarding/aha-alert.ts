import { apiFetch } from "@/lib/api";
import type { AlertRule } from "@/lib/alerts/api";
import type { RuleType } from "@/lib/alerts/validate";

export type AhaAlertResult = {
  rule: AlertRule;
  rule_type: RuleType;
  /** True when bottom_predicted required Premium; real_sale was created instead. */
  premium_deferred: boolean;
};

/**
 * Prefer bottom_predicted (R45 “chạm đáy”); free tier gets real_sale so activation
 * still completes without a billing detour.
 */
export async function createAhaAlert(productId: number): Promise<AhaAlertResult> {
  const bottom = await apiFetch("/v1/alerts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      product_id: productId,
      rule_type: "bottom_predicted",
      threshold: null,
      channel: ["push"],
    }),
  });

  if (bottom.ok) {
    return {
      rule: (await bottom.json()) as AlertRule,
      rule_type: "bottom_predicted",
      premium_deferred: false,
    };
  }

  if (bottom.status === 402) {
    const sale = await apiFetch("/v1/alerts", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        product_id: productId,
        rule_type: "real_sale",
        threshold: null,
        channel: ["push"],
      }),
    });
    if (!sale.ok) {
      throw new Error("Không thể tạo cảnh báo");
    }
    return {
      rule: (await sale.json()) as AlertRule,
      rule_type: "real_sale",
      premium_deferred: true,
    };
  }

  throw new Error("Không thể tạo cảnh báo");
}
