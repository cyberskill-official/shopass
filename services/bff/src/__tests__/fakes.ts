import { GraphQLError } from "graphql";
import type { RestClient } from "../rest";
import type { ChartData, User, Wishlist } from "../types";

export function sampleChart(): ChartData {
  return {
    maturity: "mature",
    daily: [{ day: "2026-07-01", minP: 100000, maxP: 120000, closeP: 110000 }],
    annotations: {
      median90: 115000,
      trailingMin: 100000,
      verdict: "wait",
      accumulating: false,
      doubleDates: ["2026-06-01"],
    },
  };
}

export interface FakeOpts {
  wishlists?: Wishlist[];
  failChart?: boolean; // simulate a track/deal 404 already mapped to NOT_FOUND
  failWishlists?: boolean;
}

// FakeRest stands in for the REST layer. It counts getChart calls so tests can
// prove DataLoader batching/caching (one call per distinct key, not per field).
export class FakeRest implements RestClient {
  chartCalls = 0;
  chartKeys: string[] = [];

  constructor(private readonly opts: FakeOpts = {}) {}

  async getMe(): Promise<User> {
    return { id: "u1", displayName: "Chi" };
  }

  async listWishlists(): Promise<Wishlist[]> {
    if (this.opts.failWishlists) {
      throw new GraphQLError("NOT_FOUND", { extensions: { code: "NOT_FOUND" } });
    }
    return this.opts.wishlists ?? [];
  }

  async getChart(productId: string, range: string): Promise<ChartData> {
    this.chartCalls++;
    this.chartKeys.push(`${productId}:${range}`);
    if (this.opts.failChart) {
      throw new GraphQLError("NOT_FOUND", { extensions: { code: "NOT_FOUND" } });
    }
    return sampleChart();
  }
}
