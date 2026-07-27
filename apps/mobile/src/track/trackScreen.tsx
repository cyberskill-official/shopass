import { displayVerdict, type WishlistItem } from "./trackClient";
import { buildChartModel, type ChartModel } from "./priceChart";
import type { PriceHistoryPoint } from "./trackClient";

export interface TrackScreenModel {
  items: Array<{ product_id: number; verdictLabel: string }>;
  chart: ChartModel | null;
}

export function buildTrackScreen(
  wishlist: WishlistItem[],
  history?: PriceHistoryPoint[],
): TrackScreenModel {
  return {
    items: wishlist.map((w) => ({
      product_id: w.product_id,
      verdictLabel: displayVerdict(w.verdict),
    })),
    chart: history ? buildChartModel(history) : null,
  };
}
