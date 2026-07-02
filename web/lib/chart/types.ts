export interface DailyPoint {
  day: string;
  min_p: number;
  max_p: number;
  close_p: number;
}

export type Verdict = "SALE_AO" | "SALE_XIN" | "TAM_DUOC" | "UNKNOWN";
export type Maturity = "MATURE" | "WARMING" | "NEW";

export interface Annotations {
  median90: number;
  trailing_min: number;
  verdict: Verdict;
  accumulating: boolean;
  double_dates: string[];
}

export interface ChartResponse {
  product_id: number;
  range: string;
  maturity: Maturity;
  daily: DailyPoint[];
  annotations: Annotations;
}

export const RANGE_ALLOWLIST = ["7d", "30d", "90d", "180d", "1y"] as const;
export type Range = typeof RANGE_ALLOWLIST[number];
