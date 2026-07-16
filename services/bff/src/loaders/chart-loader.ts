import DataLoader from "dataloader";
import type { ChartData } from "../types";
import type { RestClient } from "../rest";

export interface ChartKey {
  productId: string;
  range: string;
}

// makeChartLoader batches and caches chart fetches within a single request
// (TASK-WEB-005 §1 #3 / DEC-WEB-23). A query that loads the chart for every
// wishlist item collapses into one batched tick instead of N sequential REST
// calls, and duplicate (productId, range) keys are de-duplicated by the cache.
export function makeChartLoader(rest: RestClient): DataLoader<ChartKey, ChartData, string> {
  return new DataLoader<ChartKey, ChartData, string>(
    (keys) => Promise.all(keys.map((k) => rest.getChart(k.productId, k.range))),
    { cacheKeyFn: (k) => `${k.productId}:${k.range}` },
  );
}
