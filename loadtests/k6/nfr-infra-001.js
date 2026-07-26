/**
 * R20 — NFR-INFRA-001 mixed load gate.
 *
 * Route classes (gateway-published beta surface):
 *   price_chart  → GET /v1/products/{id}/chart   p95 < 500ms
 *   read_cached  → GET /v1/tracked-products       p95 < 300ms
 *
 * Note: /v1/compare and /v1/wishlists are gateway-rejected (404) in closed beta;
 * tracked-products stands in for the cache-class read.
 *
 * Env:
 *   K6_BASE_URL       default https://api.shopass.cyberskill.world
 *   K6_ACCESS_TOKEN   Bearer JWT (required for real runs)
 *   K6_PRODUCT_ID     tracked product with chart data (default 1)
 *   K6_VUS            concurrent VUs (default 10)
 *   K6_DURATION       steady duration (default 3m)
 *
 * Rate-limit note: gateway default is ~100 req/min/user. For 50 RPS, provision
 * a token pool (K6_ACCESS_TOKENS comma-separated) or a staging bypass.
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate } from "k6/metrics";

const BASE = (__ENV.K6_BASE_URL || "https://api.shopass.cyberskill.world").replace(/\/$/, "");
const PRODUCT_ID = __ENV.K6_PRODUCT_ID || "1";
const tokens = (__ENV.K6_ACCESS_TOKENS || __ENV.K6_ACCESS_TOKEN || "")
  .split(",")
  .map((t) => t.trim())
  .filter(Boolean);

const chartTrend = new Trend("nfr_chart_duration", true);
const cacheTrend = new Trend("nfr_cache_duration", true);
const errorRate = new Rate("nfr_errors");

export const options = {
  scenarios: {
    mixed: {
      executor: "ramping-arrival-rate",
      startRate: 5,
      timeUnit: "1s",
      preAllocatedVUs: Number(__ENV.K6_VUS || 20),
      maxVUs: Number(__ENV.K6_MAX_VUS || 80),
      stages: [
        { target: 20, duration: "30s" },
        { target: 50, duration: "1m" },
        { target: 50, duration: __ENV.K6_DURATION || "3m" },
        { target: 0, duration: "30s" },
      ],
    },
  },
  thresholds: {
    // Soft-fail friendly locally when secrets missing: CI sets K6_STRICT=1.
    nfr_chart_duration: ["p(95)<500"],
    nfr_cache_duration: ["p(95)<300"],
    nfr_errors: ["rate<0.05"],
    http_req_failed: ["rate<0.05"],
  },
};

function authHeaders() {
  if (tokens.length === 0) {
    return { headers: { Accept: "application/json" } };
  }
  const tok = tokens[Math.floor(Math.random() * tokens.length)];
  return {
    headers: {
      Accept: "application/json",
      Authorization: `Bearer ${tok}`,
    },
  };
}

export function setup() {
  if (tokens.length === 0) {
    console.warn(
      "K6_ACCESS_TOKEN(S) unset — requests will 401; thresholds will fail. " +
        "Set secrets for weekly CI / local staging runs.",
    );
  }
  return { base: BASE, productId: PRODUCT_ID };
}

export default function (data) {
  const opts = authHeaders();
  // ~60% chart / ~40% cache
  if (Math.random() < 0.6) {
    const res = http.get(
      `${data.base}/v1/products/${data.productId}/chart?range=90d`,
      { ...opts, tags: { name: "price_chart" } },
    );
    chartTrend.add(res.timings.duration);
    const ok = check(res, {
      "chart status 200": (r) => r.status === 200,
    });
    errorRate.add(!ok);
  } else {
    const res = http.get(`${data.base}/v1/tracked-products`, {
      ...opts,
      tags: { name: "read_cached" },
    });
    cacheTrend.add(res.timings.duration);
    const ok = check(res, {
      "cache status 200": (r) => r.status === 200,
    });
    errorRate.add(!ok);
  }
  sleep(0.05);
}

export function handleSummary(data) {
  return {
    "loadtests/k6/summary.json": JSON.stringify(data, null, 2),
    stdout: textSummary(data),
  };
}

function textSummary(data) {
  const chart = data.metrics.nfr_chart_duration;
  const cache = data.metrics.nfr_cache_duration;
  const lines = [
    "=== NFR-INFRA-001 k6 summary ===",
    `base=${BASE} product=${PRODUCT_ID} tokens=${tokens.length}`,
    chart
      ? `chart p95=${(chart.values["p(95)"] || 0).toFixed(1)}ms (target <500)`
      : "chart: no samples",
    cache
      ? `cache p95=${(cache.values["p(95)"] || 0).toFixed(1)}ms (target <300)`
      : "cache: no samples",
  ];
  return lines.join("\n") + "\n";
}
