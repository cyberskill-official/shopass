import DataLoader from "dataloader";
import { randomUUID } from "node:crypto";
import type { RestClient } from "./rest";
import { HttpRestClient } from "./rest";
import { makeChartLoader, type ChartKey } from "./loaders/chart-loader";
import type { ChartData } from "./types";

export interface Loaders {
  chart: DataLoader<ChartKey, ChartData, string>;
}

// GqlContext carries the identity the gateway already verified (DEC-WEB-21): the
// BFF does NOT parse or verify the token itself. Per-request DataLoaders live here
// so batching/caching never leaks across requests or users.
export interface GqlContext {
  userId: string | null; // sub, forwarded by the gateway (already verified)
  requestId: string; // forwarded by the gateway (TASK-INFRA-001)
  rest: RestClient;
  loaders: Loaders;
}

// header reads one forwarded header value regardless of Node's string|string[] typing.
function header(v: string | string[] | undefined): string | undefined {
  if (Array.isArray(v)) return v[0];
  return v;
}

export interface ContextConfig {
  trackBase: string;
  dealBase: string;
}

// buildContext turns a request's forwarded headers into a context. It trusts
// x-user-id (the gateway set it after verifying the JWT) and never re-verifies.
export function buildContext(
  headers: Record<string, string | string[] | undefined>,
  cfg: ContextConfig,
): GqlContext {
  const userId = header(headers["x-user-id"]) ?? null;
  const requestId = header(headers["x-request-id"]) ?? randomUUID();
  const rest = new HttpRestClient(userId, requestId, cfg.trackBase, cfg.dealBase);
  return { userId, requestId, rest, loaders: { chart: makeChartLoader(rest) } };
}

// contextFromRest builds a context around an already-constructed RestClient.
// Used by tests (inject a fake) and anywhere the REST client is pre-built.
export function contextFromRest(
  userId: string | null,
  rest: RestClient,
  requestId = randomUUID(),
): GqlContext {
  return { userId, requestId, rest, loaders: { chart: makeChartLoader(rest) } };
}
