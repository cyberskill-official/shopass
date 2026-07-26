import type { DailyPoint, Maturity, Verdict } from "@/lib/chart/types";

export type FakeSaleCheckResult =
  | {
      tracked: true;
      platform: string;
      product_id: number;
      maturity: Maturity;
      verdict: Verdict;
      current_price: number;
      median90: number;
      trailing_min: number;
      daily: DailyPoint[];
    }
  | {
      tracked: false;
      platform: string;
      message: string;
    };

export async function checkFakeSale(itemURL: string): Promise<FakeSaleCheckResult> {
  const res = await fetch("/api/tools/fake-sale-check", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ item_url: itemURL }),
  });
  const data = (await res.json().catch(() => ({}))) as FakeSaleCheckResult & { error?: string };
  if (!res.ok) {
    throw new Error(data.error || "Không kiểm tra được liên kết");
  }
  return data as FakeSaleCheckResult;
}
