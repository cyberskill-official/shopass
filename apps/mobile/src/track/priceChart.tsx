import type { PriceHistoryPoint } from "./trackClient";

/** Pure render model — no price math (DEC-MOBILE-10). */
export interface ChartModel {
  labels: string[];
  values: number[];
  empty: boolean;
}

export function buildChartModel(points: PriceHistoryPoint[]): ChartModel {
  if (!points.length) {
    return { labels: [], values: [], empty: true };
  }
  return {
    labels: points.map((p) => p.day),
    values: points.map((p) => p.close_p),
    empty: false,
  };
}
