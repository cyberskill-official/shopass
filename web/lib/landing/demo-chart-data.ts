import type { ChartResponse } from "@/lib/chart/types";

/** Synthetic 90-day series for the public landing demo (not live product data). */
export const LANDING_DEMO_CHART: ChartResponse = {
  product_id: 0,
  range: "90d",
  maturity: "MATURE",
  daily: Array.from({ length: 90 }, (_, i) => {
    const day = new Date(Date.UTC(2026, 3, 1 + i));
    const wave = Math.sin(i / 9) * 180_000;
    const close = Math.round(7_200_000 - i * 8_000 + wave);
    return {
      day: day.toISOString().slice(0, 10),
      min_p: close - 40_000,
      max_p: close + 60_000,
      close_p: close,
    };
  }),
  annotations: {
    median90: 6_850_000,
    trailing_min: 6_190_000,
    verdict: "SALE_XIN",
    accumulating: false,
    double_dates: ["2026-05-05", "2026-06-06"],
  },
};
