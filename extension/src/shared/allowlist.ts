/**
 * Single source of truth for extension outbound hosts (R8 + R31).
 * host_permissions, build-time config, and DNR allow rules must stay aligned.
 */

/** Marketplace pages the content scripts may run on / read. */
export const MARKETPLACE_HOST_PERMISSIONS = [
  "https://shopee.vn/*",
  "https://*.tiktok.com/*",
  "https://www.lazada.vn/*",
] as const;

/** Shopass API / gateway hosts the extension may call (sync, ws upgrade path). */
export const API_HOST_PERMISSIONS = [
  "https://shopass.cyberskill.world/*",
  "http://127.0.0.1:8080/*",
] as const;

export const ALL_HOST_PERMISSIONS = [
  ...MARKETPLACE_HOST_PERMISSIONS,
  ...API_HOST_PERMISSIONS,
] as const;

/** Hostnames allowed for extension-initiated API traffic (no path/scheme). */
export const ALLOWED_API_HOSTNAMES = [
  "shopass.cyberskill.world",
  "127.0.0.1",
] as const;

export function isAllowedApiUrl(urlString: string): boolean {
  let url: URL;
  try {
    url = new URL(urlString);
  } catch {
    return false;
  }
  if (url.protocol !== "http:" && url.protocol !== "https:" && url.protocol !== "ws:" && url.protocol !== "wss:") {
    return false;
  }
  const host = url.hostname;
  if (!(ALLOWED_API_HOSTNAMES as readonly string[]).includes(host)) return false;
  if (host === "127.0.0.1" && url.port && url.port !== "8080") return false;
  return true;
}
