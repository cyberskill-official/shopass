import { GraphQLError } from "graphql";
import type { ChartData, User, Wishlist } from "./types";

// RestClient is the only way resolvers reach data. Resolvers MUST NOT touch the
// DB directly (TASK-WEB-005 §1 #2 / DEC-WEB-22): ownership checks (anti-IDOR) live
// in track-svc (TASK-TRACK-002/003), and the chart feed in deal-svc (TASK-DEAL-003).
// One interface so tests inject a fake and the server injects an HTTP client.
export interface RestClient {
  getMe(): Promise<User>;
  listWishlists(): Promise<Wishlist[]>;
  getChart(productId: string, range: string): Promise<ChartData>;
}

// neutralError maps an upstream failure to a GraphQL error that never leaks
// another user's resource (TASK-WEB-005 §1 #10). A 403/404 from a REST service
// (e.g. an IDOR attempt caught by track-svc) becomes an opaque NOT_FOUND; a
// network/5xx failure becomes a generic server error, not swallowed silently.
function neutralError(status: number | null, what: string): GraphQLError {
  if (status === 403 || status === 404) {
    return new GraphQLError("NOT_FOUND", { extensions: { code: "NOT_FOUND" } });
  }
  return new GraphQLError(`upstream ${what} failed`, {
    extensions: { code: "UPSTREAM_ERROR" },
  });
}

// HttpRestClient forwards the gateway-provided identity and request id downstream
// (TASK-WEB-005 §1 #1, #8): x-user-id lets each REST service enforce ownership,
// x-request-id keeps one trace across services (TASK-INFRA-004).
export class HttpRestClient implements RestClient {
  constructor(
    private readonly userId: string | null,
    private readonly requestId: string,
    private readonly trackBase: string,
    private readonly dealBase: string,
    private readonly fetchImpl: typeof fetch = fetch,
  ) {}

  private headers(): Record<string, string> {
    const h: Record<string, string> = { "x-request-id": this.requestId };
    if (this.userId) h["x-user-id"] = this.userId;
    return h;
  }

  private async getJSON(url: string, what: string): Promise<unknown> {
    let resp: Response;
    try {
      resp = await this.fetchImpl(url, { headers: this.headers() });
    } catch {
      throw neutralError(null, what); // network error -> generic server error
    }
    if (!resp.ok) throw neutralError(resp.status, what);
    try {
      return await resp.json();
    } catch {
      throw neutralError(null, what);
    }
  }

  async getMe(): Promise<User> {
    const raw = (await this.getJSON(`${this.trackBase}/v1/me`, "me")) as {
      id?: string;
      display_name?: string | null;
    };
    return { id: String(raw.id ?? this.userId ?? ""), displayName: raw.display_name ?? null };
  }

  async listWishlists(): Promise<Wishlist[]> {
    const raw = (await this.getJSON(`${this.trackBase}/v1/wishlists`, "wishlists")) as Array<{
      id: string | number;
      name: string;
      items?: Array<{ product_id: string | number; target_price?: number | null }>;
    }>;
    return (raw ?? []).map((w) => ({
      id: String(w.id),
      name: w.name,
      items: (w.items ?? []).map((it) => ({
        productId: String(it.product_id),
        targetPrice: it.target_price ?? null,
      })),
    }));
  }

  async getChart(productId: string, range: string): Promise<ChartData> {
    const raw = (await this.getJSON(
      `${this.dealBase}/v1/products/${encodeURIComponent(productId)}/chart?range=${encodeURIComponent(range)}`,
      "chart",
    )) as ChartData;
    return raw;
  }
}
