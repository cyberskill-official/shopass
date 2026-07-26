"use client";

import { PriceChart } from "@/components/price-chart/price-chart";
import { LANDING_DEMO_CHART } from "@/lib/landing/demo-chart-data";

export function LandingDemoChart() {
  return (
    <div className="landing-demo-chart w-full" aria-label="Biểu đồ giá minh họa">
      <PriceChart data={LANDING_DEMO_CHART} />
    </div>
  );
}
