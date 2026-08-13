/**
 * Build-time endpoint config for the Shopass MV3 extension.
 * Selected via SHOPASS_ENV at esbuild time (see build.mjs).
 *
 * Production API is same-origin /v1 on the web host. api.shopass.cyberskill.world
 * has no DNS and must not be used.
 */

export type ShopassEnv = "development" | "staging" | "production";

export type EndpointConfig = {
  env: ShopassEnv;
  webOrigin: string;
  apiOrigin: string;
  syncUrl: string;
  wssUrl: string;
  dsarUrl: string;
};

function resolveEnv(): ShopassEnv {
  // esbuild replaces process.env.SHOPASS_ENV with a string literal at build time.
  const raw = process.env.SHOPASS_ENV || "production";
  if (raw === "development" || raw === "staging" || raw === "production") return raw;
  return "production";
}

const ENV = resolveEnv();

const configs: Record<ShopassEnv, EndpointConfig> = {
  development: {
    env: "development",
    webOrigin: "http://127.0.0.1:3000",
    apiOrigin: "http://127.0.0.1:8080",
    syncUrl: "http://127.0.0.1:8080/v1/ext/sync",
    wssUrl: "ws://127.0.0.1:8080/v1/ext/ws",
    dsarUrl: "http://127.0.0.1:3000/dsar",
  },
  staging: {
    env: "staging",
    webOrigin: "https://shopass.cyberskill.world",
    apiOrigin: "https://shopass.cyberskill.world",
    syncUrl: "https://shopass.cyberskill.world/v1/ext/sync",
    wssUrl: "wss://shopass.cyberskill.world/v1/ext/ws",
    dsarUrl: "https://shopass.cyberskill.world/dsar",
  },
  production: {
    env: "production",
    webOrigin: "https://shopass.cyberskill.world",
    apiOrigin: "https://shopass.cyberskill.world",
    syncUrl: "https://shopass.cyberskill.world/v1/ext/sync",
    wssUrl: "wss://shopass.cyberskill.world/v1/ext/ws",
    dsarUrl: "https://shopass.cyberskill.world/dsar",
  },
};

export const config: EndpointConfig = configs[ENV] ?? configs.production;

export const SYNC_URL = config.syncUrl;
export const WSS_URL = config.wssUrl;
export const DSAR_URL = config.dsarUrl;
export const WEB_ORIGIN = config.webOrigin;
export const API_ORIGIN = config.apiOrigin;
