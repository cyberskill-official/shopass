import type { NextRequest } from "next/server";

const THIRTY_DAYS_SECONDS = 30 * 24 * 60 * 60;

export function refreshCookieCandidates(env = process.env.NODE_ENV): readonly string[] {
  if (env === "production") {
    return ["__Host-sandeal_refresh"];
  }
  // The old name is local-development compatibility only. Never accept a
  // Domain-scoped legacy cookie in production, where it would defeat __Host-
  // cookie isolation.
  return ["sandeal_refresh", "refresh_token"];
}

export const refreshCookieName = refreshCookieCandidates()[0];

export const refreshCookieOptions = {
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "strict" as const,
  path: "/",
  maxAge: THIRTY_DAYS_SECONDS,
};

export function requestHasAllowedOrigin(request: NextRequest): boolean {
  const origin = request.headers.get("origin");
  if (!origin) {
    // Browser fetch requests have Origin. Permit omitted Origin only for local
    // tooling; a production browser request without it must be rejected.
    return process.env.NODE_ENV !== "production";
  }
  const configuredOrigin = process.env.APP_ORIGIN;
  if (configuredOrigin) {
    return origin === configuredOrigin;
  }
  return origin === request.nextUrl.origin;
}

export async function gatewayFetch(
  path: string,
  init: RequestInit,
  request?: NextRequest,
): Promise<Response> {
  const base = process.env.GATEWAY_INTERNAL_BASE_URL || "http://gateway:8080";
  const headers = new Headers(init.headers);
  // Caddy overwrites X-Real-IP with its peer's remote host before traffic
  // reaches Next. Do not permit an arbitrary RequestInit value to cross the
  // private hop; only relay the value from that trusted inbound request.
  headers.delete("X-Real-IP");
  const clientIP = request?.headers.get("x-real-ip")?.trim();
  if (clientIP) headers.set("X-Real-IP", clientIP);

  return fetch(new URL(path, base), {
    ...init,
    headers,
    cache: "no-store",
  });
}

export function readRefreshCookie(request: NextRequest, env = process.env.NODE_ENV): string | undefined {
  for (const name of refreshCookieCandidates(env)) {
    const value = request.cookies.get(name)?.value;
    if (value) return value;
  }
  return undefined;
}
