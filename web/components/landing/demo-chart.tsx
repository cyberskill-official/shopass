"use client";

import dynamic from "next/dynamic";
import { LANDING_DEMO_CHART } from "@/lib/landing/demo-chart-data";

const PriceChart = dynamic(
  () => import("@/components/price-chart/price-chart").then((m) => m.PriceChart),
  {
    ssr: false,
    loading: () => (
      <div
        className="flex h-[280px] items-center justify-center rounded-2xl bg-slate-100 text-sm font-bold text-slate-500"
        aria-hidden
      >
        Đang tải biểu đồ…
      </div>
    ),
  },
);

export function LandingDemoChart() {
  return (
    <div className="landing-demo-chart w-full" aria-label="Biểu đồ giá minh họa">
      <PriceChart data={LANDING_DEMO_CHART} />
    </div>
  );
}
