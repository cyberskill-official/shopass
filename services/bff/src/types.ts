// GraphQL-facing domain types (camelCase), mirroring the schema. The RestClient
// maps each backend REST payload into these shapes so resolvers stay thin.

export interface User {
  id: string;
  displayName: string | null;
}

export interface WishlistItem {
  productId: string;
  targetPrice: number | null; // VND, int
}

export interface Wishlist {
  id: string;
  name: string;
  items: WishlistItem[];
}

export interface DailyPoint {
  day: string; // ISO date
  minP: number;
  maxP: number;
  closeP: number;
}

export interface Annotations {
  median90: number;
  trailingMin: number;
  verdict: string;
  accumulating: boolean;
  doubleDates: string[];
}

// Mirror of the FR-DEAL-003 chart feed. The BFF forwards this shape unchanged;
// it does not compute verdict/median itself (FR-WEB-005 §1 #9).
export interface ChartData {
  maturity: string;
  daily: DailyPoint[];
  annotations: Annotations;
}
